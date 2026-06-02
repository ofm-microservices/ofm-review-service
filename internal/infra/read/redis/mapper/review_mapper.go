package mapper

import (
	"time"

	domain "review-service/internal/domain"
	"review-service/internal/infra/read/redis/model"
)

// MapDomainReviewToCache maps the domain review into its Redis projection model.
func MapDomainReviewToCache(review *domain.Review) model.ReviewCache {
	out := model.ReviewCache{
		ID:        review.ID,
		GigID:     review.GigID,
		Content:   review.Content,
		BuyerID:   review.BuyerID,
		Rating:    review.Rating,
		CreatedAt: review.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
	if review.Author != nil {
		out.Author = &model.Author{
			UserID:      review.Author.UserID,
			Username:    review.Author.Username,
			DisplayName: review.Author.DisplayName,
			AvatarURL:   review.Author.AvatarURL,
			AvatarID:    review.Author.AvatarID,
		}
	}
	return out
}
