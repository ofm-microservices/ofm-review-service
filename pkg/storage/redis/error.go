package redis

import "fmt"

// AnnotateRedisPingError annotates Redis connectivity check failures.
func AnnotateRedisPingError(err error) error {
	return fmt.Errorf("ping redis: %w", err)
}
