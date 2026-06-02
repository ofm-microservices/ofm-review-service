package appfx

import (
	"github.com/ofm-microservices/ofm-common/pkg/cursor"
	"review-service/config"
	app "review-service/internal/application"

	"go.uber.org/fx"
)

// ConfigModule provides parsed runtime configuration for review-service.
var ConfigModule = fx.Options(
	fx.Provide(
		ProvideConfig,
		ProvidePaginationConfig,
		ProvideWindowCoordinatorConfig,
		ProvideCursorCodec,
	),
)

// ProvideConfig loads the service configuration from environment variables.
func ProvideConfig() (*config.Config, error) {
	return config.Load()
}

// ProvidePaginationConfig adapts review-service config into the application layer pagination config.
func ProvidePaginationConfig(cfg *config.Config) app.PaginationConfig {
	return app.PaginationConfig{
		PageSize:   cfg.Paging.PageSize,
		WindowSize: cfg.Paging.WindowSize,
		WindowTTL:  cfg.Paging.WindowTTL,
	}
}

// ProvideCursorCodec constructs the shared cursor codec for encrypted pagination tokens.
func ProvideCursorCodec(cfg *config.Config) (app.CursorCodec, error) {
	return cursor.NewCodec(cursor.Config{Secret: cfg.Paging.CursorSecret})
}

// ProvideWindowCoordinatorConfig adapts review-service config into the
// application layer queue and lease settings.
func ProvideWindowCoordinatorConfig(cfg *config.Config) app.WindowCoordinatorConfig {
	return app.WindowCoordinatorConfig{
		LeaseTTL:                cfg.Window.LeaseTTL,
		LeaseRenewInterval:      cfg.Window.LeaseRenewInterval,
		ReconcileInterval:       cfg.Window.ReconcileInterval,
		QueueDepthWarnThreshold: cfg.Window.QueueDepthWarnThreshold,
		AckAfterEnqueue:         cfg.Window.AckAfterEnqueue,
	}
}
