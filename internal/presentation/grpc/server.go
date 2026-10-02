package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"review-service/config"
	app "review-service/internal/application"
	domain "review-service/internal/domain"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	transportgrpc "github.com/ofm-microservices/ofm-common/pkg/observability/grpc"
	"github.com/ofm-microservices/ofm-common/pkg/observability/metrics"
	reviewv1 "github.com/ofm-microservices/ofm-common/proto/review/v1"
	otelgrpc "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	reviewv1.UnimplementedReviewServiceServer
	svc      app.ReviewService
	mapper   ReviewMapper
	cfg      config.GRPCConfig
	log      logging.Logger
	srv      *grpc.Server
	listener net.Listener
}

// NewServer constructs the review-service gRPC server.
func NewServer(svc app.ReviewService, mapper ReviewMapper, cfg config.GRPCConfig, log logging.Logger) (Server, error) {
	if svc == nil {
		return nil, ErrNilReviewService
	}
	if mapper == nil {
		return nil, ErrNilReviewMapper
	}
	if log == nil {
		return nil, ErrNilLogger
	}
	grpcSrv := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(metrics.UnaryServerInterceptor(), transportgrpc.UnaryServerInterceptor(log)),
	)
	s := &server{
		svc:    svc,
		mapper: mapper,
		cfg:    cfg,
		log:    log.With(logging.String("module", "grpc-server")),
		srv:    grpcSrv,
	}
	reviewv1.RegisterReviewServiceServer(grpcSrv, s)
	return s, nil
}

func (s *server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = lis
	s.log.Info("starting grpc server", logging.String("addr", addr))
	return s.srv.Serve(lis)
}

