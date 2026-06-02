package grpc

import (
	"context"
	"fmt"
	"strings"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"github.com/ofm-microservices/ofm-common/pkg/observability/metrics"
	userv1 "github.com/ofm-microservices/ofm-common/proto/user/v1"
	grpcpkg "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	app "review-service/internal/application"
)

type client struct {
	conn *grpcpkg.ClientConn
	cl   userv1.UserQueryServiceClient
	log  logging.Logger
}

// New constructs the review-service user preview client.
func New(address string, log logging.Logger) (UserPreviewClient, error) {
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
		cl:   userv1.NewUserQueryServiceClient(conn),
		log:  log.With(logging.String("module", "user-preview-client"), logging.String("address", address)),
	}, nil
}

func (c *client) GetUserPreviewByID(ctx context.Context, userID string) (*app.UserPreview, error) {
	res, err := c.cl.GetUserPreviewByID(ctx, &userv1.GetUserPreviewByIDRequest{UserId: strings.TrimSpace(userID)})
	if err != nil {
		c.log.Error("get user preview failed",
			logging.Operation("grpc.user.preview"),
			logging.String("user_id", strings.TrimSpace(userID)),
			logging.Err(err),
		)
		return nil, err
	}
	user := res.GetUser()
	if user == nil {
		err := fmt.Errorf("%w: user_id=%s", ErrEmptyUserResponse, strings.TrimSpace(userID))
		c.log.Error("get user preview returned empty user",
			logging.Operation("grpc.user.preview"),
			logging.String("user_id", strings.TrimSpace(userID)),
			logging.Err(err),
		)
		return nil, err
	}
	return &app.UserPreview{
		UserID:      user.GetUserId(),
		Username:    user.GetUsername(),
		DisplayName: user.GetDisplayName(),
		AvatarID:    user.GetAvatarId(),
		AvatarURL:   user.GetAvatarUrl(),
	}, nil
}

func (c *client) GetUserPreviewByIDNoCache(ctx context.Context, userID string) (*app.UserPreview, error) {
	res, err := c.cl.GetUserPreviewByIDNoCache(ctx, &userv1.GetUserPreviewByIDNoCacheRequest{UserId: strings.TrimSpace(userID)})
	if err != nil {
		c.log.Error("get user preview no cache failed",
			logging.Operation("grpc.user.preview_no_cache"),
			logging.String("user_id", strings.TrimSpace(userID)),
			logging.Err(err),
		)
		return nil, err
	}
	user := res.GetUser()
	if user == nil {
		err := fmt.Errorf("%w: user_id=%s", ErrEmptyUserResponse, strings.TrimSpace(userID))
		c.log.Error("get user preview no cache returned empty user",
			logging.Operation("grpc.user.preview_no_cache"),
			logging.String("user_id", strings.TrimSpace(userID)),
			logging.Err(err),
		)
		return nil, err
	}
	return &app.UserPreview{
		UserID:      user.GetUserId(),
		Username:    user.GetUsername(),
		DisplayName: user.GetDisplayName(),
		AvatarID:    user.GetAvatarId(),
		AvatarURL:   user.GetAvatarUrl(),
	}, nil
}

func (c *client) GetDetailedUserByUsername(ctx context.Context, username string) (*app.UserPreview, error) {
	res, err := c.cl.GetDetailedUserByUsername(ctx, &userv1.GetDetailedUserByUsernameRequest{Username: strings.TrimSpace(username)})
	if err != nil {
		c.log.Error("get detailed user by username failed",
			logging.Operation("grpc.user.detailed"),
			logging.String("username", strings.TrimSpace(username)),
			logging.Err(err),
		)
		return nil, err
	}
	user := res.GetUser()
	if user == nil {
		err := fmt.Errorf("%w: username=%s", ErrEmptyUserResponse, strings.TrimSpace(username))
		c.log.Error("get detailed user by username returned empty user",
			logging.Operation("grpc.user.detailed"),
			logging.String("username", strings.TrimSpace(username)),
			logging.Err(err),
		)
		return nil, err
	}
	return &app.UserPreview{
		UserID:      user.GetUserId(),
		Username:    user.GetUsername(),
		DisplayName: user.GetDisplayName(),
		AvatarID:    user.GetAvatarId(),
		AvatarURL:   user.GetAvatarUrl(),
	}, nil
}

func (c *client) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}
