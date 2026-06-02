package appfx

import (
	"context"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"go.uber.org/fx"
	"review-service/config"
	app "review-service/internal/application"
	pubnats "review-service/internal/presentation/event_broker/nats"
	natsbootstrap "review-service/pkg/messaging/nats"
)

// MessagingModule wires review-service NATS bootstrap into the FX lifecycle.
var MessagingModule = fx.Options(
	fx.Invoke(InvokeEnsureStream),
	fx.Invoke(InvokeRunProjectionConsumers),
)

var ensureStream = natsbootstrap.EnsureStream

// InvokeEnsureStream ensures the JetStream stream review-service depends on.
func InvokeEnsureStream(cfg *config.Config, lg logging.Logger) error {
	if err := ensureStream(cfg.NATS, lg); err != nil {
		lg.Error("bootstrap jetstream resources failed", logging.Err(err))
		return err
	}
	return nil
}

// InvokeNoopLifecycle is kept for symmetry with other modules when messaging has no runtime consumer.
func InvokeNoopLifecycle(lc fx.Lifecycle) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error { return nil },
		OnStop:  func(context.Context) error { return nil },
	})
}

// InvokeRunProjectionConsumers starts the review projection consumers with the FX lifecycle.
func InvokeRunProjectionConsumers(lc fx.Lifecycle, subscriber *pubnats.ReviewProjectionSubscriber, bootstrap app.RatingBootstrapper, coordinator app.ReviewWindowCoordinator, lg logging.Logger) {
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
				stop()
				return err
			}
			lg.Info("completed rating bootstrap preload",
				logging.Operation("review.rating.bootstrap"),
			)
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
