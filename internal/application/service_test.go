package service

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	domain "review-service/internal/domain"
)

type fakeCursorCodec struct{}

func (fakeCursorCodec) Encode(v any) (string, error) { return "cursor-token", nil }
func (fakeCursorCodec) Decode(string, any) error     { return nil }

type fakeReviewRepo struct{}

func (fakeReviewRepo) Create(context.Context, domain.CreateReviewParams) (*domain.Review, error) {
	return nil, errors.New("unexpected call")
}
func (fakeReviewRepo) GetByID(context.Context, string) (*domain.Review, error) {
	return nil, errors.New("unexpected call")
}
func (fakeReviewRepo) GetByOrderID(context.Context, string) (*domain.Review, error) {
	return nil, errors.New("unexpected call")
}
func (fakeReviewRepo) ListByGigID(context.Context, domain.ListReviewsQuery) (*domain.ListReviewsResult, error) {
	return nil, errors.New("unexpected call")
}
func (fakeReviewRepo) ListBySellerID(context.Context, domain.ListReviewsQuery) (*domain.ListReviewsResult, error) {
	return nil, errors.New("unexpected call")
}
func (fakeReviewRepo) GetGigRatingSummary(context.Context, string) (*domain.RatingSummary, error) {
	return nil, errors.New("unexpected call")
}
func (fakeReviewRepo) GetSellerRatingSummary(context.Context, string) (*domain.RatingSummary, error) {
	return nil, errors.New("unexpected call")
}
func (fakeReviewRepo) ListGigRatingSummaries(context.Context) ([]domain.GigRatingSummary, error) {
	return nil, errors.New("unexpected call")
}
func (fakeReviewRepo) ListSellerRatingSummaries(context.Context) ([]domain.SellerRatingSummary, error) {
	return nil, errors.New("unexpected call")
}
func (fakeReviewRepo) DeleteByID(context.Context, string) error { return errors.New("unexpected call") }

type fakeReadRepo struct {
	upsertedWindows []int
}

func (r *fakeReadRepo) UpsertGigRating(context.Context, string, int32) error    { return nil }
func (r *fakeReadRepo) UpsertSellerRating(context.Context, string, int32) error { return nil }
func (r *fakeReadRepo) UpsertSellerRatingByUsername(context.Context, string, int32) error {
	return nil
}
func (r *fakeReadRepo) GetGigRatingSummary(context.Context, string) (*domain.RatingSummary, error) {
	return nil, errors.New("unexpected call")
}
func (r *fakeReadRepo) GetSellerRatingSummary(context.Context, string) (*domain.RatingSummary, error) {
	return nil, errors.New("unexpected call")
}
func (r *fakeReadRepo) GetSellerRatingSummaryByUsername(context.Context, string) (*domain.RatingSummary, error) {
	return nil, errors.New("unexpected call")
}
func (r *fakeReadRepo) SetGigRating(context.Context, string, domain.RatingSummary) error {
	return nil
}
func (r *fakeReadRepo) SetSellerRating(context.Context, string, domain.RatingSummary) error {
	return nil
}
func (r *fakeReadRepo) SetSellerRatingByUsername(context.Context, string, domain.RatingSummary) error {
	return nil
}
func (r *fakeReadRepo) ListByGigID(context.Context, domain.ListReviewsQuery) (*domain.ListReviewsResult, error) {
	return nil, errors.New("unexpected call")
}
func (r *fakeReadRepo) ListBySellerID(context.Context, domain.ListReviewsQuery) (*domain.ListReviewsResult, error) {
	return nil, errors.New("unexpected call")
}
func (r *fakeReadRepo) ListByGigWindow(context.Context, string, int) (*domain.ListReviewsResult, error) {
	return &domain.ListReviewsResult{
		Reviews: []*domain.Review{
			{
				ID:        "r1",
				GigID:     "gig-1",
				BuyerID:   "buyer-1",
				Content:   "first",
				Rating:    5,
				CreatedAt: time.UnixMicro(1000).UTC(),
				Author: &domain.ReviewAuthor{
					UserID:      "buyer-1",
					Username:    "user-1",
					DisplayName: "User One",
					AvatarURL:   "https://example.test/avatar.png",
				},
			},
			{
				ID:        "r2",
				GigID:     "gig-1",
				BuyerID:   "buyer-2",
				Content:   "second",
				Rating:    4,
				CreatedAt: time.UnixMicro(2000).UTC(),
				Author: &domain.ReviewAuthor{
					UserID:      "buyer-2",
					Username:    "user-2",
					DisplayName: "User Two",
					AvatarURL:   "https://example.test/avatar.png",
				},
			},
			{
				ID:        "r3",
				GigID:     "gig-1",
				BuyerID:   "buyer-3",
				Content:   "third",
				Rating:    3,
				CreatedAt: time.UnixMicro(3000).UTC(),
				Author: &domain.ReviewAuthor{
					UserID:      "buyer-3",
					Username:    "user-3",
					DisplayName: "User Three",
					AvatarURL:   "https://example.test/avatar.png",
				},
			},
			{
				ID:        "r4",
				GigID:     "gig-1",
				BuyerID:   "buyer-4",
				Content:   "fourth",
				Rating:    2,
				CreatedAt: time.UnixMicro(4000).UTC(),
				Author: &domain.ReviewAuthor{
					UserID:      "buyer-4",
					Username:    "user-4",
					DisplayName: "User Four",
					AvatarURL:   "https://example.test/avatar.png",
				},
			},
		},
		HasMore: false,
	}, nil
}
func (r *fakeReadRepo) ListBySellerWindow(context.Context, string, int) (*domain.ListReviewsResult, error) {
	return nil, errors.New("unexpected call")
}
func (r *fakeReadRepo) UpsertGigWindow(_ context.Context, _ string, window int, _ []*domain.Review, _ bool, _ time.Duration) error {
	r.upsertedWindows = append(r.upsertedWindows, window)
	return nil
}
func (r *fakeReadRepo) UpsertSellerWindow(context.Context, string, int, []*domain.Review, bool, time.Duration) error {
	return errors.New("unexpected call")
}
func (r *fakeReadRepo) DeleteByID(context.Context, string) error { return nil }

