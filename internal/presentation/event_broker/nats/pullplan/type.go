package pullplan

import (
	"time"

	"review-service/config"
)

// Resolver chooses an adaptive pull-consumer plan.
type Resolver interface {
	Resolve(cfg config.PullConsumerConfig, pending int) (string, int, time.Duration)
}
