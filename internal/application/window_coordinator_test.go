package service

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "review-service/internal/domain"
)

type seedSourceStub struct {
	gigResult    *domain.ListReviewsResult
	sellerResult *domain.ListReviewsResult
	gigErr       error
	sellerErr    error
}

func (s seedSourceStub) Create(context.Context, domain.CreateReviewParams) (*domain.Review, error) {
	return nil, errors.New("unexpected call")
}
func (s seedSourceStub) GetByID(context.Context, string) (*domain.Review, error) {
	return nil, errors.New("unexpected call")
}
func (s seedSourceStub) GetByOrderID(context.Context, string) (*domain.Review, error) {
	return nil, errors.New("unexpected call")
}
func (s seedSourceStub) ListByGigID(context.Context, domain.ListReviewsQuery) (*domain.ListReviewsResult, error) {
	return s.gigResult, s.gigErr
}
func (s seedSourceStub) ListBySellerID(context.Context, domain.ListReviewsQuery) (*domain.ListReviewsResult, error) {
	return s.sellerResult, s.sellerErr
}
func (s seedSourceStub) GetGigRatingSummary(context.Context, string) (*domain.RatingSummary, error) {
	return nil, errors.New("unexpected call")
}
func (s seedSourceStub) GetSellerRatingSummary(context.Context, string) (*domain.RatingSummary, error) {
	return nil, errors.New("unexpected call")
}
func (s seedSourceStub) ListGigRatingSummaries(context.Context) ([]domain.GigRatingSummary, error) {
	return nil, errors.New("unexpected call")
}
func (s seedSourceStub) ListSellerRatingSummaries(context.Context) ([]domain.SellerRatingSummary, error) {
	return nil, errors.New("unexpected call")
}
func (s seedSourceStub) DeleteByID(context.Context, string) error {
	return errors.New("unexpected call")
}

type seedProjectorStub struct {
	projected *domain.Review
	err       error
	calls     int
}

func (p *seedProjectorStub) CreateReview(context.Context, CreateReviewCommand) (*domain.Review, error) {
	return nil, errors.New("unexpected call")
}
func (p *seedProjectorStub) ListGigReviews(context.Context, ListGigReviewsQuery) (*domain.ListReviewsResult, error) {
	return nil, errors.New("unexpected call")
}
func (p *seedProjectorStub) ListSellerReviews(context.Context, ListSellerReviewsQuery) (*domain.ListReviewsResult, error) {
	return nil, errors.New("unexpected call")
}
func (p *seedProjectorStub) GetGigRatingSummary(context.Context, string) (*domain.RatingSummary, error) {
	return nil, errors.New("unexpected call")
}
func (p *seedProjectorStub) GetUserRatingSummaryByUsername(context.Context, string) (*domain.RatingSummary, error) {
	return nil, errors.New("unexpected call")
}
func (p *seedProjectorStub) ProjectReview(context.Context, *domain.Review, ProjectionPolicy) (*ProjectionResult, error) {
	p.calls++
	if p.err != nil {
		return nil, p.err
	}
	return &ProjectionResult{Review: p.projected}, nil
}

type windowLogStub struct{ testLogger }

func TestParseOwnerFromQueueKey(t *testing.T) {
	if owner, ok := parseOwnerFromQueueKey("reviews.queue.gig.abc.def"); !ok || owner.kind != ownerKindGig || owner.id != "abc.def" {
		t.Fatalf("unexpected owner parse result: ok=%v owner=%#v", ok, owner)
	}
	if owner, ok := parseOwnerFromQueueKey("reviews.queue.seller.owner-1"); !ok || owner.kind != ownerKindSeller || owner.id != "owner-1" {
		t.Fatalf("unexpected seller owner parse result: ok=%v owner=%#v", ok, owner)
	}
	if _, ok := parseOwnerFromQueueKey("bad.key"); ok {
		t.Fatalf("expected invalid key to fail")
	}
}

func TestSeedOwnerWindowProjectsInitialWindow(t *testing.T) {
	review := &domain.Review{ID: "review-1", GigID: "gig-1", CreatedAt: time.UnixMicro(1000).UTC()}
	projected := &domain.Review{ID: review.ID, GigID: review.GigID, CreatedAt: review.CreatedAt, Author: &domain.ReviewAuthor{UserID: "buyer-1"}}
	coord := &reviewWindowCoordinator{
		source:    seedSourceStub{gigResult: &domain.ListReviewsResult{Reviews: []*domain.Review{review}}},
		projector: &seedProjectorStub{projected: projected},
		paging:    PaginationConfig{WindowSize: 10},
	}

	ops, seedID, err := coord.seedOwnerWindow(context.Background(), newGigOwner("gig-1"))
	if err != nil {
		t.Fatalf("seedOwnerWindow returned error: %v", err)
	}
	if seedID != review.ID {
		t.Fatalf("expected seed review id %q, got %q", review.ID, seedID)
	}
	if len(ops) != 1 || ops[0].Window != 0 || len(ops[0].Insert) != 1 || ops[0].Insert[0].Review.ID != review.ID {
		t.Fatalf("unexpected seed ops: %#v", ops)
	}
}

