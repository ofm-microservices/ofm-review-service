package grpc

import (
	"context"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	app "review-service/internal/application"
)

// OrderLookupClient is the adapter boundary used by review-service to read
// authoritative order state.
type OrderLookupClient interface {
	GetOrderLifecycleSnapshot(ctx context.Context, orderID string) (*app.OrderLifecycleSnapshot, error)
	Close() error
}

// Logger aliases the shared structured logger.
type Logger = logging.Logger
