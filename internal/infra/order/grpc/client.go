package grpc

import (
	"context"
	"strings"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"github.com/ofm-microservices/ofm-common/pkg/observability/metrics"
	orderwritev1 "github.com/ofm-microservices/ofm-common/proto/orderwrite/v1"
	grpcpkg "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	app "review-service/internal/application"
)

type client struct {
	conn *grpcpkg.ClientConn
	cl   orderwritev1.OrderWriteServiceClient
	log  logging.Logger
}

// New constructs the review-service order lookup client.
func New(address string, log logging.Logger) (OrderLookupClient, error) {
	if strings.TrimSpace(address) == "" {
		return nil, ErrEmptyAddress
	}
	if log == nil {
		return nil, ErrNilLogger
	}
	conn, err := grpcpkg.NewClient(address, grpcpkg.WithTransportCredentials(insecure.NewCredentials()), grpcpkg.WithUnaryInterceptor(metrics.UnaryClientInterceptor()))
	if err != nil {
		return nil, err
	}
	return &client{
		conn: conn,
		cl:   orderwritev1.NewOrderWriteServiceClient(conn),
		log:  log.With(logging.String("module", "order-lookup-client"), logging.String("address", address)),
	}, nil
}

func (c *client) GetOrderLifecycleSnapshot(ctx context.Context, orderID string) (*app.OrderLifecycleSnapshot, error) {
	res, err := c.cl.GetOrderLifecycleSnapshot(ctx, &orderwritev1.GetOrderLifecycleSnapshotRequest{OrderId: strings.TrimSpace(orderID)})
	if err != nil {
		return nil, err
	}
	order := res.GetOrder()
	if order == nil {
		return &app.OrderLifecycleSnapshot{}, nil
	}
	return &app.OrderLifecycleSnapshot{
		OrderID: order.GetOrderId(),
		BuyerID: order.GetBuyerUserId(),
		GigID:   order.GetGigId(),
		SellerID: order.GetSellerUserId(),
		Status:  order.GetStatus(),
	}, nil
}

func (c *client) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}
