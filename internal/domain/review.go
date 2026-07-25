package review

import (
	"context"
	"time"
)

// RatingSummary captures the full review rating aggregate for a gig or seller.
type RatingSummary struct {
	RatingAvg    float64
	TotalReviews int64
	Stars5       int64
	Stars4       int64
	Stars3       int64
	Stars2       int64
	Stars1       int64
}

// GigRatingSummary couples a gig identifier with its aggregate rating data.
type GigRatingSummary struct {
	GigID string
	RatingSummary
}

// SellerRatingSummary couples a seller identifier with its aggregate rating data.
type SellerRatingSummary struct {
	SellerID string
	RatingSummary
}

// Review is the write-model entity owned by review-service.
type Review struct {
	ID             string        `json:"review_id"`
	OrderID        string        `json:"order_id"`
	GigID          string        `json:"gig_id"`
	Content        string        `json:"content"`
	BuyerID        string        `json:"buyer_id"`
	BuyerUsername  string        `json:"buyer_username,omitempty"`
	Rating         int32         `json:"rating"`
	SellerID       string        `json:"seller_id"`
	SellerUsername string        `json:"seller_username"`
	Author         *ReviewAuthor `json:"author,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

// ReviewAuthor carries the author preview projected into Redis.
type ReviewAuthor struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	AvatarID    string `json:"avatar_id,omitempty"`
}

// CreateReviewParams contains the input required to create a new review.
type CreateReviewParams struct {
	ID             string
	OrderID        string
	GigID          string
	Content        string
	BuyerID        string
	BuyerUsername  string
	SellerID       string
	SellerUsername string
	Rating         int32
}

// ReviewRepository persists the review-service write model.
type ReviewRepository interface {
	Create(ctx context.Context, params CreateReviewParams) (*Review, error)
	GetByID(ctx context.Context, reviewID string) (*Review, error)
	GetByOrderID(ctx context.Context, orderID string) (*Review, error)
	GetSellerIDByUsername(ctx context.Context, username string) (string, error)
	ListByGigID(ctx context.Context, query ListReviewsQuery) (*ListReviewsResult, error)
	ListBySellerID(ctx context.Context, query ListReviewsQuery) (*ListReviewsResult, error)
	GetGigRatingSummary(ctx context.Context, gigID string) (*RatingSummary, error)
	GetSellerRatingSummary(ctx context.Context, sellerID string) (*RatingSummary, error)
	ListGigRatingSummaries(ctx context.Context) ([]GigRatingSummary, error)
	ListSellerRatingSummaries(ctx context.Context) ([]SellerRatingSummary, error)
	DeleteByID(ctx context.Context, reviewID string) error
}

// ReviewReadRepository persists the review-service read-model projection.
type ReviewReadRepository interface {
	UpsertGigRating(ctx context.Context, gigID string, rating int32) error
	UpsertSellerRating(ctx context.Context, sellerID string, rating int32) error
	UpsertSellerRatingByUsername(ctx context.Context, username string, rating int32) error
	GetSellerIDByUsername(ctx context.Context, username string) (string, error)
	GetGigRatingSummary(ctx context.Context, gigID string) (*RatingSummary, error)
	GetSellerRatingSummary(ctx context.Context, sellerID string) (*RatingSummary, error)
	GetSellerRatingSummaryByUsername(ctx context.Context, username string) (*RatingSummary, error)
	SetSellerIDByUsername(ctx context.Context, username, sellerID string) error
	SetGigRating(ctx context.Context, gigID string, summary RatingSummary) error
	SetSellerRating(ctx context.Context, sellerID string, summary RatingSummary) error
	SetSellerRatingByUsername(ctx context.Context, username string, summary RatingSummary) error
	ListByGigID(ctx context.Context, query ListReviewsQuery) (*ListReviewsResult, error)
	ListBySellerID(ctx context.Context, query ListReviewsQuery) (*ListReviewsResult, error)
	ListByGigWindow(ctx context.Context, gigID string, window int) (*ListReviewsResult, error)
	ListBySellerWindow(ctx context.Context, sellerID string, window int) (*ListReviewsResult, error)
	UpsertGigWindow(ctx context.Context, gigID string, window int, reviews []*Review, hasMore bool, ttl time.Duration) error
	UpsertSellerWindow(ctx context.Context, sellerID string, window int, reviews []*Review, hasMore bool, ttl time.Duration) error
	DeleteByID(ctx context.Context, reviewID string) error
}

// ListReviewsQuery describes a cursor-paged review read-model lookup.
type ListReviewsQuery struct {
	GigID    string
	SellerID string
	Cursor   string
	Limit    int
}

// ListReviewsResult contains one page of reviews from a Redis read model.
type ListReviewsResult struct {
	Reviews []*Review
	Cursor  string
	HasMore bool
}

// ReviewMember couples a review payload with the Redis score used to preserve
// the newest-first ordering across review windows.
type ReviewMember struct {
	Review *Review
	Score  float64
}

// WindowSnapshot describes the current boundaries for one cached review window.
type WindowSnapshot struct {
	Window int
	Count  int
	Newest ReviewMember
	Oldest ReviewMember
}

// WindowOp describes one atomic review-window mutation step.
type WindowOp struct {
	Window int
	Insert []ReviewMember
	Evict  []ReviewMember
}

// ReviewProjectionRepository owns the Redis queue, lease, and review-window
// mutation contract used by the review projection coordinator.
type ReviewProjectionRepository interface {
	EnqueueWindowJob(ctx context.Context, queueKey, activeKey string, review *Review) (int64, error)
	QueueLength(ctx context.Context, queueKey string) (int64, error)
	PeekWindowJob(ctx context.Context, queueKey string) (*Review, error)
	PopWindowJob(ctx context.Context, queueKey string) (*Review, error)
	ActiveQueueKeys(ctx context.Context, activeKey string) ([]string, error)
	RemoveActiveQueueIfEmpty(ctx context.Context, activeKey, queueKey string) (bool, error)
	TryAcquireLease(ctx context.Context, leaseKey, token string, ttl time.Duration) (bool, error)
	RenewLease(ctx context.Context, leaseKey, token string, ttl time.Duration) (bool, error)
	ReleaseLease(ctx context.Context, leaseKey, token string) (bool, error)
	LoadWindowSnapshots(ctx context.Context, ownerPrefix string) ([]WindowSnapshot, error)
	ApplyWindowOpsAndPopJob(ctx context.Context, queueKey, leaseKey, token, ownerPrefix string, reviewID string, ops []WindowOp, ttl time.Duration) (bool, error)
	CheckDurability(ctx context.Context) (bool, error)
}

// ReviewRatingRepository fetches full rating aggregates from the canonical write model.
type ReviewRatingRepository interface {
	GetGigRatingSummary(ctx context.Context, gigID string) (*RatingSummary, error)
	GetSellerRatingSummary(ctx context.Context, sellerID string) (*RatingSummary, error)
}
