package taxicall

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/domain/value/enum"
	"github.com/taco-labs/taco/go/repository"
	"github.com/taco-labs/taco/go/utils"
	"github.com/uptrace/bun"
)

type actor struct {
	app.Transactor // TODO(taekyeom) Do not use transactor?
	termChan       chan string
	callRequestId  string
	cancel         context.CancelFunc
	err            error
	repository     struct {
		user            repository.UserRepository
		driver          repository.DriverRepository
		taxiCallRequest repository.TaxiCallRepository
	}
}

func (a *actor) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	a.cancel = cancel
	a.loop(ctx)
	return nil
}

func (a *actor) Stop(ctx context.Context) error {
	a.cancel()
	return a.err
}

func (a actor) loop(ctx context.Context) {
	timer := time.NewTimer(time.Duration(0))
	go func() {
		for {
			select {
			case <-ctx.Done():
				if !timer.Stop() {
					<-timer.C
				}
			case t := <-timer.C:
				terminate, nextSchedule, err := a.tick(ctx, t)
				if err != nil {
					// TODO (taekyeom) restart?
					a.err = err
					fmt.Printf("Error: %+v\n", err)
					return
				}
				if terminate {
					a.termChan <- a.callRequestId
					fmt.Printf("Terminate actor [%s]\n", a.callRequestId)
					return
				}
				timer.Reset(time.Until(nextSchedule))
			}
		}
	}()
}

func (a actor) callMatched(ctx context.Context, ticketId string, t time.Time) error {
	return a.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		return nil
	})
}

func (a actor) tick(ctx context.Context, t time.Time) (bool, time.Time, error) {
	// get taxi call ticket & taxi call request
	var nextSchedule time.Time
	var terminate bool
	err := a.Run(context.Background(), func(ctx context.Context, i bun.IDB) error {
		taxiCallRequest, err := a.repository.taxiCallRequest.GetById(ctx, i, a.callRequestId)
		if err != nil {
			return fmt.Errorf("service.actor.tick [%s]: error while get call request: %w", a.callRequestId, err)
		}

		if taxiCallRequest.CurrentState.Complete() {
			if err := a.repository.taxiCallRequest.DeleteTicketByRequestId(ctx, i, taxiCallRequest.Id); err != nil {
				return fmt.Errorf("service.actor.ticket: [%s] failed to delete failed tickets: %w", a.callRequestId, err)
			}
			terminate = true
			return nil
		}

		if taxiCallRequest.CurrentState.InDriving() {
			// Send location push message to user
			nextSchedule = t.Add(time.Second * 10)
			return nil
		}

		ticket, err := a.repository.taxiCallRequest.GetLatestTicketByRequestId(ctx, i, taxiCallRequest.Id)
		if err != nil && !errors.Is(err, value.ErrNotFound) {
			return fmt.Errorf("service.actor.tick [%s]: error while get call request ticket: %w", a.callRequestId, err)
		}
		if err != nil && errors.Is(err, value.ErrNotFound) {
			ticket = entity.TaxiCallTicket{
				Id:                utils.MustNewUUID(),
				TaxiCallRequestId: taxiCallRequest.Id,
				Attempt:           0,
				AdditionalPrice:   0,
				CreateTime:        t,
			}
		}

		expectedSchedule := ticket.UpdateTime.Add(time.Second * 10)
		if t.Before(expectedSchedule) {
			nextSchedule = expectedSchedule
			return nil
		}

		validTicketOperation := true
		if !ticket.IncreaseAttempt(t) {
			validTicketOperation = ticket.IncreasePrice(taxiCallRequest.RequestMaxAdditionalPrice, t)
		}

		if validTicketOperation {
			// TODO (taekyeom) terminate & push message
			if err := taxiCallRequest.UpdateState(t, enum.TaxiCallState_FAILED); err != nil {
				return fmt.Errorf("service.actor.ticket [%s]: failed to update state: %w", a.callRequestId, err)
			}
			if err := a.repository.taxiCallRequest.Update(ctx, i, taxiCallRequest); err != nil {
				return fmt.Errorf("service.actor.ticket: [%s] failed to update call request to failed state: %w", a.callRequestId, err)
			}
			if err := a.repository.taxiCallRequest.DeleteTicketByRequestId(ctx, i, taxiCallRequest.Id); err != nil {
				return fmt.Errorf("service.actor.ticket: [%s] failed to delete failed tickets: %w", a.callRequestId, err)
			}
			terminate = true
			return nil
		}

		// Update new ticket
		if err := a.repository.taxiCallRequest.UpsertTicket(ctx, i, ticket); err != nil {
			return fmt.Errorf("service.actor.ticket: [%s] error while update new ticket: %w", a.callRequestId, err)
		}

		// Get drivers

		// Send pushes

		nextSchedule = ticket.UpdateTime.Add(10 * time.Second)

		return nil
	})

	return terminate, nextSchedule, err
}

