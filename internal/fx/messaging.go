package appfx

import (
	"context"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"go.uber.org/fx"
	app "review-service/internal/application"
	pubkafka "review-service/internal/presentation/event_broker/kafka"
)

// MessagingModule wires review-service Kafka consumers into the FX lifecycle.
var MessagingModule = fx.Options(
	fx.Invoke(InvokeRunProjectionConsumers),
)

// InvokeNoopLifecycle is kept for symmetry with other modules when messaging has no runtime consumer.
func InvokeNoopLifecycle(lc fx.Lifecycle) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error { return nil },
		OnStop:  func(context.Context) error { return nil },
	})
}

// InvokeRunProjectionConsumers starts the review projection consumers with the FX lifecycle.
func InvokeRunProjectionConsumers(lc fx.Lifecycle, subscriber *pubkafka.ReviewProjectionSubscriber, bootstrap app.RatingBootstrapper, coordinator app.ReviewWindowCoordinator, lg logging.Logger) {
	var cancel context.CancelFunc
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			runCtx, stop := context.WithCancel(context.Background())
			cancel = stop
			if err := coordinator.Start(runCtx); err != nil {
				stop()
				return err
			}
			lg.Info("starting rating bootstrap preload",
				logging.Operation("review.rating.bootstrap"),
			)
			if err := bootstrap.Preload(runCtx); err != nil {
				lg.Warn("rating bootstrap preload failed; continuing with live projection consumers",
					logging.Operation("review.rating.bootstrap"),
					logging.Err(err),
				)
			} else {
				lg.Info("completed rating bootstrap preload",
					logging.Operation("review.rating.bootstrap"),
				)
			}
			return subscriber.Start(runCtx)
		},
		OnStop: func(context.Context) error {
			if cancel != nil {
				cancel()
			}
			return nil
		},
	})
}