func TestSeedOwnerWindowReturnsNotFoundWhenProjectionDropsAllReviews(t *testing.T) {
	review := &domain.Review{ID: "review-1", SellerID: "seller-1", CreatedAt: time.UnixMicro(1000).UTC()}
	coord := &reviewWindowCoordinator{
		source:    seedSourceStub{sellerResult: &domain.ListReviewsResult{Reviews: []*domain.Review{review}}},
		projector: &seedProjectorStub{projected: nil},
		paging:    PaginationConfig{WindowSize: 10},
	}

	_, _, err := coord.seedOwnerWindow(context.Background(), newSellerOwner("seller-1"))
	if !errors.Is(err, domain.ErrReviewNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestNewReviewWindowCoordinatorValidatesDependencies(t *testing.T) {
	if _, err := NewReviewWindowCoordinator(nil, seedSourceStub{}, &seedProjectorStub{}, PaginationConfig{WindowSize: 1}, WindowCoordinatorConfig{LeaseTTL: time.Second, LeaseRenewInterval: time.Second, ReconcileInterval: time.Second}, testLogger{}); !errors.Is(err, ErrNilReviewReadRepository) {
		t.Fatalf("expected nil repo error, got %v", err)
	}
	if _, err := NewReviewWindowCoordinator(&projectionRepoStub{}, nil, &seedProjectorStub{}, PaginationConfig{WindowSize: 1}, WindowCoordinatorConfig{LeaseTTL: time.Second, LeaseRenewInterval: time.Second, ReconcileInterval: time.Second}, testLogger{}); !errors.Is(err, ErrNilReviewRepository) {
		t.Fatalf("expected nil source error, got %v", err)
	}
	if _, err := NewReviewWindowCoordinator(&projectionRepoStub{}, seedSourceStub{}, nil, PaginationConfig{WindowSize: 1}, WindowCoordinatorConfig{LeaseTTL: time.Second, LeaseRenewInterval: time.Second, ReconcileInterval: time.Second}, testLogger{}); !errors.Is(err, ErrNilReviewService) {
		t.Fatalf("expected nil projector error, got %v", err)
	}
	if _, err := NewReviewWindowCoordinator(&projectionRepoStub{}, seedSourceStub{}, &seedProjectorStub{}, PaginationConfig{WindowSize: 0}, WindowCoordinatorConfig{LeaseTTL: time.Second, LeaseRenewInterval: time.Second, ReconcileInterval: time.Second}, testLogger{}); !errors.Is(err, ErrInvalidPaginationConfig) {
		t.Fatalf("expected invalid pagination error, got %v", err)
	}
	if _, err := NewReviewWindowCoordinator(&projectionRepoStub{}, seedSourceStub{}, &seedProjectorStub{}, PaginationConfig{WindowSize: 1}, WindowCoordinatorConfig{}, testLogger{}); !errors.Is(err, ErrInvalidPaginationConfig) {
		t.Fatalf("expected invalid window coordinator config error, got %v", err)
	}
	if _, err := NewReviewWindowCoordinator(&projectionRepoStub{}, seedSourceStub{}, &seedProjectorStub{}, PaginationConfig{WindowSize: 1}, WindowCoordinatorConfig{LeaseTTL: time.Second, LeaseRenewInterval: time.Second, ReconcileInterval: time.Second}, nil); !errors.Is(err, ErrNilLogger) {
		t.Fatalf("expected nil logger error, got %v", err)
	}
}

type projectionRepoStub struct{}

func (projectionRepoStub) EnqueueWindowJob(context.Context, string, string, *domain.Review) (int64, error) {
	return 0, nil
}
func (projectionRepoStub) QueueLength(context.Context, string) (int64, error) { return 0, nil }
func (projectionRepoStub) PeekWindowJob(context.Context, string) (*domain.Review, error) {
	return nil, domain.ErrReviewNotFound
}
func (projectionRepoStub) PopWindowJob(context.Context, string) (*domain.Review, error) {
	return nil, domain.ErrReviewNotFound
}
func (projectionRepoStub) ActiveQueueKeys(context.Context, string) ([]string, error) {
	return nil, nil
}
func (projectionRepoStub) RemoveActiveQueueIfEmpty(context.Context, string, string) (bool, error) {
	return false, nil
}
func (projectionRepoStub) TryAcquireLease(context.Context, string, string, time.Duration) (bool, error) {
	return false, nil
}
func (projectionRepoStub) RenewLease(context.Context, string, string, time.Duration) (bool, error) {
	return false, nil
}
func (projectionRepoStub) ReleaseLease(context.Context, string, string) (bool, error) {
	return false, nil
}
func (projectionRepoStub) LoadWindowSnapshots(context.Context, string) ([]domain.WindowSnapshot, error) {
	return nil, nil
}
func (projectionRepoStub) ApplyWindowOpsAndPopJob(context.Context, string, string, string, string, string, []domain.WindowOp, time.Duration) (bool, error) {
	return false, nil
}
func (projectionRepoStub) CheckDurability(context.Context) (bool, error) { return true, nil }
