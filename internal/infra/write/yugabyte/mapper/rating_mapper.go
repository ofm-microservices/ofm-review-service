package mapper

import (
	domain "review-service/internal/domain"
	"review-service/internal/infra/write/yugabyte/model"
)

// MapRatingSummaryRowToDomain maps a grouped rating row into the shared
// domain summary shape.
func MapRatingSummaryRowToDomain(row model.RatingSummaryRow) domain.RatingSummary {
	return domain.RatingSummary{
		RatingAvg:    row.RatingAvg,
		TotalReviews: row.TotalReviews,
		Stars5:       row.Stars5,
		Stars4:       row.Stars4,
		Stars3:       row.Stars3,
		Stars2:       row.Stars2,
		Stars1:       row.Stars1,
	}
}

// MapGigRatingSummaryRowsToDomain maps grouped gig rating rows into domain
// summaries.
func MapGigRatingSummaryRowsToDomain(rows []model.RatingSummaryRow) []domain.GigRatingSummary {
	result := make([]domain.GigRatingSummary, 0, len(rows))
	for _, row := range rows {
		summary := MapRatingSummaryRowToDomain(row)
		result = append(result, domain.GigRatingSummary{
			GigID:         row.OwnerID,
			RatingSummary: summary,
		})
	}
	return result
}

// MapSellerRatingSummaryRowsToDomain maps grouped seller rating rows into
// domain summaries.
func MapSellerRatingSummaryRowsToDomain(rows []model.RatingSummaryRow) []domain.SellerRatingSummary {
	result := make([]domain.SellerRatingSummary, 0, len(rows))
	for _, row := range rows {
		summary := MapRatingSummaryRowToDomain(row)
		result = append(result, domain.SellerRatingSummary{
			SellerID:      row.OwnerID,
			RatingSummary: summary,
		})
	}
	return result
}
