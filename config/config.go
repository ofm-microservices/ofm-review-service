package config

import "github.com/caarlos0/env/v11"

// Config groups the full review-service runtime configuration.
type Config struct {
	App     AppConfig
	DB      DBConfig
	GRPC    GRPCConfig
	Paging  ReviewPaginationConfig
	Window  ReviewWindowCoordinatorConfig
	Order   OrderServiceConfig
	User    UserServiceConfig
	File    FileServiceConfig
	Metrics MetricsConfig
	Tracing TracingConfig
	Redis   RedisConfig
	Kafka   KafkaConfig
	NATS    NATSConfig
}

// Load reads environment variables into Config and applies defaults.
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, WrapParseEnvConfigError(err)
	}
	if err := validateReviewPaginationConfig(cfg.Paging); err != nil {
		return nil, err
	}
	if err := validateReviewWindowCoordinatorConfig(cfg.Window); err != nil {
		return nil, err
	}

	return cfg, nil
}