func (s *server) Shutdown(context.Context) error {
	if s.srv != nil {
		s.log.Info("shutting down grpc server")
		s.srv.GracefulStop()
	}
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *server) CreateReview(ctx context.Context, req *reviewv1.CreateReviewRequest) (*reviewv1.CreateReviewResponse, error) {
	started := time.Now()
	log := logging.WithContext(ctx, s.log)
	res, err := s.svc.CreateReview(ctx, app.CreateReviewCommand{
		OrderID:        req.GetOrderId(),
		BuyerID:        req.GetBuyerUserId(),
		BuyerUsername:  req.GetBuyerUsername(),
		Content:        req.GetContent(),
		Rating:         req.GetRating(),
		IdempotencyKey: req.GetIdempotencyKey(),
		RequestedAt:    req.GetRequestedAt(),
	})
	if err != nil {
		log.Warn("create review returned a business condition",
			logging.Operation("grpc.review.create"),
			logging.Attempt(1),
			logging.Retryable(false),
			logging.DurationMS(time.Since(started)),
			logging.String("order_id", req.GetOrderId()),
			logging.String("buyer_id", req.GetBuyerUserId()),
			logging.String("buyer_username", req.GetBuyerUsername()),
			logging.Err(err),
		)
		switch {
		case errors.Is(err, domain.ErrInvalidOrderID),
			errors.Is(err, domain.ErrInvalidBuyerID),
			errors.Is(err, domain.ErrInvalidContent):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, domain.ErrReviewNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		case errors.Is(err, domain.ErrReviewOwnerMismatch):
			return nil, status.Error(codes.PermissionDenied, err.Error())
		case errors.Is(err, domain.ErrReviewNotCompleted):
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		default:
			log.Error("create review failed internally", logging.Err(err))
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}
	return &reviewv1.CreateReviewResponse{
		Review: s.mapper.ToProto(res),
	}, nil
}

func (s *server) ListGigReviews(ctx context.Context, req *reviewv1.ListGigReviewsRequest) (*reviewv1.ListGigReviewsResponse, error) {
	started := time.Now()
	log := logging.WithContext(ctx, s.log)
	res, err := s.svc.ListGigReviews(ctx, app.ListGigReviewsQuery{
		GigID:  req.GetGigId(),
		Cursor: req.GetCursor(),
	})
	if err != nil {
		log.Warn("list gig reviews returned a business condition",
			logging.Operation("grpc.review.list_gig"),
			logging.Attempt(1),
			logging.Retryable(false),
			logging.DurationMS(time.Since(started)),
			logging.String("gig_id", req.GetGigId()),
			logging.Err(err),
		)
		if errors.Is(err, domain.ErrInvalidGigID) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		log.Error("list gig reviews failed internally", logging.Err(err))
		return nil, status.Error(codes.Internal, "internal server error")
	}

	reviews := make([]*reviewv1.Review, 0, len(res.Reviews))
	for _, review := range res.Reviews {
		reviews = append(reviews, s.mapper.ToProto(review))
	}
	return &reviewv1.ListGigReviewsResponse{
		Reviews: reviews,
		Cursor:  res.Cursor,
		HasMore: res.HasMore,
	}, nil
}

func (s *server) ListSellerReviews(ctx context.Context, req *reviewv1.ListSellerReviewsRequest) (*reviewv1.ListSellerReviewsResponse, error) {
	started := time.Now()
	log := logging.WithContext(ctx, s.log)
	res, err := s.svc.ListSellerReviews(ctx, app.ListSellerReviewsQuery{
		SellerID: req.GetSellerId(),
		Cursor:   req.GetCursor(),
	})
	if err != nil {
		log.Warn("list seller reviews returned a business condition",
			logging.Operation("grpc.review.list_seller"),
			logging.Attempt(1),
			logging.Retryable(false),
			logging.DurationMS(time.Since(started)),
			logging.String("seller_id", req.GetSellerId()),
			logging.Err(err),
		)
		if errors.Is(err, domain.ErrInvalidSellerID) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		log.Error("list seller reviews failed internally", logging.Err(err))
		return nil, status.Error(codes.Internal, "internal server error")
	}

	reviews := make([]*reviewv1.Review, 0, len(res.Reviews))
	for _, review := range res.Reviews {
		reviews = append(reviews, s.mapper.ToProto(review))
	}
	return &reviewv1.ListSellerReviewsResponse{
		Reviews: reviews,
		Cursor:  res.Cursor,
		HasMore: res.HasMore,
	}, nil
}

func (s *server) GetReviewsBySellerUsername(ctx context.Context, req *reviewv1.GetReviewsBySellerUsernameRequest) (*reviewv1.ListSellerReviewsResponse, error) {
	started := time.Now()
	log := logging.WithContext(ctx, s.log)
	res, err := s.svc.GetReviewsBySellerUsername(ctx, req.GetUsername(), req.GetCursor())
	if err != nil {
		if errors.Is(err, domain.ErrInvalidUsername) || errors.Is(err, domain.ErrReviewNotFound) {
			log.Warn("get reviews by seller username returned a business condition",
				logging.Operation("grpc.review.get_reviews_by_seller_username"),
				logging.Attempt(1),
				logging.Retryable(false),
				logging.DurationMS(time.Since(started)),
				logging.String("username", req.GetUsername()),
				logging.Err(err),
			)
			if errors.Is(err, domain.ErrReviewNotFound) {
				return nil, status.Error(codes.NotFound, err.Error())
			}
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		log.Error("get reviews by seller username failed internally", logging.Err(err))
		return nil, status.Error(codes.Internal, "internal server error")
	}

	reviews := make([]*reviewv1.Review, 0, len(res.Reviews))
	for _, review := range res.Reviews {
		reviews = append(reviews, s.mapper.ToProto(review))
	}
	return &reviewv1.ListSellerReviewsResponse{
		Reviews: reviews,
		Cursor:  res.Cursor,
		HasMore: res.HasMore,
	}, nil
}

func (s *server) GetGigRatingSummary(ctx context.Context, req *reviewv1.GetGigRatingSummaryRequest) (*reviewv1.RatingSummary, error) {
	started := time.Now()
	log := logging.WithContext(ctx, s.log)
	res, err := s.svc.GetGigRatingSummary(ctx, req.GetGigId())
	if err != nil {
		if errors.Is(err, domain.ErrInvalidGigID) {
			log.Warn("get gig rating summary returned a business condition",
				logging.Operation("grpc.review.get_gig_rating_summary"),
				logging.Attempt(1),
				logging.Retryable(false),
				logging.DurationMS(time.Since(started)),
				logging.String("gig_id", req.GetGigId()),
				logging.Err(err),
			)
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, domain.ErrReviewNotFound) {
			log.Warn("get gig rating summary returned a business condition",
				logging.Operation("grpc.review.get_gig_rating_summary"),
				logging.Attempt(1),
				logging.Retryable(false),
				logging.DurationMS(time.Since(started)),
				logging.String("gig_id", req.GetGigId()),
				logging.Err(err),
			)
			return nil, status.Error(codes.NotFound, err.Error())
		}
		log.Error("get gig rating summary failed internally", logging.Err(err))
		return nil, status.Error(codes.Internal, "internal server error")
	}
	return s.mapper.ToRatingSummaryProto(res), nil
}

func (s *server) GetUserRatingSummaryByUsername(ctx context.Context, req *reviewv1.GetUserRatingSummaryByUsernameRequest) (*reviewv1.RatingSummary, error) {
	started := time.Now()
	log := logging.WithContext(ctx, s.log)
	res, err := s.svc.GetUserRatingSummaryByUsername(ctx, req.GetUsername())
	if err != nil {
		if errors.Is(err, domain.ErrInvalidUsername) || errors.Is(err, domain.ErrReviewNotFound) {
			log.Warn("user rating summary returned a business condition",
				logging.Operation("grpc.review.get_user_rating_summary_by_username"),
				logging.String("username", req.GetUsername()),
				logging.Err(err),
			)
		} else {
			log.Error("get user rating summary by username failed",
				logging.Operation("grpc.review.get_user_rating_summary_by_username"),
				logging.Attempt(1),
				logging.Retryable(false),
				logging.DurationMS(time.Since(started)),
				logging.String("username", req.GetUsername()),
				logging.Err(err),
			)
		}
		if errors.Is(err, domain.ErrInvalidUsername) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, domain.ErrReviewNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}
	return s.mapper.ToRatingSummaryProto(res), nil
}
