package config

import "fmt"

func validateReviewPaginationConfig(cfg ReviewPaginationConfig) error {
	switch {
	case cfg.PageSize <= 0:
		return WrapInvalidReviewPaginationConfigError(fmt.Errorf("REVIEW_PAGE_SIZE must be greater than zero"))
	case cfg.WindowSize <= 0:
		return WrapInvalidReviewPaginationConfigError(fmt.Errorf("REVIEW_WINDOW_SIZE must be greater than zero"))
	case cfg.WindowSize%cfg.PageSize != 0:
		return WrapInvalidReviewPaginationConfigError(fmt.Errorf("REVIEW_WINDOW_SIZE must be divisible by REVIEW_PAGE_SIZE"))
	case cfg.WindowTTL < 0:
		return WrapInvalidReviewPaginationConfigError(fmt.Errorf("REVIEW_WINDOW_TTL must not be negative"))
	}
	return nil
}
