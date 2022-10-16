package outbox

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/repository"
	"github.com/taco-labs/taco/go/service"
	"github.com/uptrace/bun"
)

type outboxApp struct {
	app.Transactor
	cancel     context.CancelFunc
	waitCh     chan struct{}
	repository struct {
		event repository.EventRepository
	}
	service struct {
		eventPub service.EventPublishService
	}
	conf struct {
		targetEventUris []string
		pollInterval    time.Duration // TODO(taekkyeom) exponential backoff for error case
		maxMessages     int           // TODO (taekyeom) adaptive message batch sizes
	}
}

func (o *outboxApp) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	o.cancel = cancel
	go o.loop(ctx)
	return nil
}

func (o outboxApp) Shuwdown() error {
	o.cancel()
	<-o.waitCh
	return nil
}

func (o outboxApp) loop(ctx context.Context) error {
	timer := time.NewTimer(0 * time.Second)
	for {
		select {
		case <-timer.C:
			err := o.sendBestAffort(ctx)
			if err != nil {
				// TODO (taekyeom) logging
			}
			timer.Reset(o.conf.pollInterval)
		}
	}
}

func (o outboxApp) sendBestAffort(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			o.waitCh <- struct{}{}
			return nil
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
}

func (o outboxApp) sendMessageBatch(ctx context.Context) (bool, error) {
	nonEmpty := true
	err := o.Run(ctx, func(ctx context.Context, i bun.IDB) error {
		events, err := o.repository.event.BatchGet(ctx, i, o.conf.targetEventUris, o.conf.maxMessages)
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

	if len(o.conf.targetEventUris) == 0 {
		return errors.New("outbox app need target event uris")
	}

	if o.conf.pollInterval.Microseconds() < 5 {
		return errors.New("outbox app required at least 5 microseconds poll interval")
	}

	if o.conf.maxMessages > 100 || o.conf.maxMessages < 0 {
		return errors.New("outbox's max messages should between 1 ~ 100")
	}

	return nil
}

func NewOutboxApp(opts ...outboxOpts) (outboxApp, error) {
	app := outboxApp{
		waitCh: make(chan struct{}),
	}

	for _, opt := range opts {
		opt(&app)
	}

	return app, app.validateApp()
}
