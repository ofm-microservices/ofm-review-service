package service

import (
	"context"
	"errors"
	"strings"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"review-service/internal/application/cursordb"
	domain "review-service/internal/domain"
)

// ReviewListScope describes the key family used by the paginator.
type ReviewListScope string

const (
	// ReviewListScopeGig identifies gig review pagination.
	ReviewListScopeGig ReviewListScope = "gig"
	// ReviewListScopeSeller identifies seller review pagination.
	ReviewListScopeSeller ReviewListScope = "seller"
)

// ReviewListPaginator owns the cursor-driven windowed list behavior.
type ReviewListPaginator struct {
	repo      domain.ReviewRepository
	readRepo  domain.ReviewReadRepository
	cursor    CursorCodec
	projector ReviewProjector
	paging    PaginationConfig
	log       Logger
}

// ReviewProjector projects a review for the read path.
type ReviewProjector interface {
	ProjectReview(ctx context.Context, review *domain.Review, policy ProjectionPolicy) (*ProjectionResult, error)
}

// NewReviewListPaginator constructs the list paginator.
func NewReviewListPaginator(repo domain.ReviewRepository, readRepo domain.ReviewReadRepository, cursor CursorCodec, projector ReviewProjector, paging PaginationConfig, log Logger) (*ReviewListPaginator, error) {
	if repo == nil {
		return nil, ErrNilReviewRepository
	}
	if readRepo == nil {
		return nil, ErrNilReviewReadRepository
	}
	if cursor == nil {
		return nil, ErrNilCursorCodec
	}
	if paging.PageSize <= 0 || paging.WindowSize <= 0 || paging.WindowSize%paging.PageSize != 0 {
		return nil, ErrInvalidPaginationConfig
	}
	if log == nil {
		return nil, ErrNilLogger
	}
	return &ReviewListPaginator{
		repo:      repo,
		readRepo:  readRepo,
		cursor:    cursor,
		projector: projector,
		paging:    paging,
		log:       log.With(logging.String("module", "review-list-paginator")),
	}, nil
}

// ListGigReviews returns a gig-scoped list response.
func (p *ReviewListPaginator) ListGigReviews(ctx context.Context, gigID, rawCursor string) (*domain.ListReviewsResult, error) {
	if strings.TrimSpace(gigID) == "" {
		return nil, domain.ErrInvalidGigID
	}
	state, err := p.DecodeReviewCursor(rawCursor, string(ReviewListScopeGig))
	if err != nil {
		return nil, err
	}
	return p.List(ctx, listRequest{
		scope: string(ReviewListScopeGig),
		canonical: func(ctx context.Context, cursor string, limit int) (*domain.ListReviewsResult, error) {
			return p.repo.ListByGigID(ctx, domain.ListReviewsQuery{GigID: gigID, Cursor: cursor, Limit: limit})
		},
		window: func(ctx context.Context, window int) (*domain.ListReviewsResult, error) {
			return p.readRepo.ListByGigWindow(ctx, gigID, window)
		},
		windowUpsert: func(ctx context.Context, window int, reviews []*domain.Review, hasMore bool) error {
			ttl := p.paging.WindowTTL
			if window == 0 {
				ttl = 0
			}
			return p.readRepo.UpsertGigWindow(ctx, gigID, window, reviews, hasMore, ttl)
		},
		state: state,
	})
}

// ListSellerReviews returns a seller-scoped list response.
func (p *ReviewListPaginator) ListSellerReviews(ctx context.Context, sellerID, rawCursor string) (*domain.ListReviewsResult, error) {
	if strings.TrimSpace(sellerID) == "" {
		return nil, domain.ErrInvalidSellerID
	}
	state, err := p.DecodeReviewCursor(rawCursor, string(ReviewListScopeSeller))
	if err != nil {
		return nil, err
	}
	return p.List(ctx, listRequest{
		scope: string(ReviewListScopeSeller),
		canonical: func(ctx context.Context, cursor string, limit int) (*domain.ListReviewsResult, error) {
			return p.repo.ListBySellerID(ctx, domain.ListReviewsQuery{SellerID: sellerID, Cursor: cursor, Limit: limit})
		},
		window: func(ctx context.Context, window int) (*domain.ListReviewsResult, error) {
			return p.readRepo.ListBySellerWindow(ctx, sellerID, window)
		},
		windowUpsert: func(ctx context.Context, window int, reviews []*domain.Review, hasMore bool) error {
			ttl := p.paging.WindowTTL
			if window == 0 {
				ttl = 0
			}
			return p.readRepo.UpsertSellerWindow(ctx, sellerID, window, reviews, hasMore, ttl)
		},
		state: state,
	})
}

