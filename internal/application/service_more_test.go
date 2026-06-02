package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	domain "review-service/internal/domain"
)

type testLogger struct{}

func (testLogger) Debug(string, ...logging.Field) {}
func (testLogger) Info(string, ...logging.Field)  {}
func (testLogger) Warn(string, ...logging.Field)  {}
func (testLogger) Error(string, ...logging.Field) {}
func (testLogger) With(...logging.Field) Logger   { return testLogger{} }
func (testLogger) Sync() error                    { return nil }

type createReviewOrderStub struct {
	snap *OrderLifecycleSnapshot
	err  error
}

func (s createReviewOrderStub) GetOrderLifecycleSnapshot(context.Context, string) (*OrderLifecycleSnapshot, error) {
	return s.snap, s.err
}

type createReviewRepoStub struct {
	existing      *domain.Review
	existingErr   error
	createReview  *domain.Review
	createErr     error
	createParams  []domain.CreateReviewParams
	orderIDChecks int
}

func (r *createReviewRepoStub) Create(ctx context.Context, params domain.CreateReviewParams) (*domain.Review, error) {
	r.createParams = append(r.createParams, params)
	return r.createReview, r.createErr
}

func (r *createReviewRepoStub) GetByID(context.Context, string) (*domain.Review, error) {
	return nil, errors.New("unexpected call")
}

func (r *createReviewRepoStub) GetByOrderID(ctx context.Context, orderID string) (*domain.Review, error) {
	r.orderIDChecks++
	return r.existing, r.existingErr
}

func (r *createReviewRepoStub) ListByGigID(context.Context, domain.ListReviewsQuery) (*domain.ListReviewsResult, error) {
	return nil, errors.New("unexpected call")
}

func (r *createReviewRepoStub) ListBySellerID(context.Context, domain.ListReviewsQuery) (*domain.ListReviewsResult, error) {
	return nil, errors.New("unexpected call")
}

func (r *createReviewRepoStub) GetGigRatingSummary(context.Context, string) (*domain.RatingSummary, error) {
	return nil, errors.New("unexpected call")
}

func (r *createReviewRepoStub) GetSellerRatingSummary(context.Context, string) (*domain.RatingSummary, error) {
	return nil, errors.New("unexpected call")
}

func (r *createReviewRepoStub) ListGigRatingSummaries(context.Context) ([]domain.GigRatingSummary, error) {
	return nil, errors.New("unexpected call")
}

func (r *createReviewRepoStub) ListSellerRatingSummaries(context.Context) ([]domain.SellerRatingSummary, error) {
	return nil, errors.New("unexpected call")
}

func (r *createReviewRepoStub) DeleteByID(context.Context, string) error {
	return errors.New("unexpected call")
}

type createReviewUserStub struct {
	user *UserPreview
	err  error
}

func (s createReviewUserStub) GetUserPreviewByID(context.Context, string) (*UserPreview, error) {
	return s.user, s.err
}

func (s createReviewUserStub) GetUserPreviewByIDNoCache(context.Context, string) (*UserPreview, error) {
	return s.user, s.err
}

func (s createReviewUserStub) GetDetailedUserByUsername(context.Context, string) (*UserPreview, error) {
	return s.user, s.err
}

type createReviewFileStub struct {
	url string
	err error
}

func (s createReviewFileStub) GetFileURL(context.Context, string) (string, error) {
	return s.url, s.err
}

func (s createReviewFileStub) GetFileURLs(context.Context, []string) (map[string]string, error) {
	return nil, s.err
}

type createReviewPublisherStub struct {
	gigCalls     int
	userCalls    int
	gigRating    int
	sellerRating int
	err          error
}

func (p *createReviewPublisherStub) PublishGigReviewProjectionRequested(context.Context, *domain.Review) error {
	p.gigCalls++
	return p.err
}

func (p *createReviewPublisherStub) PublishUserReviewProjectionRequested(context.Context, *domain.Review) error {
	p.userCalls++
	return p.err
}

func (p *createReviewPublisherStub) PublishGigRatingRequested(context.Context, *domain.Review) error {
	p.gigRating++
	return p.err
}

func (p *createReviewPublisherStub) PublishSellerRatingRequested(context.Context, *domain.Review) error {
	p.sellerRating++
	return p.err
}

func newReviewServiceForTest(repo domain.ReviewRepository, readRepo domain.ReviewReadRepository, orders OrderLookupClient, users UserPreviewClient, pub ReviewPublisher) *reviewService {
	return &reviewService{
		repo:     repo,
		readRepo: readRepo,
		orders:   orders,
		users:    users,
		pub:      pub,
		cursor:   fakeCursorCodec{},
		paging:   PaginationConfig{PageSize: 2, WindowSize: 4, WindowTTL: time.Minute},
		log:      testLogger{},
	}
}

