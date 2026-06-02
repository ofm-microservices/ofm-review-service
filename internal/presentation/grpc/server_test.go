package grpc

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"review-service/config"
	app "review-service/internal/application"
	domain "review-service/internal/domain"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	reviewv1 "github.com/ofm-microservices/ofm-common/proto/review/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverSvcStub struct {
	createReview  *domain.Review
	createErr     error
	listReviews   *domain.ListReviewsResult
	listErr       error
	gigSummary    *domain.RatingSummary
	sellerSummary *domain.RatingSummary
}

func (s *serverSvcStub) CreateReview(context.Context, app.CreateReviewCommand) (*domain.Review, error) {
	return s.createReview, s.createErr
}
func (s *serverSvcStub) ListGigReviews(context.Context, app.ListGigReviewsQuery) (*domain.ListReviewsResult, error) {
	return s.listReviews, s.listErr
}
func (s *serverSvcStub) ListSellerReviews(context.Context, app.ListSellerReviewsQuery) (*domain.ListReviewsResult, error) {
	return s.listReviews, s.listErr
}
func (s *serverSvcStub) GetGigRatingSummary(context.Context, string) (*domain.RatingSummary, error) {
	return s.gigSummary, s.createErr
}
func (s *serverSvcStub) GetUserRatingSummaryByUsername(context.Context, string) (*domain.RatingSummary, error) {
	return s.sellerSummary, s.createErr
}
func (s *serverSvcStub) ProjectReview(context.Context, *domain.Review, app.ProjectionPolicy) (*app.ProjectionResult, error) {
	return nil, errors.New("unexpected call")
}

type grpcTestLogger struct{}

func (grpcTestLogger) Debug(string, ...logging.Field)       {}
func (grpcTestLogger) Info(string, ...logging.Field)        {}
func (grpcTestLogger) Warn(string, ...logging.Field)        {}
func (grpcTestLogger) Error(string, ...logging.Field)       {}
func (grpcTestLogger) With(...logging.Field) logging.Logger { return grpcTestLogger{} }
func (grpcTestLogger) Sync() error                          { return nil }

func TestNewServerValidatesDependencies(t *testing.T) {
	if _, err := NewServer(nil, NewReviewMapper(), config.GRPCConfig{}, grpcTestLogger{}); !errors.Is(err, ErrNilReviewService) {
		t.Fatalf("expected nil service error, got %v", err)
	}
	if _, err := NewServer(&serverSvcStub{}, nil, config.GRPCConfig{}, grpcTestLogger{}); !errors.Is(err, ErrNilReviewMapper) {
		t.Fatalf("expected nil mapper error, got %v", err)
	}
	if _, err := NewServer(&serverSvcStub{}, NewReviewMapper(), config.GRPCConfig{}, nil); !errors.Is(err, ErrNilLogger) {
		t.Fatalf("expected nil logger error, got %v", err)
	}
}

func TestServerHandlers(t *testing.T) {
	svc := &serverSvcStub{
		createReview: &domain.Review{
			ID:        "review-1",
			OrderID:   "order-1",
			GigID:     "gig-1",
			BuyerID:   "buyer-1",
			Content:   "content",
			Rating:    5,
			CreatedAt: time.UnixMicro(1234).UTC(),
		},
		listReviews: &domain.ListReviewsResult{
			Reviews: []*domain.Review{{ID: "review-1", GigID: "gig-1", BuyerID: "buyer-1", Rating: 5, CreatedAt: time.UnixMicro(1234).UTC()}},
			Cursor:  "cursor",
			HasMore: true,
		},
		gigSummary:    &domain.RatingSummary{RatingAvg: 5, TotalReviews: 1},
		sellerSummary: &domain.RatingSummary{RatingAvg: 4, TotalReviews: 2},
	}
	srv, err := NewServer(svc, NewReviewMapper(), config.GRPCConfig{Host: "127.0.0.1", Port: 0}, grpcTestLogger{})
	if err != nil {
		t.Fatalf("NewServer returned error: %v", err)
	}
	s := srv.(*server)

	createRes, err := s.CreateReview(context.Background(), &reviewv1.CreateReviewRequest{OrderId: "order-1", BuyerUserId: "buyer-1", Content: "content", Rating: 5})
	if err != nil || createRes.GetReview().GetReviewId() != "review-1" {
		t.Fatalf("unexpected create response: res=%#v err=%v", createRes, err)
	}

	listRes, err := s.ListGigReviews(context.Background(), &reviewv1.ListGigReviewsRequest{GigId: "gig-1"})
	if err != nil || len(listRes.GetReviews()) != 1 || listRes.GetCursor() != "cursor" || !listRes.GetHasMore() {
		t.Fatalf("unexpected list response: res=%#v err=%v", listRes, err)
	}

	ratingRes, err := s.GetGigRatingSummary(context.Background(), &reviewv1.GetGigRatingSummaryRequest{GigId: "gig-1"})
	if err != nil || ratingRes.GetRatingAvg() != 5 {
		t.Fatalf("unexpected gig rating response: res=%#v err=%v", ratingRes, err)
	}

	sellerRatingRes, err := s.GetUserRatingSummaryByUsername(context.Background(), &reviewv1.GetUserRatingSummaryByUsernameRequest{Username: "seller-1"})
	if err != nil || sellerRatingRes.GetRatingAvg() != 4 {
		t.Fatalf("unexpected seller rating response: res=%#v err=%v", sellerRatingRes, err)
	}
}

func TestServerShutdownWithoutListener(t *testing.T) {
	s := &server{}
	if err := s.Shutdown(context.Background()); err != nil {
		t.Fatalf("expected nil shutdown error, got %v", err)
	}
}

func TestServerStartListenerError(t *testing.T) {
	s := &server{cfg: config.GRPCConfig{Host: "127.0.0.1", Port: -1}, log: grpcTestLogger{}, srv: nil}
	if err := s.Start(); err == nil {
		t.Fatalf("expected listen error")
	}
	_ = net.IPv4len
}

func TestServerMapsErrorsToStatusCodes(t *testing.T) {
	s := &server{
		svc:    &serverSvcStub{createErr: domain.ErrReviewNotCompleted, listErr: domain.ErrInvalidGigID},
		mapper: NewReviewMapper(),
		log:    grpcTestLogger{},
	}

	if _, err := s.CreateReview(context.Background(), &reviewv1.CreateReviewRequest{OrderId: "order-1", BuyerUserId: "buyer-1", Content: "content"}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("expected failed precondition, got %v", err)
	}

	if _, err := s.ListGigReviews(context.Background(), &reviewv1.ListGigReviewsRequest{GigId: "gig-1"}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected invalid argument, got %v", err)
	}

	if _, err := s.ListSellerReviews(context.Background(), &reviewv1.ListSellerReviewsRequest{SellerId: "seller-1"}); status.Code(err) != codes.Internal {
		t.Fatalf("expected internal error for stubbed seller listing, got %v", err)
	}
}
