package grpc

import (
	"time"

	reviewv1 "github.com/ofm-microservices/ofm-common/proto/review/v1"
	domain "review-service/internal/domain"
)

// ReviewMapper translates review domain entities into gRPC payloads.
type ReviewMapper interface {
	ToProto(review *domain.Review) *reviewv1.Review
	ToRatingSummaryProto(summary *domain.RatingSummary) *reviewv1.RatingSummary
}

type reviewMapper struct{}

// NewReviewMapper constructs the gRPC review mapper.
func NewReviewMapper() ReviewMapper {
	return reviewMapper{}
}

func (reviewMapper) ToProto(review *domain.Review) *reviewv1.Review {
	if review == nil {
		return nil
	}
	out := &reviewv1.Review{
		ReviewId:    review.ID,
		OrderId:     review.OrderID,
		GigId:       review.GigID,
		BuyerUserId: review.BuyerID,
		Content:     review.Content,
		Rating:      review.Rating,
		CreatedAt:   review.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
	if review.Author != nil {
		out.Author = &reviewv1.ReviewAuthor{
			UserId:      review.Author.UserID,
			Username:    review.Author.Username,
			DisplayName: review.Author.DisplayName,
			AvatarUrl:   review.Author.AvatarURL,
		}
	}
	return out
}

// ToRatingSummaryProto maps the domain rating summary into the gRPC response payload.
func (reviewMapper) ToRatingSummaryProto(summary *domain.RatingSummary) *reviewv1.RatingSummary {
	if summary == nil {
		return nil
	}
	return &reviewv1.RatingSummary{
		RatingAvg:    summary.RatingAvg,
		TotalReviews: summary.TotalReviews,
		Stars_5:      summary.Stars5,
		Stars_4:      summary.Stars4,
		Stars_3:      summary.Stars3,
		Stars_2:      summary.Stars2,
		Stars_1:      summary.Stars1,
	}
}
