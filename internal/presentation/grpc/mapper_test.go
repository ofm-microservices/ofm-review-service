package grpc

import (
	"testing"
	"time"

	domain "review-service/internal/domain"
)

func TestReviewMapperToProto(t *testing.T) {
	m := NewReviewMapper()
	if m.ToProto(nil) != nil {
		t.Fatalf("expected nil review to map to nil")
	}
	review := &domain.Review{
		ID:        "review-1",
		OrderID:   "order-1",
		GigID:     "gig-1",
		BuyerID:   "buyer-1",
		Content:   "content",
		Rating:    5,
		CreatedAt: time.UnixMicro(1234).UTC(),
		Author: &domain.ReviewAuthor{
			UserID:      "buyer-1",
			Username:    "user-1",
			DisplayName: "User One",
			AvatarURL:   "https://example.test/avatar.png",
		},
	}
	proto := m.ToProto(review)
	if proto.GetReviewId() != review.ID || proto.GetBuyerUserId() != review.BuyerID {
		t.Fatalf("unexpected proto mapping: %#v", proto)
	}
	if proto.GetAuthor() == nil || proto.GetAuthor().GetUsername() != "user-1" {
		t.Fatalf("expected author to be mapped, got %#v", proto.GetAuthor())
	}
}

func TestReviewMapperToRatingSummaryProto(t *testing.T) {
	m := NewReviewMapper()
	if m.ToRatingSummaryProto(nil) != nil {
		t.Fatalf("expected nil summary to map to nil")
	}
	proto := m.ToRatingSummaryProto(&domain.RatingSummary{RatingAvg: 4.5, TotalReviews: 3, Stars5: 2, Stars4: 1})
	if proto.GetRatingAvg() != 4.5 || proto.GetTotalReviews() != 3 {
		t.Fatalf("unexpected rating summary mapping: %#v", proto)
	}
}
