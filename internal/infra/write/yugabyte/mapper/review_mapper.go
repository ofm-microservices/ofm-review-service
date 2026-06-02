package mapper

import (
	domain "review-service/internal/domain"
	"review-service/internal/infra/write/yugabyte/model"
	"review-service/internal/infra/write/yugabyte/nullstring"
)

// MapReviewRowToDomain maps the Yugabyte row to the review domain entity.
func MapReviewRowToDomain(row model.ReviewRow) *domain.Review {
	conv := nullstring.New()
	return &domain.Review{
		ID:        row.ID,
		OrderID:   row.OrderID,
		GigID:     row.GigID,
		Content:   row.Content,
		BuyerID:   row.BuyerID,
		SellerID:  conv.Convert(row.SellerID),
		Rating:    row.Rating,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
