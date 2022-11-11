package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
	"gocloud.dev/pubsub"
)

type EventPublishService interface {
	SendMessage(context.Context, entity.Event) error
}

type EventSubscriptionService interface {
	GetMessage(context.Context) (entity.Event, error)
}

type sqsPubService struct {
	pub *pubsub.Topic
}

func (s sqsPubService) SendMessage(ctx context.Context, event entity.Event) error {
	return s.pub.Send(ctx, ToMessage(event))
}

func NewSqsPubService(pub *pubsub.Topic) *sqsPubService {
	return &sqsPubService{
		pub: pub,
	}
}

type sqsSubService struct {
	sub *pubsub.Subscription
}

func (s sqsSubService) GetMessage(ctx context.Context) (entity.Event, error) {
	message, err := s.sub.Receive(ctx)
	if err != nil {
		return entity.Event{}, fmt.Errorf("%w: error from recieve message from sqs: %v", value.ErrExternal, err)
	}

	return ToEvent(message), nil
}

func NewSqsSubService(sub *pubsub.Subscription) *sqsSubService {
	return &sqsSubService{
		sub: sub,
	}
}

func ToMessage(event entity.Event) *pubsub.Message {
	message := pubsub.Message{
		Metadata: map[string]string{
			entity.MetaDataKey_EventUri:   event.EventUri,
			entity.MetadataKey_RetryCount: fmt.Sprint(event.RetryCount),
		},
		Body: event.Payload,
		BeforeSend: func(asFunc func(interface{}) bool) error {
			req := &types.SendMessageBatchRequestEntry{}
			if asFunc(&req) {
				req.DelaySeconds = event.DelaySeconds
			}
			return nil
		},
	}
	return &message
}

func ToEvent(msg *pubsub.Message) entity.Event {
	event := entity.Event{
		MessageId: msg.LoggableID,
		Payload:   msg.Body,
	}
	eventUri, ok := msg.Metadata[entity.MetaDataKey_EventUri]
	if ok {
		event.EventUri = eventUri
	}
	retryCount, ok := msg.Metadata[entity.MetadataKey_RetryCount]
	if ok {
		retryCount, _ := strconv.Atoi(retryCount)
		event.RetryCount = retryCount
	}
	rawMsg := types.Message{}
	if msg.As(&rawMsg) {
		sentTimestampStr := rawMsg.Attributes[string(types.MessageSystemAttributeNameSentTimestamp)]
		sentTimestamp, _ := strconv.ParseInt(sentTimestampStr, 10, 64)
		sentTime := time.UnixMilli(sentTimestamp)
		event.CreateTime = sentTime
	}

	event.SetAck(func() error {
		msg.Ack()
		return nil
	})

	return event
}
