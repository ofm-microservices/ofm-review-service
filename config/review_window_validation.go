package config

import "fmt"

func validateReviewWindowCoordinatorConfig(cfg ReviewWindowCoordinatorConfig) error {
	switch {
	case cfg.LeaseTTL <= 0:
		return WrapInvalidReviewWindowCoordinatorConfigError(fmt.Errorf("REVIEW_QUEUE_LEASE_TTL must be greater than zero"))
	case cfg.LeaseRenewInterval <= 0:
		return WrapInvalidReviewWindowCoordinatorConfigError(fmt.Errorf("REVIEW_QUEUE_LEASE_RENEW_INTERVAL must be greater than zero"))
	case cfg.ReconcileInterval <= 0:
		return WrapInvalidReviewWindowCoordinatorConfigError(fmt.Errorf("REVIEW_QUEUE_RECONCILE_INTERVAL must be greater than zero"))
	case cfg.QueueDepthWarnThreshold < 0:
		return WrapInvalidReviewWindowCoordinatorConfigError(fmt.Errorf("REVIEW_QUEUE_DEPTH_WARN_THRESHOLD must not be negative"))
	}
	return nil
}
