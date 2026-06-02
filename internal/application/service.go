package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"review-service/internal/application/cursordb"
	domain "review-service/internal/domain"
)

type reviewService struct {
	repo     domain.ReviewRepository
	readRepo domain.ReviewReadRepository
	orders   OrderLookupClient
	users    UserPreviewClient
	pub      ReviewPublisher
	cursor   CursorCodec
	paging   PaginationConfig
	log      Logger
}

// New constructs the review application service.
func New(repo domain.ReviewRepository, readRepo domain.ReviewReadRepository, orders OrderLookupClient, users UserPreviewClient, pub ReviewPublisher, codec CursorCodec, paging PaginationConfig, log Logger) (ReviewService, error) {
	if repo == nil {
		return nil, ErrNilReviewRepository
	}
	if readRepo == nil {
		return nil, ErrNilReviewReadRepository
	}
	if orders == nil {
		return nil, ErrNilOrderLookupClient
	}
	if users == nil {
		return nil, ErrNilUserPreviewClient
	}
	if pub == nil {
		return nil, ErrNilReviewPublisher
	}
	if codec == nil {
		return nil, ErrNilCursorCodec
	}
	if paging.PageSize <= 0 || paging.WindowSize <= 0 || paging.WindowSize%paging.PageSize != 0 {
		return nil, ErrInvalidPaginationConfig
	}
	if log == nil {
		return nil, ErrNilLogger
	}
	return &reviewService{
		repo:     repo,
		readRepo: readRepo,
		orders:   orders,
		users:    users,
		pub:      pub,
		cursor:   codec,
		paging:   paging,
		log:      log.With(logging.String("module", "application")),
	}, nil
}

