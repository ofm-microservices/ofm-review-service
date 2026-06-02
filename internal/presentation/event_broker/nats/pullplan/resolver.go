package pullplan

import (
	"time"

	"review-service/config"
)

type resolver struct{}

// New constructs the adaptive pull-plan resolver.
func New() Resolver {
	return resolver{}
}

func (resolver) Resolve(cfg config.PullConsumerConfig, pending int) (string, int, time.Duration) {
	if !cfg.Adaptive.Enabled {
		return "base", cfg.BatchSize, cfg.MaxWait
	}
	if pending >= cfg.Adaptive.HighPending {
		return "high", cfg.Adaptive.HighBatchSize, cfg.Adaptive.HighMaxWait
	}
	if pending >= cfg.Adaptive.MediumPending {
		return "medium", cfg.Adaptive.MediumBatchSize, cfg.Adaptive.MediumMaxWait
	}
	return "low", cfg.Adaptive.LowBatchSize, cfg.Adaptive.LowMaxWait
}
