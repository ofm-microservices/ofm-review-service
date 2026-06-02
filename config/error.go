package config

import (
	"fmt"
)

// WrapParseEnvConfigError annotates env parsing failures from Config loading.
func WrapParseEnvConfigError(err error) error {
	return fmt.Errorf("parse env config: %w", err)
}

// WrapInvalidReviewPaginationConfigError annotates invalid review pagination settings.
func WrapInvalidReviewPaginationConfigError(err error) error {
	return fmt.Errorf("validate review pagination config: %w", err)
}

// WrapInvalidReviewWindowCoordinatorConfigError annotates invalid review window coordinator settings.
func WrapInvalidReviewWindowCoordinatorConfigError(err error) error {
	return fmt.Errorf("validate review window coordinator config: %w", err)
}