func TestNewReviewServiceValidatesDependencies(t *testing.T) {
	ctxLogger := testLogger{}
	if _, err := New(nil, nil, nil, nil, nil, nil, PaginationConfig{PageSize: 1, WindowSize: 1}, ctxLogger); !errors.Is(err, ErrNilReviewRepository) {
		t.Fatalf("expected nil repo error, got %v", err)
	}
	if _, err := New(&createReviewRepoStub{}, nil, nil, nil, nil, nil, PaginationConfig{PageSize: 1, WindowSize: 1}, ctxLogger); !errors.Is(err, ErrNilReviewReadRepository) {
		t.Fatalf("expected nil read repo error, got %v", err)
	}
	if _, err := New(&createReviewRepoStub{}, &fakeReadRepo{}, nil, nil, nil, nil, PaginationConfig{PageSize: 1, WindowSize: 1}, ctxLogger); !errors.Is(err, ErrNilOrderLookupClient) {
		t.Fatalf("expected nil order client error, got %v", err)
	}
	if _, err := New(&createReviewRepoStub{}, &fakeReadRepo{}, createReviewOrderStub{}, nil, nil, nil, PaginationConfig{PageSize: 1, WindowSize: 1}, ctxLogger); !errors.Is(err, ErrNilUserPreviewClient) {
		t.Fatalf("expected nil user client error, got %v", err)
	}
	if _, err := New(&createReviewRepoStub{}, &fakeReadRepo{}, createReviewOrderStub{}, createReviewUserStub{}, nil, nil, PaginationConfig{PageSize: 1, WindowSize: 1}, ctxLogger); !errors.Is(err, ErrNilReviewPublisher) {
		t.Fatalf("expected nil publisher error, got %v", err)
	}
	if _, err := New(&createReviewRepoStub{}, &fakeReadRepo{}, createReviewOrderStub{}, createReviewUserStub{}, &createReviewPublisherStub{}, nil, PaginationConfig{PageSize: 1, WindowSize: 1}, ctxLogger); !errors.Is(err, ErrNilCursorCodec) {
		t.Fatalf("expected nil cursor codec error, got %v", err)
	}
	if _, err := New(&createReviewRepoStub{}, &fakeReadRepo{}, createReviewOrderStub{}, createReviewUserStub{}, &createReviewPublisherStub{}, fakeCursorCodec{}, PaginationConfig{PageSize: 0, WindowSize: 1}, ctxLogger); !errors.Is(err, ErrInvalidPaginationConfig) {
		t.Fatalf("expected invalid pagination error, got %v", err)
	}
	if _, err := New(&createReviewRepoStub{}, &fakeReadRepo{}, createReviewOrderStub{}, createReviewUserStub{}, &createReviewPublisherStub{}, fakeCursorCodec{}, PaginationConfig{PageSize: 1, WindowSize: 1}, nil); !errors.Is(err, ErrNilLogger) {
		t.Fatalf("expected nil logger error, got %v", err)
	}
}

func TestCreateReviewValidatesAndPublishes(t *testing.T) {
	pub := &createReviewPublisherStub{}
	repo := &createReviewRepoStub{
		createReview: &domain.Review{ID: "review-1", OrderID: "order-1", GigID: "gig-1", BuyerID: "buyer-1", Content: "content", Rating: 5},
		existingErr:  domain.ErrReviewNotFound,
	}
	svc := newReviewServiceForTest(repo, &fakeReadRepo{}, createReviewOrderStub{
		snap: &OrderLifecycleSnapshot{OrderID: "order-1", BuyerID: "buyer-1", GigID: "gig-1", SellerID: "seller-1", Status: "completed"},
	}, createReviewUserStub{}, pub)

	review, err := svc.CreateReview(context.Background(), CreateReviewCommand{OrderID: " order-1 ", BuyerID: " buyer-1 ", Content: " content ", Rating: 5})
	if err != nil {
		t.Fatalf("CreateReview returned error: %v", err)
	}
	if review == nil || review.ID != "review-1" {
		t.Fatalf("expected created review, got %#v", review)
	}
	if review.SellerID != "seller-1" {
		t.Fatalf("expected seller id to be copied from order snapshot, got %q", review.SellerID)
	}
	if review.Author != nil {
		t.Fatalf("expected author to be cleared on create, got %#v", review.Author)
	}
	if len(repo.createParams) != 1 {
		t.Fatalf("expected create to be called once, got %d", len(repo.createParams))
	}
	if pub.gigCalls != 1 || pub.userCalls != 1 || pub.gigRating != 1 || pub.sellerRating != 1 {
		t.Fatalf("expected all publish methods to be called once, got %#v", pub)
	}
}