type fakeOrderClient struct{}

func (fakeOrderClient) GetOrderLifecycleSnapshot(context.Context, string) (*OrderLifecycleSnapshot, error) {
	return nil, errors.New("unexpected call")
}

type fakeUserClient struct{}

func (fakeUserClient) GetUserPreviewByID(context.Context, string) (*UserPreview, error) {
	return nil, errors.New("unexpected call")
}

func (fakeUserClient) GetUserPreviewByIDNoCache(context.Context, string) (*UserPreview, error) {
	return nil, errors.New("unexpected call")
}

func (fakeUserClient) GetDetailedUserByUsername(context.Context, string) (*UserPreview, error) {
	return nil, errors.New("unexpected call")
}

type fakeFileClient struct{}

func (fakeFileClient) GetFileURL(context.Context, string) (string, error) {
	return "", errors.New("unexpected call")
}

func (fakeFileClient) GetFileURLs(context.Context, []string) (map[string]string, error) {
	return nil, errors.New("unexpected call")
}

type fakePublisher struct{}

func (fakePublisher) PublishGigReviewProjectionRequested(context.Context, *domain.Review) error {
	return nil
}

func (fakePublisher) PublishUserReviewProjectionRequested(context.Context, *domain.Review) error {
	return nil
}

func (fakePublisher) PublishGigRatingRequested(context.Context, *domain.Review) error {
	return nil
}

func (fakePublisher) PublishSellerRatingRequested(context.Context, *domain.Review) error {
	return nil
}

