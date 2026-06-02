package config

import "time"

// ReviewWindowCoordinatorConfig controls the Redis queue and lease behaviour
// used to serialize review window mutations.
type ReviewWindowCoordinatorConfig struct {
	LeaseTTL                time.Duration `env:"REVIEW_QUEUE_LEASE_TTL" envDefault:"30s"`
	LeaseRenewInterval      time.Duration `env:"REVIEW_QUEUE_LEASE_RENEW_INTERVAL" envDefault:"10s"`
	ReconcileInterval       time.Duration `env:"REVIEW_QUEUE_RECONCILE_INTERVAL" envDefault:"30s"`
	QueueDepthWarnThreshold int           `env:"REVIEW_QUEUE_DEPTH_WARN_THRESHOLD" envDefault:"1000"`
	AckAfterEnqueue         bool          `env:"REVIEW_QUEUE_ACK_AFTER_ENQUEUE" envDefault:"true"`
}
