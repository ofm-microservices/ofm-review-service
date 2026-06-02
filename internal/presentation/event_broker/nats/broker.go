package nats

import (
	"context"
	"time"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"github.com/ofm-microservices/ofm-common/pkg/observability/natstrace"
	"review-service/config"
	eventbroker "review-service/internal/presentation/event_broker"

	"github.com/nats-io/nats.go"
)

type natsConn interface {
	PublishMsg(msg *nats.Msg) error
	FlushWithContext(ctx context.Context) error
	Close()
}

type natsBroker struct {
	nc             natsConn
	jsConn         *nats.Conn
	log            logging.Logger
	validator      PullConsumerConfigValidator
	runtimeFactory PullConsumerRuntimeFactory
}

// NewBroker constructs the review-service NATS publisher.
func NewBroker(cfg config.NATSConfig, log logging.Logger) (eventbroker.EventBroker, error) {
	if cfg.URL == "" {
		return nil, ErrEmptyNATSURL
	}
	if log == nil {
		return nil, ErrNilLogger
	}
	opts := []nats.Option{nats.Name("review-service"), nats.MaxReconnects(-1)}
	if cfg.Review != "" {
		opts = append(opts, nats.UserInfo(cfg.Review, cfg.Password))
	}
	nc, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, err
	}
	return &natsBroker{
		nc:             nc,
		jsConn:         nc,
		log:            log.With(logging.String("module", "nats-broker"), logging.String("nats_url", cfg.URL)),
		validator:      NewPullConsumerConfigValidator(),
		runtimeFactory: NewPullConsumerRuntimeFactory(),
	}, nil
}

func (b *natsBroker) Publish(ctx context.Context, subject string, payload []byte) error {
	started := time.Now()
	if err := b.nc.PublishMsg(natstrace.NewMessage(ctx, subject, payload)); err != nil {
		b.log.Error("message publish failed",
			logging.Operation("nats.publish"),
			logging.Attempt(1),
			logging.Retryable(true),
			logging.DurationMS(time.Since(started)),
			logging.String("subject", subject),
			logging.Err(err),
		)
		return err
	}
	if err := Flush(ctx, b.nc); err != nil {
		return err
	}
	return nil
}

func (b *natsBroker) RunPullConsumer(ctx context.Context, cfg config.PullConsumerConfig, handler eventbroker.MessageHandler) error {
	validator := b.validator
	if validator == nil {
		validator = NewPullConsumerConfigValidator()
	}
	if err := validator.Validate(cfg); err != nil {
		return err
	}

	runtimeFactory := b.runtimeFactory
	if runtimeFactory == nil {
		runtimeFactory = NewPullConsumerRuntimeFactory()
	}
	runtime, err := runtimeFactory.Create(b.jsConn, b.log, cfg, handler)
	if err != nil {
		return err
	}

	b.log.Info("starting pull consumer",
		logging.Operation("nats.pull_consumer.start"),
		logging.String("stream", cfg.Stream),
		logging.String("subject", cfg.Subject),
		logging.String("durable", cfg.Durable),
		logging.Int("batch_size", cfg.BatchSize),
		logging.Any("max_wait", cfg.MaxWait),
		logging.Int("workers", cfg.Workers),
	)
	if cfg.Adaptive.Enabled {
		b.log.Info("adaptive pull plans enabled",
			logging.Operation("nats.pull_consumer.adaptive"),
			logging.String("subject", cfg.Subject),
			logging.Int("medium_pending", cfg.Adaptive.MediumPending),
			logging.Int("high_pending", cfg.Adaptive.HighPending),
		)
	}

	runtime.Start(ctx)
	return nil
}

func (b *natsBroker) Close() {
	if b.nc != nil {
		b.nc.Close()
	}
}

// Flush blocks until buffered NATS publications are acknowledged or the
// supplied context expires.
func Flush(ctx context.Context, nc natsConn) error {
	flushCtx := ctx
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		flushCtx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
	}

	return nc.FlushWithContext(flushCtx)
}
