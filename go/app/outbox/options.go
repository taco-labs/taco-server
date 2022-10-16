package outbox

import (
	"time"

	"github.com/taco-labs/taco/go/app"
	"github.com/taco-labs/taco/go/repository"
	"github.com/taco-labs/taco/go/service"
)

type outboxOpts func(*outboxApp)

func WithTransactor(transactor app.Transactor) outboxOpts {
	return func(oa *outboxApp) {
		oa.Transactor = transactor
	}
}

func WithEventRepository(repo repository.EventRepository) outboxOpts {
	return func(oa *outboxApp) {
		oa.repository.event = repo
	}
}

func WithEventPublishService(svc service.EventPublishService) outboxOpts {
	return func(oa *outboxApp) {
		oa.service.eventPub = svc
	}
}

func WithTargetEventUirs(uris []string) outboxOpts {
	return func(oa *outboxApp) {
		oa.conf.targetEventUris = uris
	}
}

func WithPollInterval(pollInterval time.Duration) outboxOpts {
	return func(oa *outboxApp) {
		oa.conf.pollInterval = pollInterval
	}
}

func WithMaxMessages(maxMessages int) outboxOpts {
	return func(oa *outboxApp) {
		oa.conf.maxMessages = maxMessages
	}
}
