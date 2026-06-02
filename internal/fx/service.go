package appfx

import (
	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"review-service/config"
	app "review-service/internal/application"
	review "review-service/internal/domain"

	"go.uber.org/fx"
)

// ServiceModule provides the application service used by review-service.
var ServiceModule = fx.Options(
	fx.Provide(
		ProvideReviewService,
		ProvideRatingBootstrapper,
		ProvideReviewWindowCoordinator,
	),
)

// ProvideReviewService constructs the review-service application service.
func ProvideReviewService(writeRepo review.ReviewRepository, readRepo review.ReviewReadRepository, orderLookup app.OrderLookupClient, userPreview app.UserPreviewClient, pub app.ReviewPublisher, codec app.CursorCodec, paging app.PaginationConfig, lg logging.Logger) (app.ReviewService, error) {
	return app.New(writeRepo, readRepo, orderLookup, userPreview, pub, codec, paging, lg)
}

// ProvideRatingBootstrapper constructs the startup warmer for rating summaries.
func ProvideRatingBootstrapper(writeRepo review.ReviewRepository, readRepo review.ReviewReadRepository, userPreview app.UserPreviewClient, lg logging.Logger) (app.RatingBootstrapper, error) {
	return app.NewRatingBootstrapper(writeRepo, readRepo, userPreview, lg)
}

// ProvideReviewWindowCoordinator constructs the Redis-backed review-window coordinator.
func ProvideReviewWindowCoordinator(writeRepo review.ReviewRepository, projectionRepo review.ReviewProjectionRepository, service app.ReviewService, paging app.PaginationConfig, cfg *config.Config, lg logging.Logger) (app.ReviewWindowCoordinator, error) {
	return app.NewReviewWindowCoordinator(projectionRepo, writeRepo, service, paging, app.WindowCoordinatorConfig{
		LeaseTTL:                cfg.Window.LeaseTTL,
		LeaseRenewInterval:      cfg.Window.LeaseRenewInterval,
		ReconcileInterval:       cfg.Window.ReconcileInterval,
		QueueDepthWarnThreshold: cfg.Window.QueueDepthWarnThreshold,
		AckAfterEnqueue:         cfg.Window.AckAfterEnqueue,
	}, lg)
}
