package outbox

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/repository"
	"github.com/taco-labs/taco/go/service"
	"github.com/taco-labs/taco/go/utils"
	"github.com/uptrace/bun"
	"go.uber.org/zap"
)

type outboxApp struct {
	app.Transactor
	waitCh     chan struct{}
	shutdownCh chan struct{}
	repository struct {
		event repository.EventRepository
	}
	service struct {
		eventPub service.EventPublishService
	}
	conf struct {
		targetEventUriPrefix string
		pollInterval         time.Duration // TODO(taekkyeom) exponential backoff for error case
		maxMessages          int           // TODO (taekyeom) adaptive message batch sizes
	}
}

func (o *outboxApp) Start(ctx context.Context) error {
	go o.loop(ctx)
	fmt.Printf("outbox app started...\n")
	return nil
}

func (o outboxApp) Shuwdown() error {
	o.shutdownCh <- struct{}{}
	<-o.waitCh
	return nil
}

func (o outboxApp) loop(ctx context.Context) error {
	timer := time.NewTimer(0 * time.Second)
	logger := utils.GetLogger(ctx)
	for range timer.C {
		err := o.sendBestAffort(ctx)
		if err != nil {
			logger.Error("app.outbox.app: error while sending message", zap.Error(err))
		}
		timer.Reset(o.conf.pollInterval)
	}
	return nil
}

func (o outboxApp) sendBestAffort(ctx context.Context) error {
OUTBOX:
	for {
		select {
		case <-ctx.Done():
			<-o.shutdownCh
			break OUTBOX
		case <-o.shutdownCh:
			break OUTBOX
		default:
			nonEmpty, err := o.sendMessageBatch(ctx)
			if err != nil {
				return err
			}
			if !nonEmpty {
				return nil
			}
		}
	}
	o.waitCh <- struct{}{}
	return nil
}

func (o outboxApp) sendMessageBatch(ctx context.Context) (bool, error) {
	nonEmpty := true
	err := o.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		events, err := o.repository.event.BatchGet(ctx, i, o.conf.targetEventUriPrefix, o.conf.maxMessages)
		if err != nil {
			return fmt.Errorf("app.outbox.loop: error while batch get events: %w", err)
		}

		if len(events) == 0 {
			nonEmpty = false
			return nil
		}

		for _, event := range events {
			if err := o.service.eventPub.SendMessage(ctx, event); err != nil {
				return fmt.Errorf("app.outbox.sendMessageBatch: error while send event: %w", err)
			}
		}

		if err = o.repository.event.BatchCommit(ctx, i, events); err != nil {
			return fmt.Errorf("app.outbox.loop: error while batch commit events: %w", err)
		}

		return nil
	})

	if err != nil {
		return false, err
	}

	return nonEmpty, nil
}

func (o outboxApp) validateApp() error {
	if o.Transactor == nil {
		return errors.New("outbox app need transactor")
	}

	if o.repository.event == nil {
		return errors.New("outbox app need event repository")
	}

	if o.service.eventPub == nil {
		return errors.New("outbox app need event publisher app")
	}

	if o.conf.pollInterval.Microseconds() < 5 {
		return errors.New("outbox app required at least 5 microseconds poll interval")
	}

	if o.conf.maxMessages > 100 || o.conf.maxMessages < 0 {
		return errors.New("outbox's max messages should between 1 ~ 100")
	}

	return nil
}

func NewOutboxApp(opts ...outboxAppOption) (outboxApp, error) {
	app := outboxApp{
		waitCh:     make(chan struct{}),
		shutdownCh: make(chan struct{}),
	}

	for _, opt := range opts {
		opt(&app)
	}

	return app, app.validateApp()
}
