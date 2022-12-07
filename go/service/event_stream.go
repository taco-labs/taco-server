package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/utils"
	"go.uber.org/zap"
)

type SubscriberStream interface {
	Accept(ctx context.Context, event entity.Event) bool
	Process(ctx context.Context, event entity.Event) error
	OnFailure(ctx context.Context, event entity.Event, lastErr error) error
}

type EventSubscriptionStreamService struct {
	eventSub          EventSubscriptionService
	streams           []SubscriberStream
	workerPoolService WorkerPoolService
	shutdownCh        chan struct{}
}

func (s *EventSubscriptionStreamService) Add(subscriberStream SubscriberStream) {
	s.streams = append(s.streams, subscriberStream)
}

func (s EventSubscriptionStreamService) Run(ctx context.Context) {
	go s.loop(ctx)
}

func (s EventSubscriptionStreamService) loop(ctx context.Context) {
	logger := utils.GetLogger(ctx)
	for {
		select {
		case <-ctx.Done():
			fmt.Println("shutting down EventSubscriberStream...")
			s.shutdownCh <- struct{}{}
			return
		default:
			err := s.consume(ctx, logger)
			if err != nil && errors.Is(err, context.Canceled) {
				logger.Error("[EventSubscriptionStreamService] error while consume event", zap.Error(err))
			}
		}
	}
}

func (s EventSubscriptionStreamService) consume(ctx context.Context, logger *zap.Logger) error {
	event, err := s.eventSub.GetMessage(ctx)
	if err != nil {
		return err
	}

	for idx := range s.streams {
		if s.streams[idx].Accept(ctx, event) {
			return s.workerPoolService.Submit(func() {
				err := s.streams[idx].Process(ctx, event)
				if err != nil {
					logger.Error("[EventSubscriptionStreamService.consume] error while handle consumed message", zap.Error(err))
					if errors.Is(err, context.Canceled) {
						return
					}
					// If error occurred, resend event with increased retry event count
					if event.Attempt < 4 {
						event.Nack()
						return
					}
					// If error limit reached.. handle failure
					if failreHandlerErr := s.streams[idx].OnFailure(ctx, event, err); failreHandlerErr != nil {
						logger.Error("[EventSubscriberStream.consume] error while handle failure message", zap.Error(failreHandlerErr))
					}
				}
				event.Ack()
			})
		}
	}

	// TODO (structued error)
	return fmt.Errorf("service.EventSubscriptionStreamService: Invalid event %v", event)
}

func (s EventSubscriptionStreamService) Shutdown(ctx context.Context) {
	<-s.shutdownCh
}

func NewEventSubscriptionStreamService(eventSub EventSubscriptionService, workerPool WorkerPoolService) *EventSubscriptionStreamService {
	return &EventSubscriptionStreamService{
		eventSub:          eventSub,
		streams:           make([]SubscriberStream, 0),
		workerPoolService: workerPool,
		shutdownCh:        make(chan struct{}),
	}
}
