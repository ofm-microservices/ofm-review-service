package service

import (
	"context"
	"time"

	"github.com/ofm-microservices/ofm-common/pkg/cursor"
	"github.com/ofm-microservices/ofm-common/pkg/logging"
	domain "review-service/internal/domain"
)

// ReviewService owns buyer-authenticated review creation.
type ReviewService interface {
	CreateReview(ctx context.Context, cmd CreateReviewCommand) (*domain.Review, error)
	ListGigReviews(ctx context.Context, query ListGigReviewsQuery) (*domain.ListReviewsResult, error)
	ListSellerReviews(ctx context.Context, query ListSellerReviewsQuery) (*domain.ListReviewsResult, error)
	GetReviewsBySellerUsername(ctx context.Context, username, cursor string) (*domain.ListReviewsResult, error)
	GetGigRatingSummary(ctx context.Context, gigID string) (*domain.RatingSummary, error)
	GetUserRatingSummaryByUsername(ctx context.Context, username string) (*domain.RatingSummary, error)
	ProjectReview(ctx context.Context, review *domain.Review, policy ProjectionPolicy) (*ProjectionResult, error)
}

// CursorCodec encodes and decodes opaque pagination cursors shared with the
// review read path.
type CursorCodec = cursor.Codec

// PaginationConfig controls the fixed page and window sizing for review reads.
type PaginationConfig struct {
	PageSize   int
	WindowSize int
	WindowTTL  time.Duration
}

// OrderLookupClient resolves authoritative order state for review writes.
type OrderLookupClient interface {
	GetOrderLifecycleSnapshot(ctx context.Context, orderID string) (*OrderLifecycleSnapshot, error)
}

// UserPreviewClient resolves the author preview used by the review read model.
type UserPreviewClient interface {
	GetUserPreviewByID(ctx context.Context, userID string) (*UserPreview, error)
	GetUserPreviewByIDNoCache(ctx context.Context, userID string) (*UserPreview, error)
	GetDetailedUserByUsername(ctx context.Context, username string) (*UserPreview, error)
}

// FileURLClient resolves the public avatar URL for a file identifier.
type FileURLClient interface {
	GetFileURL(ctx context.Context, fileID string) (string, error)
	GetFileURLs(ctx context.Context, fileIDs []string) (map[string]string, error)
}

// ReviewPublisher emits review lifecycle events after successful writes.
type ReviewPublisher interface {
	PublishGigReviewProjectionRequested(ctx context.Context, review *domain.Review) error
	PublishUserReviewProjectionRequested(ctx context.Context, review *domain.Review) error
	PublishGigRatingRequested(ctx context.Context, review *domain.Review) error
	PublishSellerRatingRequested(ctx context.Context, review *domain.Review) error
}

// RatingBootstrapper warms the rating read model before event consumers start.
type RatingBootstrapper interface {
	Preload(ctx context.Context) error
}

// ReviewWindowCoordinator serializes review-window mutations through the Redis
// queue and lease mechanism.
type ReviewWindowCoordinator interface {
	Start(ctx context.Context) error
	EnqueueGigReview(ctx context.Context, review *domain.Review) error
	EnqueueSellerReview(ctx context.Context, review *domain.Review) error
}

// WindowCoordinatorConfig controls the lease and reconciler timings used by
// the review-window coordinator.
type WindowCoordinatorConfig struct {
	LeaseTTL                time.Duration
	LeaseRenewInterval      time.Duration
	ReconcileInterval       time.Duration
	QueueDepthWarnThreshold int
	AckAfterEnqueue         bool
}

// OrderLifecycleSnapshot carries the order state required to validate and
// enrich a review write.
type OrderLifecycleSnapshot struct {
	OrderID        string
	BuyerID        string
	GigID          string
	SellerID       string
	SellerUsername string
	Status         string
}

// UserPreview carries the preview data needed for read-model enrichment.
type UserPreview struct {
	UserID      string
	Username    string
	DisplayName string
	AvatarID    string
	AvatarURL   string
}

// CreateReviewCommand is the application input for a new review.
type CreateReviewCommand struct {
	OrderID        string
	BuyerID        string
	BuyerUsername  string
	Content        string
	Rating         int32
	IdempotencyKey string
	RequestedAt    string
}

// ProjectionPolicy controls which enrichments the projection worker should
// attempt for a given review payload.
type ProjectionPolicy struct {
	EnrichAuthor bool
	EnrichAvatar bool
}

// ProjectionResult reports the enriched payload and any retry jobs that should
// be emitted by the caller.
type ProjectionResult struct {
	Review      *domain.Review
	AuthorRetry bool
	AvatarRetry bool
}

// ListGigReviewsQuery is the application input for gig review pagination.
type ListGigReviewsQuery struct {
	GigID  string
	Cursor string
}

// ListSellerReviewsQuery is the application input for seller review pagination.
type ListSellerReviewsQuery struct {
	SellerID string
	Cursor   string
}

// Logger aliases the shared structured logger used by the application layer.
type Logger = logging.Logger