func TestListReviewsPrefetchesNextWindowAtBoundary(t *testing.T) {
	ctx := context.Background()
	paging := PaginationConfig{PageSize: 1, WindowSize: 10, WindowTTL: time.Minute}
	windowSize := paging.WindowSize
	readRepo := &fakeReadRepo{}
	svc := &reviewService{
		repo:     fakeReviewRepo{},
		readRepo: readRepo,
		orders:   fakeOrderClient{},
		users:    fakeUserClient{},
		pub:      fakePublisher{},
		cursor:   fakeCursorCodec{},
		paging:   paging,
	}

	nextCalled := false
	windowReviews := make([]*domain.Review, 0, windowSize)
	for i := 1; i <= windowSize; i++ {
		windowReviews = append(windowReviews, &domain.Review{
			ID:        "r" + strconv.Itoa(i),
			GigID:     "gig-1",
			BuyerID:   "buyer-1",
			Content:   "review",
			Rating:    5,
			CreatedAt: time.UnixMicro(int64(i) * 1000).UTC(),
			Author: &domain.ReviewAuthor{
				UserID:      "buyer-1",
				Username:    "user-1",
				DisplayName: "User One",
				AvatarURL:   "https://example.test/avatar.png",
			},
		})
	}
	res, err := svc.ListReviews(ctx, listRequest{
		scope: "gig",
		canonical: func(context.Context, string, int) (*domain.ListReviewsResult, error) {
			nextCalled = true
			return &domain.ListReviewsResult{
				Reviews: []*domain.Review{
					{
						ID:        "r5",
						GigID:     "gig-1",
						BuyerID:   "buyer-5",
						Content:   "fifth",
						Rating:    5,
						CreatedAt: time.UnixMicro(5000).UTC(),
						Author: &domain.ReviewAuthor{
							UserID:      "buyer-5",
							Username:    "user-5",
							DisplayName: "User Five",
							AvatarURL:   "https://example.test/avatar.png",
						},
					},
				},
				HasMore: false,
			}, nil
		},
		window: func(context.Context, int) (*domain.ListReviewsResult, error) {
			return &domain.ListReviewsResult{Reviews: windowReviews, HasMore: true}, nil
		},
		windowUpsert: func(ctx context.Context, window int, reviews []*domain.Review, hasMore bool) error {
			return readRepo.UpsertGigWindow(ctx, "gig-1", window, reviews, hasMore, paging.WindowTTL)
		},
		state: ReviewCursorState{
			Scope:     "gig",
			Window:    0,
			Page:      paging.WindowSize,
			LastScore: int64(windowSize-1) * 1000,
			LastID:    "r" + strconv.Itoa(windowSize-1),
		},
	})
	if err != nil {
		t.Fatalf("listReviews returned error: %v", err)
	}
	if !nextCalled {
		t.Fatalf("expected canonical fetch for next window")
	}
	if len(readRepo.upsertedWindows) != 1 || readRepo.upsertedWindows[0] != 1 {
		t.Fatalf("expected next window 1 to be cached, got %v", readRepo.upsertedWindows)
	}
	if !res.HasMore {
		t.Fatalf("expected hasMore to be true")
	}
	if res.Cursor == "" {
		t.Fatalf("expected next cursor")
	}
}

func TestListReviewsDoesNotRebuildShortFinalWindow(t *testing.T) {
	ctx := context.Background()
	paging := PaginationConfig{PageSize: 10, WindowSize: 100, WindowTTL: time.Minute}
	readRepo := &fakeReadRepo{}
	svc := &reviewService{
		repo:     fakeReviewRepo{},
		readRepo: readRepo,
		orders:   fakeOrderClient{},
		users:    fakeUserClient{},
		pub:      fakePublisher{},
		cursor:   fakeCursorCodec{},
		paging:   paging,
	}

	windowReviews := []*domain.Review{
		{
			ID:        "r1",
			GigID:     "gig-1",
			BuyerID:   "buyer-1",
			Content:   "first",
			Rating:    5,
			CreatedAt: time.UnixMicro(1000).UTC(),
			Author: &domain.ReviewAuthor{
				UserID:      "buyer-1",
				Username:    "user-1",
				DisplayName: "User One",
				AvatarURL:   "https://example.test/avatar.png",
			},
		},
		{
			ID:        "r2",
			GigID:     "gig-1",
			BuyerID:   "buyer-2",
			Content:   "second",
			Rating:    4,
			CreatedAt: time.UnixMicro(2000).UTC(),
			Author: &domain.ReviewAuthor{
				UserID:      "buyer-2",
				Username:    "user-2",
				DisplayName: "User Two",
				AvatarURL:   "https://example.test/avatar.png",
			},
		},
	}

	nextCalled := false
	res, err := svc.ListReviews(ctx, listRequest{
		scope: "gig",
		canonical: func(context.Context, string, int) (*domain.ListReviewsResult, error) {
			nextCalled = true
			return nil, errors.New("unexpected canonical fetch")
		},
		window: func(context.Context, int) (*domain.ListReviewsResult, error) {
			return &domain.ListReviewsResult{Reviews: windowReviews, HasMore: false}, nil
		},
		windowUpsert: func(context.Context, int, []*domain.Review, bool) error {
			return errors.New("unexpected cache write")
		},
		state: ReviewCursorState{
			Scope:     "gig",
			Window:    0,
			Page:      1,
			LastScore: 1000,
			LastID:    "r1",
		},
	})
	if err != nil {
		t.Fatalf("listReviews returned error: %v", err)
	}
	if nextCalled {
		t.Fatalf("did not expect next-window canonical fetch for a final short window")
	}
	if len(res.Reviews) != len(windowReviews) {
		t.Fatalf("expected the cached short window to be returned as-is, got %d reviews", len(res.Reviews))
	}
}

