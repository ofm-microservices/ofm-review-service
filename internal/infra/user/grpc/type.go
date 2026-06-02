package grpc

import (
	"context"

	app "review-service/internal/application"
)

// UserPreviewClient exposes the user-preview lookup used by review-service.
type UserPreviewClient interface {
	GetUserPreviewByID(ctx context.Context, userID string) (*app.UserPreview, error)
	GetUserPreviewByIDNoCache(ctx context.Context, userID string) (*app.UserPreview, error)
	GetDetailedUserByUsername(ctx context.Context, username string) (*app.UserPreview, error)
	Close() error
}