func (s *reviewService) CreateReview(ctx context.Context, cmd CreateReviewCommand) (*domain.Review, error) {
	orderID := strings.TrimSpace(cmd.OrderID)
	buyerID := strings.TrimSpace(cmd.BuyerID)
	content := strings.TrimSpace(cmd.Content)
	if orderID == "" {
		return nil, domain.ErrInvalidOrderID
	}
	if buyerID == "" {
		return nil, domain.ErrInvalidBuyerID
	}
	if content == "" {
		return nil, domain.ErrInvalidContent
	}

	snap, err := s.orders.GetOrderLifecycleSnapshot(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if snap == nil || strings.TrimSpace(snap.OrderID) == "" {
		return nil, domain.ErrReviewNotFound
	}
	if strings.TrimSpace(snap.Status) != "completed" {
		return nil, domain.ErrReviewNotCompleted
	}
	if strings.TrimSpace(snap.BuyerID) != buyerID {
		return nil, domain.ErrReviewOwnerMismatch
	}

	if existing, err := s.repo.GetByOrderID(ctx, orderID); err == nil {
		return existing, nil
	} else if !errors.Is(err, domain.ErrReviewNotFound) {
		return nil, err
	}

	reviewID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	review, err := s.repo.Create(ctx, domain.CreateReviewParams{
		ID:       reviewID.String(),
		OrderID:  orderID,
		GigID:    snap.GigID,
		Content:  content,
		BuyerID:  buyerID,
		SellerID: snap.SellerID,
		Rating:   cmd.Rating,
	})
	if err != nil {
		return nil, err
	}
	review.SellerID = snap.SellerID
	review.Author = nil

	if err := s.pub.PublishGigReviewProjectionRequested(ctx, review); err != nil {
		s.log.Error("review gig projection request publish failed",
			logging.Operation("review.create"),
			logging.String("review_id", review.ID),
			logging.String("order_id", review.OrderID),
			logging.Err(err),
		)
	}
	if err := s.pub.PublishUserReviewProjectionRequested(ctx, review); err != nil {
		s.log.Error("review user projection request publish failed",
			logging.Operation("review.create"),
			logging.String("review_id", review.ID),
			logging.String("order_id", review.OrderID),
			logging.Err(err),
		)
	}
	if err := s.pub.PublishGigRatingRequested(ctx, review); err != nil {
		s.log.Error("review gig rating request publish failed",
			logging.Operation("review.create"),
			logging.String("review_id", review.ID),
			logging.String("order_id", review.OrderID),
			logging.Err(err),
		)
	}
	if err := s.pub.PublishSellerRatingRequested(ctx, review); err != nil {
		s.log.Error("review seller rating request publish failed",
			logging.Operation("review.create"),
			logging.String("review_id", review.ID),
			logging.String("order_id", review.OrderID),
			logging.Err(err),
		)
	}
	s.log.Info("review projection and rating requested events published",
		logging.Operation("review.create"),
		logging.String("review_id", review.ID),
		logging.String("order_id", review.OrderID),
	)
	return review, nil
}

func (s *reviewService) ListGigReviews(ctx context.Context, query ListGigReviewsQuery) (*domain.ListReviewsResult, error) {
	gigID := strings.TrimSpace(query.GigID)
	if gigID == "" {
		return nil, domain.ErrInvalidGigID
	}
	return s.ListGigReviewsByCursor(ctx, gigID, strings.TrimSpace(query.Cursor))
}

func (s *reviewService) ListSellerReviews(ctx context.Context, query ListSellerReviewsQuery) (*domain.ListReviewsResult, error) {
	sellerID := strings.TrimSpace(query.SellerID)
	if sellerID == "" {
		return nil, domain.ErrInvalidSellerID
	}
	return s.ListSellerReviewsByCursor(ctx, sellerID, strings.TrimSpace(query.Cursor))
}

func (s *reviewService) ProjectReview(ctx context.Context, review *domain.Review, policy ProjectionPolicy) (*ProjectionResult, error) {
	if review == nil {
		return nil, domain.ErrReviewNotFound
	}

	projected := *review
	result := &ProjectionResult{Review: &projected}

	if policy.EnrichAuthor && (projected.Author == nil || strings.TrimSpace(projected.Author.UserID) == "") {
		author, err := s.users.GetUserPreviewByIDNoCache(ctx, projected.BuyerID)
		if err != nil {
			result.AuthorRetry = true
			return result, nil
		}
		projected.Author = &domain.ReviewAuthor{
			UserID:      author.UserID,
			Username:    author.Username,
			DisplayName: author.DisplayName,
			AvatarID:    author.AvatarID,
			AvatarURL:   author.AvatarURL,
		}
	}

	if policy.EnrichAvatar && projected.Author != nil {
		projected.Author.AvatarURL = strings.TrimSpace(projected.Author.AvatarURL)
	}

	result.Review = &projected
	return result, nil
}

func (s *reviewService) projectReviewsForList(ctx context.Context, reviews []*domain.Review) ([]*domain.Review, error) {
	if len(reviews) == 0 {
		return nil, nil
	}

	projected := make([]*domain.Review, 0, len(reviews))
	for _, review := range reviews {
		projectedReview, err := s.ProjectReview(ctx, review, ProjectionPolicy{EnrichAuthor: true, EnrichAvatar: false})
		if err != nil {
			return nil, err
		}
		if projectedReview == nil || projectedReview.Review == nil {
			continue
		}
		projected = append(projected, projectedReview.Review)
	}

	return projected, nil
}

func (s *reviewService) ListGigReviewsByCursor(ctx context.Context, gigID, rawCursor string) (*domain.ListReviewsResult, error) {
	state, err := s.DecodeReviewCursor(rawCursor, "gig")
	if err != nil {
		return nil, err
	}
	return s.ListReviews(ctx, listRequest{
		scope: "gig",
		canonical: func(ctx context.Context, cursor string, limit int) (*domain.ListReviewsResult, error) {
			return s.repo.ListByGigID(ctx, domain.ListReviewsQuery{GigID: gigID, Cursor: cursor, Limit: limit})
		},
		window: func(ctx context.Context, window int) (*domain.ListReviewsResult, error) {
			return s.readRepo.ListByGigWindow(ctx, gigID, window)
		},
		windowUpsert: func(ctx context.Context, window int, reviews []*domain.Review, hasMore bool) error {
			ttl := s.paging.WindowTTL
			if window == 0 {
				ttl = 0
			}
			return s.readRepo.UpsertGigWindow(ctx, gigID, window, reviews, hasMore, ttl)
		},
		state: state,
	})
}

func (s *reviewService) ListSellerReviewsByCursor(ctx context.Context, sellerID, rawCursor string) (*domain.ListReviewsResult, error) {
	state, err := s.DecodeReviewCursor(rawCursor, "seller")
	if err != nil {
		return nil, err
	}
	return s.ListReviews(ctx, listRequest{
		scope: "seller",
		canonical: func(ctx context.Context, cursor string, limit int) (*domain.ListReviewsResult, error) {
			return s.repo.ListBySellerID(ctx, domain.ListReviewsQuery{SellerID: sellerID, Cursor: cursor, Limit: limit})
		},
		window: func(ctx context.Context, window int) (*domain.ListReviewsResult, error) {
			return s.readRepo.ListBySellerWindow(ctx, sellerID, window)
		},
		windowUpsert: func(ctx context.Context, window int, reviews []*domain.Review, hasMore bool) error {
			ttl := s.paging.WindowTTL
			if window == 0 {
				ttl = 0
			}
			return s.readRepo.UpsertSellerWindow(ctx, sellerID, window, reviews, hasMore, ttl)
		},
		state: state,
	})
}

type listRequest struct {
	scope        string
	canonical    func(context.Context, string, int) (*domain.ListReviewsResult, error)
	window       func(context.Context, int) (*domain.ListReviewsResult, error)
	windowUpsert func(context.Context, int, []*domain.Review, bool) error
	state        ReviewCursorState
}

func (s *reviewService) ListReviews(ctx context.Context, req listRequest) (*domain.ListReviewsResult, error) {
	pageSize := s.paging.PageSize
	windowSize := s.paging.WindowSize
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
		fetchedReviews, err := s.projectReviewsForList(ctx, fetched.Reviews)
		if err != nil {
			return nil, err
		}
		if err := req.windowUpsert(ctx, req.state.Window, fetchedReviews, fetched.HasMore); err != nil {
			s.log.Error("review window cache upsert failed",
				logging.Operation("review.window.cache_upsert"),
				logging.String("scope", req.scope),
				logging.Int("window", req.state.Window),
				logging.Err(err),
			)
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
		nextFetched, err := req.canonical(ctx, cursordb.New().Encode(last.CreatedAt.UTC().UnixMicro(), last.ID), windowSize)
		if err == nil && nextFetched != nil && len(nextFetched.Reviews) > 0 {
			nextPageReviews, projectErr := s.projectReviewsForList(ctx, nextFetched.Reviews)
			if projectErr != nil {
				return nil, projectErr
			}
			if len(nextPageReviews) > 0 {
				nextWindow := req.state.Window + 1
				if err := req.windowUpsert(ctx, nextWindow, nextPageReviews, nextFetched.HasMore); err != nil {
					s.log.Error("review next window cache upsert failed",
						logging.Operation("review.window.next_cache_upsert"),
						logging.String("scope", req.scope),
						logging.Int("window", nextWindow),
						logging.Err(err),
					)
				}
				res.HasMore = true
			}
		}
	}

	if res.HasMore {
		nextState := req.state.Next(last.CreatedAt.UTC().UnixMicro(), last.ID, pagesPerWindow)
		res.Cursor = s.EncodeReviewCursor(nextState)
	}

	return res, nil
}
