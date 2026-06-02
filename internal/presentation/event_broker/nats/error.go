package nats

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyNATSURL                 = errors.New("nats url is empty")
	ErrNilLogger                    = errors.New("logger is nil")
	ErrNilConfig                    = errors.New("config is nil")
	ErrNilEventBroker               = errors.New("event broker is nil")
	ErrNilReviewService             = errors.New("review service is nil")
	ErrNilReviewReadRepository      = errors.New("review read repository is nil")
	ErrNilReviewWindowCoordinator   = errors.New("review window coordinator is nil")
	ErrNilUserPreviewClient         = errors.New("user preview client is nil")
	ErrEmptyStreamName              = errors.New("stream name is empty")
	ErrEmptySubject                 = errors.New("subject is empty")
	ErrEmptyDurableName             = errors.New("durable name is empty")
	ErrInvalidBatchSize             = errors.New("batch size must be greater than zero")
	ErrInvalidMaxWait               = errors.New("max wait must be greater than zero")
	ErrInvalidWorkerCount           = errors.New("worker count must be greater than zero")
	ErrInvalidQueueSize             = errors.New("queue size must be greater than zero")
	ErrInvalidAckWait               = errors.New("ack wait must be greater than zero")
	ErrInvalidMaxDeliver            = errors.New("max deliver must be greater than zero")
	ErrInvalidDeliverPolicy         = errors.New("deliver policy must be empty, all, new, or last")
	ErrInvalidAdaptiveCheckInterval = errors.New("adaptive check interval must be greater than zero")
	ErrInvalidAdaptiveThresholds    = errors.New("adaptive thresholds must satisfy: high > medium >= 0")
	ErrInvalidAdaptivePlan          = errors.New("adaptive plan batch size and max wait must be greater than zero")
)

// AnnotateConnectToNATSError annotates low-level NATS connection failures.
func AnnotateConnectToNATSError(err error) error {
	return fmt.Errorf("connect to nats: %w", err)
}

// AnnotatePublishToNATSError annotates broker publish failures.
func AnnotatePublishToNATSError(subject string, err error) error {
	return fmt.Errorf("publish to nats (%s): %w", subject, err)
}

// AnnotateFlushNATSPublisherError annotates publisher flush failures.
func AnnotateFlushNATSPublisherError(err error) error {
	return fmt.Errorf("flush nats publisher: %w", err)
}

// AnnotateInitJetStreamContextError annotates JetStream initialization failures.
func AnnotateInitJetStreamContextError(err error) error {
	return fmt.Errorf("init jetstream context: %w", err)
}

// AnnotateEnsureConsumerError annotates add-or-update pull-consumer failures.
func AnnotateEnsureConsumerError(streamName, durableName string, addErr, updateErr error) error {
	return fmt.Errorf("ensure consumer %q in stream %q: add err=%v, update err=%w", durableName, streamName, addErr, updateErr)
}

// AnnotateCreatePullSubscriberError annotates pull-subscription creation failures.
func AnnotateCreatePullSubscriberError(subject, durable string, err error) error {
	return fmt.Errorf("create pull subscriber subject=%q durable=%q: %w", subject, durable, err)
}
