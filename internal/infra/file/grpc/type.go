package grpc

import "context"

// FileURLClient exposes file URL resolution used by review-service.
type FileURLClient interface {
	GetFileURL(ctx context.Context, fileID string) (string, error)
	GetFileURLs(ctx context.Context, fileIDs []string) (map[string]string, error)
	Close() error
}