func TestReviewCursorStateNextAndDBCursor(t *testing.T) {
	state := ReviewCursorState{Scope: "gig", Window: 0, Page: 1}
	next := state.Next(1234, "r1", 2)
	if next.Window != 0 || next.Page != 2 {
		t.Fatalf("expected page to advance within the same window, got window=%d page=%d", next.Window, next.Page)
	}

	lastPage := ReviewCursorState{Scope: "gig", Window: 0, Page: 2}.Next(5678, "r2", 2)
	if lastPage.Window != 1 || lastPage.Page != 1 {
		t.Fatalf("expected window to advance after the last page, got window=%d page=%d", lastPage.Window, lastPage.Page)
	}
	if got := lastPage.DBCursor(); got == "" {
		t.Fatalf("expected db cursor to be encoded")
	}
	if got := (ReviewCursorState{}).DBCursor(); got != "" {
		t.Fatalf("expected empty db cursor for zero state, got %q", got)
	}
}

func TestPlanWindowOpsCarriesToNextWindow(t *testing.T) {
	snapshots := []domain.WindowSnapshot{
		{
			Window: 0,
			Count:  2,
			Newest: domain.ReviewMember{Review: &domain.Review{ID: "new", CreatedAt: time.UnixMicro(3000).UTC()}},
			Oldest: domain.ReviewMember{Review: &domain.Review{ID: "old", CreatedAt: time.UnixMicro(1000).UTC()}},
		},
		{
			Window: 1,
			Count:  1,
			Newest: domain.ReviewMember{Review: &domain.Review{ID: "next-new", CreatedAt: time.UnixMicro(2000).UTC()}},
			Oldest: domain.ReviewMember{Review: &domain.Review{ID: "next-old", CreatedAt: time.UnixMicro(2000).UTC()}},
		},
	}
	review := &domain.Review{ID: "insert", CreatedAt: time.UnixMicro(3500).UTC()}

	ops, ok := planWindowOps(review, snapshots, 2)
	if !ok {
		t.Fatalf("expected window ops to be planned")
	}
	if len(ops) != 2 {
		t.Fatalf("expected a carry chain across two windows, got %#v", ops)
	}
	if ops[0].Window != 0 || ops[1].Window != 1 {
		t.Fatalf("expected windows 0 and 1, got %#v", ops)
	}
	if len(ops[0].Insert) != 1 || ops[0].Insert[0].Review.ID != "insert" {
		t.Fatalf("expected inserted review to be carried into window 0, got %#v", ops[0].Insert)
	}
	if len(ops[0].Evict) != 1 || ops[0].Evict[0].Review.ID != "old" {
		t.Fatalf("expected oldest review to be evicted from window 0, got %#v", ops[0].Evict)
	}
	if len(ops[1].Insert) != 1 || ops[1].Insert[0].Review.ID != "old" {
		t.Fatalf("expected evicted review to carry into window 1, got %#v", ops[1].Insert)
	}
}
