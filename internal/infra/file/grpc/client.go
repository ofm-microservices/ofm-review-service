package grpc

import (
	"context"
	"strings"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"github.com/ofm-microservices/ofm-common/pkg/observability/metrics"
	filev1 "github.com/ofm-microservices/ofm-common/proto/file/v1"
	grpcpkg "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type client struct {
	conn *grpcpkg.ClientConn
	cl   filev1.FileServiceClient
	log  logging.Logger
}

// New constructs the review-service file URL client.
func New(address string, log logging.Logger) (FileURLClient, error) {
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
		cl:   filev1.NewFileServiceClient(conn),
		log:  log.With(logging.String("module", "file-url-client"), logging.String("address", address)),
	}, nil
}

func (c *client) GetFileURL(ctx context.Context, fileID string) (string, error) {
	res, err := c.cl.GetFileURL(ctx, &filev1.GetFileURLRequest{FileId: strings.TrimSpace(fileID)})
	if err != nil {
		return "", err
	}
	return res.GetUrl(), nil
}

func (c *client) GetFileURLs(ctx context.Context, fileIDs []string) (map[string]string, error) {
	if len(fileIDs) == 0 {
		return nil, nil
	}
	res, err := c.cl.GetFileURLs(ctx, &filev1.GetFileURLsRequest{FileIds: fileIDs})
	if err != nil {
		return nil, err
	}
	urls := make(map[string]string, len(res.GetFileUrls()))
	for _, item := range res.GetFileUrls() {
		if item == nil {
			continue
		}
		urls[item.GetFileId()] = item.GetUrl()
	}
	if len(urls) == 0 {
		return nil, nil
	}
	return urls, nil
}

func (c *client) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}
