package outbox

import (
	"time"

	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/repository"
	"github.com/taco-labs/taco/go/service"
)

type outboxAppOption func(*outboxApp)

func WithTransactor(transactor app.Transactor) outboxAppOption {
	return func(oa *outboxApp) {
		oa.Transactor = transactor
	}
}

func WithEventRepository(repo repository.EventRepository) outboxAppOption {
	return func(oa *outboxApp) {
		oa.repository.event = repo
	}
}

func WithEventPublishService(svc service.EventPublishService) outboxAppOption {
	return func(oa *outboxApp) {
		oa.service.eventPub = svc
	}
}

func WithTargetEventUriPrefix(eventUriPrefix string) outboxAppOption {
	return func(oa *outboxApp) {
		oa.conf.targetEventUriPrefix = eventUriPrefix
	}
}

func WithPollInterval(pollInterval time.Duration) outboxAppOption {
	return func(oa *outboxApp) {
		oa.conf.pollInterval = pollInterval
	}
}

func WithMaxMessages(maxMessages int) outboxAppOption {
	return func(oa *outboxApp) {
		oa.conf.maxMessages = maxMessages
	}
}