// List handles shared window pagination for gig and user review lists.
func (p *ReviewListPaginator) List(ctx context.Context, req listRequest) (*domain.ListReviewsResult, error) {
	pageSize := p.paging.PageSize
	windowSize := p.paging.WindowSize
	pagesPerWindow := windowSize / pageSize
	if pagesPerWindow <= 0 {
		pagesPerWindow = 1
	}
	if req.state.Page <= 0 {
		req.state.Page = 1
	}
	if req.state.Window < 0 {
		req.state.Window = 0
	}
	pageInWindow := ((req.state.Page - 1) % pagesPerWindow) + 1

	windowRes, err := req.window(ctx, req.state.Window)
	if err != nil && !errors.Is(err, domain.ErrReviewNotFound) {
		return nil, err
	}

	current := windowRes
	if current == nil || len(current.Reviews) == 0 {
		fetched, fetchErr := req.canonical(ctx, req.state.DBCursor(), windowSize)
		if fetchErr != nil {
			return nil, fetchErr
		}
		if fetched == nil || len(fetched.Reviews) == 0 {
			return &domain.ListReviewsResult{}, nil
		}
		fetchedReviews := make([]*domain.Review, 0, len(fetched.Reviews))
		for _, review := range fetched.Reviews {
			projected, err := p.ProjectReviewForList(ctx, review)
			if err != nil {
				return nil, err
			}
			if projected == nil || projected.Review == nil {
				continue
			}
			fetchedReviews = append(fetchedReviews, projected.Review)
		}
		if err := req.windowUpsert(ctx, req.state.Window, fetchedReviews, fetched.HasMore); err != nil {
			p.log.Error("review window cache upsert failed",
				logging.Operation("review.window.cache_upsert"),
				logging.String("scope", req.scope),
				logging.Int("window", req.state.Window),
				logging.Err(err),
			)
		}
		if fetched.HasMore && len(fetchedReviews) > 0 {
			nextWindowCursor := cursordb.New().Encode(fetchedReviews[len(fetchedReviews)-1].CreatedAt.UTC().UnixMicro(), fetchedReviews[len(fetchedReviews)-1].ID)
			nextFetched, nextFetchErr := req.canonical(ctx, nextWindowCursor, windowSize)
			if nextFetchErr == nil && nextFetched != nil && len(nextFetched.Reviews) > 0 {
				nextWindowReviews := make([]*domain.Review, 0, len(nextFetched.Reviews))
				for _, review := range nextFetched.Reviews {
					projected, err := p.ProjectReviewForList(ctx, review)
					if err != nil {
						return nil, err
					}
					if projected == nil || projected.Review == nil {
						continue
					}
					nextWindowReviews = append(nextWindowReviews, projected.Review)
				}
				if len(nextWindowReviews) > 0 {
					nextWindow := req.state.Window + 1
					if err := req.windowUpsert(ctx, nextWindow, nextWindowReviews, nextFetched.HasMore); err != nil {
						p.log.Error("review next window cache upsert failed",
							logging.Operation("review.window.next_cache_upsert"),
							logging.String("scope", req.scope),
							logging.Int("window", nextWindow),
							logging.Err(err),
						)
					}
				}
			}
		}
		current = &domain.ListReviewsResult{Reviews: fetchedReviews, HasMore: fetched.HasMore}
	}
	if current == nil || len(current.Reviews) == 0 {
		return &domain.ListReviewsResult{}, nil
	}

	pageStart := (pageInWindow - 1) * pageSize
	pageEnd := pageStart + pageSize
	if pageEnd > len(current.Reviews) {
		pageEnd = len(current.Reviews)
	}
	if pageStart >= len(current.Reviews) {
		return &domain.ListReviewsResult{}, nil
	}

	pageReviews := current.Reviews[pageStart:pageEnd]
	res := &domain.ListReviewsResult{Reviews: pageReviews, HasMore: current.HasMore || pageEnd < len(current.Reviews)}
	if len(pageReviews) == 0 {
		return res, nil
	}

	last := pageReviews[len(pageReviews)-1]
	if len(current.Reviews) == windowSize {
		nextCursor := cursordb.New().Encode(last.CreatedAt.UTC().UnixMicro(), last.ID)
		nextFetched, err := req.canonical(ctx, nextCursor, windowSize)
		if err != nil {
			return nil, err
		}
		nextWindowReviews := make([]*domain.Review, 0, len(nextFetched.Reviews))
		for _, review := range nextFetched.Reviews {
			projected, projectErr := p.ProjectReviewForList(ctx, review)
			if projectErr != nil {
				return nil, projectErr
			}
			if projected == nil || projected.Review == nil {
				continue
			}
			nextWindowReviews = append(nextWindowReviews, projected.Review)
		}
		if len(nextWindowReviews) > 0 {
			nextWindow := req.state.Window + 1
			if err := req.windowUpsert(ctx, nextWindow, nextWindowReviews, nextFetched.HasMore); err != nil {
				p.log.Error("review next window cache upsert failed",
					logging.Operation("review.window.next_cache_upsert"),
					logging.String("scope", req.scope),
					logging.Int("window", nextWindow),
					logging.Err(err),
				)
			}
			res.HasMore = true
		}
	}

	if res.HasMore {
		nextState := req.state.Next(last.CreatedAt.UTC().UnixMicro(), last.ID, pagesPerWindow)
		res.Cursor = p.EncodeReviewCursor(nextState)
	}

	return res, nil
}

func (p *ReviewListPaginator) DecodeReviewCursor(raw, scope string) (ReviewCursorState, error) {
	if strings.TrimSpace(raw) == "" {
		return ReviewCursorState{Scope: scope, Window: 0, Page: 1}, nil
	}
	var state ReviewCursorState
	if err := p.cursor.Decode(raw, &state); err != nil {
		return ReviewCursorState{}, err
	}
	if state.Scope != scope {
		return ReviewCursorState{}, domain.ErrInvalidReviewID
	}
	if state.Page <= 0 {
		state.Page = 1
	}
	return state, nil
}

func (p *ReviewListPaginator) EncodeReviewCursor(state ReviewCursorState) string {
	token, err := p.cursor.Encode(state)
	if err != nil {
		p.log.Error("encode review cursor failed",
			logging.Operation("review.cursor.encode"),
			logging.String("scope", state.Scope),
			logging.Int("window", state.Window),
			logging.Int("page", state.Page),
			logging.Err(err),
		)
		return ""
	}
	return token
}

func (p *ReviewListPaginator) ProjectReviewForList(ctx context.Context, review *domain.Review) (*ProjectionResult, error) {
	if p.projector == nil {
		return nil, ErrNilCursorCodec
	}
	return p.projector.ProjectReview(ctx, review, ProjectionPolicy{EnrichAuthor: true, EnrichAvatar: true})
}
