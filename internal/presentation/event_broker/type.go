package eventbroker

import (
	"context"
	"review-service/config"
)

// EventBroker is the transport-agnostic publish contract used by review-service.
type EventBroker interface {
	Publish(ctx context.Context, subject string, payload []byte) error
	RunPullConsumer(ctx context.Context, cfg config.PullConsumerConfig, handler MessageHandler) error
	Close()
}

// MessageHandler handles a single broker message in a transport-agnostic form.
type MessageHandler func(ctx context.Context, subject string, payload []byte) error
