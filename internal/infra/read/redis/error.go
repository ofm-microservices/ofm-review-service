package repository

import (
	"errors"
	"fmt"
)

var (
	ErrNilRedisClient      = errors.New("redis client is nil")
	ErrNilReview           = errors.New("review is nil")
	ErrNilLogger           = errors.New("logger is nil")
	ErrInvalidCursor       = errors.New("invalid review cursor")
	ErrInvalidReviewMember = errors.New("invalid review zset member")
)

// AnnotateMarshalReviewCacheError annotates cache serialization failures.
func AnnotateMarshalReviewCacheError(err error) error {
	return fmt.Errorf("marshal review cache: %w", err)
}

// AnnotateUnmarshalReviewCacheError annotates cache deserialization failures.
func AnnotateUnmarshalReviewCacheError(err error) error {
	return fmt.Errorf("unmarshal review cache: %w", err)
}

// AnnotateSetReviewCacheError annotates Redis upsert failures for the review cache.
func AnnotateSetReviewCacheError(key string, err error) error {
	return fmt.Errorf("set review cache by key %q: %w", key, err)
}

// AnnotateDeleteReviewCacheError annotates Redis delete failures for the review cache.
func AnnotateDeleteReviewCacheError(key string, err error) error {
	return fmt.Errorf("delete review cache by key %q: %w", key, err)
}

// AnnotateEnqueueReviewWindowJobError annotates Redis enqueue failures for the
// review window coordinator.
func AnnotateEnqueueReviewWindowJobError(key string, err error) error {
	return fmt.Errorf("enqueue review window job by key %q: %w", key, err)
}

// AnnotateQueueDepthReviewError annotates Redis queue-length failures.
func AnnotateQueueDepthReviewError(key string, err error) error {
	return fmt.Errorf("queue depth by key %q: %w", key, err)
}

// AnnotateQueuePeekReviewError annotates Redis queue peek failures.
func AnnotateQueuePeekReviewError(key string, err error) error {
	return fmt.Errorf("peek review window job by key %q: %w", key, err)
}

// AnnotateQueuePopReviewError annotates Redis queue pop failures.
func AnnotateQueuePopReviewError(key string, err error) error {
	return fmt.Errorf("pop review window job by key %q: %w", key, err)
}

// AnnotateActiveQueueReviewError annotates active-queue set failures.
func AnnotateActiveQueueReviewError(key string, err error) error {
	return fmt.Errorf("active queue operation by key %q: %w", key, err)
}

// AnnotateLeaseReviewError annotates Redis lease coordination failures.
func AnnotateLeaseReviewError(key string, err error) error {
	return fmt.Errorf("lease operation by key %q: %w", key, err)
}

// AnnotateLoadReviewWindowSnapshotsError annotates review-window snapshot
// loading failures.
func AnnotateLoadReviewWindowSnapshotsError(prefix string, err error) error {
	return fmt.Errorf("load review window snapshots for prefix %q: %w", prefix, err)
}

// AnnotateApplyReviewWindowOpsError annotates review-window mutation failures.
func AnnotateApplyReviewWindowOpsError(prefix string, err error) error {
	return fmt.Errorf("apply review window ops for prefix %q: %w", prefix, err)
}

// AnnotateDurabilityReviewError annotates Redis persistence inspection failures.
func AnnotateDurabilityReviewError(err error) error {
	return fmt.Errorf("check review queue durability: %w", err)
}