func TestCreateReviewSkipsCreateWhenExistingReviewFound(t *testing.T) {
	existing := &domain.Review{ID: "review-existing", OrderID: "order-1"}
	repo := &createReviewRepoStub{
		existing: existing,
	}
	pub := &createReviewPublisherStub{}
	svc := newReviewServiceForTest(repo, &fakeReadRepo{}, createReviewOrderStub{
		snap: &OrderLifecycleSnapshot{OrderID: "order-1", BuyerID: "buyer-1", GigID: "gig-1", SellerID: "seller-1", Status: "completed"},
	}, createReviewUserStub{}, pub)

	review, err := svc.CreateReview(context.Background(), CreateReviewCommand{OrderID: "order-1", BuyerID: "buyer-1", Content: "content", Rating: 5})
	if err != nil {
		t.Fatalf("CreateReview returned error: %v", err)
	}
	if review != existing {
		t.Fatalf("expected existing review to be returned")
	}
	if len(repo.createParams) != 0 {
		t.Fatalf("expected create to be skipped, got %d calls", len(repo.createParams))
	}
	if pub.gigCalls != 0 || pub.userCalls != 0 || pub.gigRating != 0 || pub.sellerRating != 0 {
		t.Fatalf("expected no publish calls for duplicate review")
	}
}

