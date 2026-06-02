package nats

import (
	"context"
	"review-service/config"
	eventbroker "review-service/internal/presentation/event_broker"

	"github.com/nats-io/nats.go"
	"github.com/ofm-microservices/ofm-common/pkg/logging"
)

// PullConsumerConfigValidator validates pull-consumer runtime configuration
// before the broker touches JetStream state.
type PullConsumerConfigValidator interface {
	Validate(cfg config.PullConsumerConfig) error
}

// PullConsumerRuntime represents one configured JetStream pull-consumer
// runtime.
type PullConsumerRuntime interface {
	Start(ctx context.Context)
}

// PullConsumerRuntimeFactory builds the runtime used by the broker after the
// config is validated.
type PullConsumerRuntimeFactory interface {
	Create(
		nc *nats.Conn,
		log logging.Logger,
		cfg config.PullConsumerConfig,
		handler eventbroker.MessageHandler,
	) (PullConsumerRuntime, error)
}
