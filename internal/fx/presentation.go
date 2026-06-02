package appfx

import (
	"context"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"review-service/config"
	app "review-service/internal/application"
	domain "review-service/internal/domain"
	grpcclient "review-service/internal/infra/order/grpc"
	usergrpc "review-service/internal/infra/user/grpc"
	eventbroker "review-service/internal/presentation/event_broker"
	pubnats "review-service/internal/presentation/event_broker/nats"
	grpcserver "review-service/internal/presentation/grpc"

	"go.uber.org/fx"
)

// PresentationModule wires the gRPC server into the FX lifecycle.
var PresentationModule = fx.Options(
	fx.Provide(
		ProvideEventBroker,
		ProvideOrderLookupClient,
		ProvideUserPreviewClient,
		ProvideReviewPublisher,
		ProvideReviewProjectionSubscriber,
		grpcserver.NewReviewMapper,
		ProvideGRPCServer,
	),
	fx.Invoke(
		InvokeRunGRPCServer,
	),
)

// ProvideOrderLookupClient constructs the outbound order-service lookup client.
func ProvideOrderLookupClient(cfg *config.Config, lg logging.Logger) (app.OrderLookupClient, error) {
	return grpcclient.New(cfg.Order.Address, lg)
}

// ProvideUserPreviewClient constructs the outbound user-service preview client.
func ProvideUserPreviewClient(cfg *config.Config, lg logging.Logger) (app.UserPreviewClient, error) {
	return usergrpc.New(cfg.User.Address, lg)
}

// ProvideEventBroker constructs the NATS publisher used by review-service.
func ProvideEventBroker(cfg *config.Config, lg logging.Logger) (eventbroker.EventBroker, error) {
	return pubnats.NewBroker(cfg.NATS, lg)
}

// ProvideReviewPublisher constructs the review lifecycle publisher.
func ProvideReviewPublisher(broker eventbroker.EventBroker, cfg *config.Config) app.ReviewPublisher {
	return pubnats.NewReviewPublisher(broker, cfg)
}

// ProvideReviewProjectionSubscriber constructs the asynchronous read-model repair subscriber.
func ProvideReviewProjectionSubscriber(
	broker eventbroker.EventBroker,
	service app.ReviewService,
	readRepo domain.ReviewReadRepository,
	coord app.ReviewWindowCoordinator,
	userPreview app.UserPreviewClient,
	cfg *config.Config,
	lg logging.Logger,
) (*pubnats.ReviewProjectionSubscriber, error) {
	return pubnats.NewReviewProjectionSubscriber(broker, service, readRepo, coord, userPreview, cfg, lg)
}

// ProvideGRPCServer constructs the gRPC server exposed by review-service.
func ProvideGRPCServer(
	service app.ReviewService,
	mapper grpcserver.ReviewMapper,
	cfg *config.Config,
	lg logging.Logger,
) (grpcserver.Server, error) {
	return grpcserver.NewServer(service, mapper, cfg.GRPC, lg)
}

// InvokeRunGRPCServer starts and gracefully stops the gRPC server with the FX lifecycle.
func InvokeRunGRPCServer(lc fx.Lifecycle, srv grpcserver.Server) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				if err := srv.Start(); err != nil {
					panic(err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
}
