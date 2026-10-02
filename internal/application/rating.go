package service

import (
	"context"
	"errors"
	"strings"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	domain "review-service/internal/domain"
)

// GetGigRatingSummary returns the gig rating summary from Redis or loads it
// from the canonical PostgreSQL read if the cache is empty.
func (s *reviewService) GetGigRatingSummary(ctx context.Context, gigID string) (*domain.RatingSummary, error) {
	gigID = strings.TrimSpace(gigID)
	if gigID == "" {
		return nil, domain.ErrInvalidGigID
	}

	summary, err := s.readRepo.GetGigRatingSummary(ctx, gigID)
	if err == nil {
		return summary, nil
	}
	if !errors.Is(err, domain.ErrReviewNotFound) {
		return nil, err
	}

	summary, err = s.repo.GetGigRatingSummary(ctx, gigID)
	if err != nil {
		return nil, err
	}
	if err := s.readRepo.SetGigRating(ctx, gigID, *summary); err != nil {
		s.log.Error("gig rating cache seed failed",
			logging.Operation("review.rating.cache_seed"),
			logging.String("gig_id", gigID),
			logging.Err(err),
		)
	}
	return summary, nil
}

// GetUserRatingSummaryByUsername returns the user rating summary from Redis or
// resolves the seller's canonical identifier from user-service before
// falling back to PostgreSQL.
func (s *reviewService) GetUserRatingSummaryByUsername(ctx context.Context, username string) (*domain.RatingSummary, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, domain.ErrInvalidUsername
	}

	summary, err := s.readRepo.GetSellerRatingSummaryByUsername(ctx, username)
	if err == nil {
		return summary, nil
	}
	if !errors.Is(err, domain.ErrReviewNotFound) {
		return nil, err
	}

	details, err := s.users.GetDetailedUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if details == nil {
		return nil, domain.ErrReviewNotFound
	}
	sellerID := strings.TrimSpace(details.UserID)
	if sellerID == "" {
		return nil, domain.ErrInvalidUsername
	}
	if err := s.readRepo.SetSellerIDByUsername(ctx, username, sellerID); err != nil {
		s.log.Error("seller username id cache seed failed",
			logging.Operation("review.rating.cache_seed"),
			logging.String("username", username),
			logging.String("seller_id", sellerID),
			logging.Err(err),
		)
	}

	summary, err = s.repo.GetSellerRatingSummary(ctx, sellerID)
	if err != nil {
		return nil, err
	}
	if err := s.readRepo.SetSellerRatingByUsername(ctx, username, *summary); err != nil {
		s.log.Error("seller rating username cache seed failed",
			logging.Operation("review.rating.cache_seed"),
			logging.String("username", username),
			logging.Err(err),
		)
	}
	return summary, nil
}