func TestCreateReviewReturnsValidationAndLookupErrors(t *testing.T) {
	tests := []struct {
		name     string
		cmd      CreateReviewCommand
		order    createReviewOrderStub
		existing error
		wantErr  error
	}{
		{name: "invalid order", cmd: CreateReviewCommand{BuyerID: "buyer-1", Content: "x"}, wantErr: domain.ErrInvalidOrderID},
		{name: "invalid buyer", cmd: CreateReviewCommand{OrderID: "order-1", Content: "x"}, wantErr: domain.ErrInvalidBuyerID},
		{name: "invalid content", cmd: CreateReviewCommand{OrderID: "order-1", BuyerID: "buyer-1", Content: "   "}, wantErr: domain.ErrInvalidContent},
		{name: "not completed", cmd: CreateReviewCommand{OrderID: "order-1", BuyerID: "buyer-1", Content: "x"}, order: createReviewOrderStub{snap: &OrderLifecycleSnapshot{OrderID: "order-1", BuyerID: "buyer-1", GigID: "gig-1", SellerID: "seller-1", Status: "pending"}}, wantErr: domain.ErrReviewNotCompleted},
		{name: "owner mismatch", cmd: CreateReviewCommand{OrderID: "order-1", BuyerID: "buyer-1", Content: "x"}, order: createReviewOrderStub{snap: &OrderLifecycleSnapshot{OrderID: "order-1", BuyerID: "buyer-2", GigID: "gig-1", SellerID: "seller-1", Status: "completed"}}, wantErr: domain.ErrReviewOwnerMismatch},
		{name: "order lookup error", cmd: CreateReviewCommand{OrderID: "order-1", BuyerID: "buyer-1", Content: "x"}, order: createReviewOrderStub{err: errors.New("boom")}, wantErr: errors.New("boom")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &createReviewRepoStub{existingErr: domain.ErrReviewNotFound}
			svc := newReviewServiceForTest(repo, &fakeReadRepo{}, tc.order, createReviewUserStub{}, &createReviewPublisherStub{})
			_, err := svc.CreateReview(context.Background(), tc.cmd)
			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				return
			}
			if tc.name == "order lookup error" {
				if err == nil || err.Error() != tc.wantErr.Error() {
					t.Fatalf("expected %v, got %v", tc.wantErr, err)
				}
				return
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestProjectReviewEnrichesAuthorAndAvatar(t *testing.T) {
	svc := newReviewServiceForTest(&createReviewRepoStub{}, &fakeReadRepo{}, createReviewOrderStub{}, createReviewUserStub{
		user: &UserPreview{UserID: "buyer-1", Username: "user-1", DisplayName: "User One", AvatarID: "avatar-1", AvatarURL: "https://example.test/avatar.png"},
	}, &createReviewPublisherStub{})
	review := &domain.Review{ID: "review-1", BuyerID: "buyer-1"}

	projected, err := svc.ProjectReview(context.Background(), review, ProjectionPolicy{EnrichAuthor: true, EnrichAvatar: true})
	if err != nil {
		t.Fatalf("ProjectReview returned error: %v", err)
	}
	if projected == nil || projected.Review == nil {
		t.Fatalf("expected projected review")
	}
	if projected.AuthorRetry || projected.AvatarRetry {
		t.Fatalf("expected no retry flags, got %#v", projected)
	}
	if projected.Review.Author == nil {
		t.Fatalf("expected author to be enriched")
	}
	if projected.Review.Author.AvatarURL != "https://example.test/avatar.png" {
		t.Fatalf("expected avatar url to be enriched, got %q", projected.Review.Author.AvatarURL)
	}
}

func TestProjectReviewMarksRetryOnLookupFailures(t *testing.T) {
	svc := newReviewServiceForTest(&createReviewRepoStub{}, &fakeReadRepo{}, createReviewOrderStub{}, createReviewUserStub{err: errors.New("user down")}, &createReviewPublisherStub{})
	projected, err := svc.ProjectReview(context.Background(), &domain.Review{ID: "review-1", BuyerID: "buyer-1"}, ProjectionPolicy{EnrichAuthor: true, EnrichAvatar: true})
	if err != nil {
		t.Fatalf("ProjectReview returned error: %v", err)
	}
	if !projected.AuthorRetry {
		t.Fatalf("expected author retry flag")
	}
	if projected.Review.Author != nil {
		t.Fatalf("expected author to remain unset on retry")
	}

	svc = newReviewServiceForTest(&createReviewRepoStub{}, &fakeReadRepo{}, createReviewOrderStub{}, createReviewUserStub{
		user: &UserPreview{UserID: "buyer-1", Username: "user-1", DisplayName: "User One", AvatarID: "avatar-1", AvatarURL: "https://example.test/avatar.png"},
	}, &createReviewPublisherStub{})
	projected, err = svc.ProjectReview(context.Background(), &domain.Review{ID: "review-1", BuyerID: "buyer-1"}, ProjectionPolicy{EnrichAuthor: true, EnrichAvatar: true})
	if err != nil {
		t.Fatalf("ProjectReview returned error: %v", err)
	}
	if projected.AvatarRetry {
		t.Fatalf("expected avatar retry flag to remain false")
	}
	if projected.Review.Author == nil || projected.Review.Author.AvatarURL != "https://example.test/avatar.png" {
		t.Fatalf("expected avatar url from preview data to be preserved, got %#v", projected.Review.Author)
	}
}

func TestProjectReviewSkipsEnrichmentWhenAlreadyPresent(t *testing.T) {
	svc := newReviewServiceForTest(&createReviewRepoStub{}, &fakeReadRepo{}, createReviewOrderStub{}, createReviewUserStub{
		err: errors.New("should not be called"),
	}, &createReviewPublisherStub{})
	review := &domain.Review{
		ID:      "review-1",
		BuyerID: "buyer-1",
		Author: &domain.ReviewAuthor{
			UserID:      "buyer-1",
			Username:    "user-1",
			DisplayName: "User One",
			AvatarURL:   "https://example.test/avatar.png",
			AvatarID:    "avatar-1",
		},
	}

	projected, err := svc.ProjectReview(context.Background(), review, ProjectionPolicy{EnrichAuthor: true, EnrichAvatar: true})
	if err != nil {
		t.Fatalf("ProjectReview returned error: %v", err)
	}
	if projected.AuthorRetry || projected.AvatarRetry {
		t.Fatalf("expected no retries, got %#v", projected)
	}
	if projected.Review.Author == nil || projected.Review.Author.AvatarURL != "https://example.test/avatar.png" {
		t.Fatalf("expected existing author to be preserved, got %#v", projected.Review.Author)
	}
}

func TestProjectReviewReturnsNotFoundForNilReview(t *testing.T) {
	svc := newReviewServiceForTest(&createReviewRepoStub{}, &fakeReadRepo{}, createReviewOrderStub{}, createReviewUserStub{}, &createReviewPublisherStub{})
	if _, err := svc.ProjectReview(context.Background(), nil, ProjectionPolicy{}); !errors.Is(err, domain.ErrReviewNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestDecodeReviewCursorDefaultsAndScopeValidation(t *testing.T) {
	svc := newReviewServiceForTest(&createReviewRepoStub{}, &fakeReadRepo{}, createReviewOrderStub{}, createReviewUserStub{}, &createReviewPublisherStub{})

	state, err := svc.DecodeReviewCursor("", "gig")
	if err != nil {
		t.Fatalf("DecodeReviewCursor returned error: %v", err)
	}
	if state.Scope != "gig" || state.Window != 0 || state.Page != 1 {
		t.Fatalf("unexpected default state: %#v", state)
	}

	if _, err := svc.DecodeReviewCursor("broken", "gig"); err == nil {
		t.Fatalf("expected decode error")
	}

	if _, err := svc.DecodeReviewCursor("cursor-token", "seller"); !errors.Is(err, domain.ErrInvalidReviewID) {
		t.Fatalf("expected scope mismatch error, got %v", err)
	}
}

func TestReviewCursorStateNextHandlesZeroWindowSize(t *testing.T) {
	state := ReviewCursorState{Scope: "gig", Window: 0, Page: 1}
	next := state.Next(10, "review-1", 0)
	if next.Window != 1 || next.Page != 1 {
		t.Fatalf("expected zero pages per window to advance to next window, got %#v", next)
	}
}
