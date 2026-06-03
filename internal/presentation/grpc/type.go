package grpc

import (
	"context"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	reviewv1 "github.com/ofm-microservices/ofm-common/proto/review/v1"
	app "review-service/internal/application"
	domain "review-service/internal/domain"
)

// Logger aliases the shared logger contract used by the gRPC adapter.
type Logger = logging.Logger

// ReviewService is the application boundary used by the review gRPC server.
type ReviewService interface {
	CreateReview(ctx context.Context, cmd app.CreateReviewCommand) (*domain.Review, error)
	ListGigReviews(ctx context.Context, query app.ListGigReviewsQuery) (*domain.ListReviewsResult, error)
	ListSellerReviews(ctx context.Context, query app.ListSellerReviewsQuery) (*domain.ListReviewsResult, error)
	GetReviewsBySellerUsername(ctx context.Context, username, cursor string) (*domain.ListReviewsResult, error)
	GetGigRatingSummary(ctx context.Context, gigID string) (*domain.RatingSummary, error)
	GetUserRatingSummaryByUsername(ctx context.Context, username string) (*domain.RatingSummary, error)
}

// Review aliases the generated review payload returned to callers.
type Review = reviewv1.Review

// Server exposes the gRPC server lifecycle.
type Server interface {
	Start() error
	Shutdown(ctx context.Context) error
}
