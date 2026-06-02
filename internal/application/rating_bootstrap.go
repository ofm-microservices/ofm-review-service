package service

import (
	"context"
	"strings"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	domain "review-service/internal/domain"
)

type ratingBootstrapper struct {
	repo     domain.ReviewRepository
	readRepo domain.ReviewReadRepository
	users    UserPreviewClient
	log      Logger
}

// NewRatingBootstrapper constructs the startup warmer that seeds Redis with
// every gig and seller rating summary before the NATS consumers start.
func NewRatingBootstrapper(repo domain.ReviewRepository, readRepo domain.ReviewReadRepository, users UserPreviewClient, log Logger) (RatingBootstrapper, error) {
	if repo == nil {
		return nil, ErrNilReviewRepository
	}
	if readRepo == nil {
		return nil, ErrNilReviewReadRepository
	}
	if users == nil {
		return nil, ErrNilUserPreviewClient
	}
	if log == nil {
		return nil, ErrNilLogger
	}

	return &ratingBootstrapper{
		repo:     repo,
		readRepo: readRepo,
		users:    users,
		log:      log.With(logging.String("module", "rating-bootstrapper")),
	}, nil
}

// Preload seeds Redis with the full gig and seller rating summaries from the
// canonical YugabyteDB read model.
func (b *ratingBootstrapper) Preload(ctx context.Context) error {
	gigSummaries, err := b.repo.ListGigRatingSummaries(ctx)
	if err != nil {
		return err
	}
	for _, summary := range gigSummaries {
		if err := b.readRepo.SetGigRating(ctx, summary.GigID, summary.RatingSummary); err != nil {
			b.log.Error("seed gig rating summary failed",
				logging.Operation("review.rating.bootstrap"),
				logging.String("gig_id", summary.GigID),
				logging.Err(err),
			)
			return err
		}
	}

	sellerSummaries, err := b.repo.ListSellerRatingSummaries(ctx)
	if err != nil {
		return err
	}
	for _, summary := range sellerSummaries {
		preview, err := b.users.GetUserPreviewByID(ctx, summary.SellerID)
		if err != nil {
			b.log.Error("resolve seller username failed",
				logging.Operation("review.rating.bootstrap"),
				logging.String("seller_id", summary.SellerID),
				logging.Err(err),
			)
			return err
		}
		if preview == nil {
			return domain.ErrInvalidUsername
		}
		username := strings.TrimSpace(preview.Username)
		if username == "" {
			return domain.ErrInvalidUsername
		}
		if err := b.readRepo.SetSellerRatingByUsername(ctx, username, summary.RatingSummary); err != nil {
			b.log.Error("seed seller rating summary failed",
				logging.Operation("review.rating.bootstrap"),
				logging.String("username", username),
				logging.Err(err),
			)
			return err
		}
	}

	b.log.Info("rating summaries seeded",
		logging.Operation("review.rating.bootstrap"),
		logging.Int("gig_summaries", len(gigSummaries)),
		logging.Int("seller_summaries", len(sellerSummaries)),
	)
	return nil
}