// TODO (taekyeom) supervisor impl
type TaxiCallActorService struct {
	app.Transactor
	actorTermChan chan string
	actorMap      map[string]actor
	lock          sync.RWMutex
	repository    struct {
		user            repository.UserRepository
		driver          repository.DriverRepository
		taxiCallRequest repository.TaxiCallRepository
	}
}

func (t *TaxiCallActorService) actorTerminationHandler() {
	for requestId := range t.actorTermChan {
		t.lock.Lock()
		delete(t.actorMap, requestId)
		t.lock.Unlock()
	}
}

func (t *TaxiCallActorService) Add(requestId string) error {
	ctx := context.Background()
	newActor := actor{
		Transactor:    t.Transactor,
		callRequestId: requestId,
		repository:    t.repository,
		termChan:      t.actorTermChan,
	}
	t.lock.Lock()
	defer t.lock.Unlock()
	if err := newActor.Start(ctx); err != nil {
		return err
	}
	t.actorMap[requestId] = newActor
	return nil
}

func (t *TaxiCallActorService) Remove(requestId string) error {
	ctx := context.Background()
	t.lock.Lock()
	defer t.lock.Unlock()
	actorRef, ok := t.actorMap[requestId]
	if !ok {
		return fmt.Errorf("actor with id %s not found: %w", requestId, value.ErrNotFound)
	}
	if err := actorRef.Stop(ctx); err != nil {
		return fmt.Errorf("actor with id %s has internal error: %w: %v", requestId, value.ErrInternal, err)
	}
	delete(t.actorMap, requestId)
	return nil
}

func (t *TaxiCallActorService) Init(ctx context.Context) error {
	// TODO (taekyeom) recover from repository
	go t.actorTerminationHandler()

	var activeRequestIds []string

	err := t.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		ids, err := t.repository.taxiCallRequest.GetActiveRequestIds(ctx, i)
		if err != nil {
			return fmt.Errorf("service.actorsystem: error while get active requests: %w", err)
		}
		activeRequestIds = ids
		return nil
	})

	if err != nil {
		return err
	}

	for _, activeRequestId := range activeRequestIds {
		t.Add(activeRequestId)
	}

	return nil
}

func (t *TaxiCallActorService) validate() error {
	if t.Transactor == nil {
		return errors.New("actor service needs transactor")
	}

	if t.repository.driver == nil {
		return errors.New("actor service needs driver repository")
	}

	if t.repository.user == nil {
		return errors.New("actor system needs user repository")
	}

	if t.repository.taxiCallRequest == nil {
		return errors.New("actor system needs taxi call request repository")
	}

	return nil
}

func NewTaxiCallActorService(opts ...actorOption) (*TaxiCallActorService, error) {
	svc := &TaxiCallActorService{
		actorMap:      make(map[string]actor),
		actorTermChan: make(chan string),
	}

	for _, opt := range opts {
		opt(svc)
	}

	return svc, svc.validate()
}
