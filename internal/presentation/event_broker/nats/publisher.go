package nats

import (
	"context"
	"encoding/json"
	"time"

	"review-service/config"
	app "review-service/internal/application"
	domain "review-service/internal/domain"
)

type reviewEvent struct {
	ID        string               `json:"review_id"`
	OrderID   string               `json:"order_id"`
	GigID     string               `json:"gig_id"`
	Content   string               `json:"content"`
	BuyerID   string               `json:"buyer_id"`
	Rating    int32                `json:"rating"`
	SellerID  string               `json:"seller_id"`
	Author    *domain.ReviewAuthor `json:"author"`
	CreatedAt time.Time            `json:"created_at"`
	UpdatedAt time.Time            `json:"updated_at"`
}

// ReviewPublisher publishes review projection jobs to NATS.
type ReviewPublisher struct {
	broker       eventBroker
	gig          string
	user         string
	gigRating    string
	sellerRating string
}

type eventBroker interface {
	Publish(ctx context.Context, subject string, payload []byte) error
}

// NewReviewPublisher constructs the review projection publisher.
func NewReviewPublisher(broker eventBroker, cfg *config.Config) app.ReviewPublisher {
	return &ReviewPublisher{
		broker:       broker,
		gig:          cfg.NATS.ReviewGigProjectionSubject,
		user:         cfg.NATS.ReviewUserProjectionSubject,
		gigRating:    cfg.NATS.ReviewGigRatingSubject,
		sellerRating: cfg.NATS.ReviewSellerRatingSubject,
	}
}

func (p *ReviewPublisher) PublishGigReviewProjectionRequested(ctx context.Context, review *domain.Review) error {
	if review == nil || p == nil || p.broker == nil {
		return nil
	}
	payload, err := json.Marshal(reviewEvent{
		ID:        review.ID,
		OrderID:   review.OrderID,
		GigID:     review.GigID,
		Content:   review.Content,
		BuyerID:   review.BuyerID,
		Rating:    review.Rating,
		SellerID:  review.SellerID,
		Author:    nil,
		CreatedAt: review.CreatedAt,
		UpdatedAt: review.UpdatedAt,
	})
	if err != nil {
		return err
	}
	return p.broker.Publish(ctx, p.gig, payload)
}

func (p *ReviewPublisher) PublishUserReviewProjectionRequested(ctx context.Context, review *domain.Review) error {
	if review == nil || p == nil || p.broker == nil {
		return nil
	}
	payload, err := json.Marshal(reviewEvent{
		ID:        review.ID,
		OrderID:   review.OrderID,
		GigID:     review.GigID,
		Content:   review.Content,
		BuyerID:   review.BuyerID,
		Rating:    review.Rating,
		SellerID:  review.SellerID,
		Author:    nil,
		CreatedAt: review.CreatedAt,
		UpdatedAt: review.UpdatedAt,
	})
	if err != nil {
		return err
	}
	return p.broker.Publish(ctx, p.user, payload)
}

// PublishGigRatingRequested emits the gig rating update event consumed by the
// rating projection subscriber.
func (p *ReviewPublisher) PublishGigRatingRequested(ctx context.Context, review *domain.Review) error {
	if review == nil || p == nil || p.broker == nil {
		return nil
	}
	payload, err := json.Marshal(review)
	if err != nil {
		return err
	}
	return p.broker.Publish(ctx, p.gigRating, payload)
}

// PublishSellerRatingRequested emits the seller rating update event consumed
// by the rating projection subscriber.
func (p *ReviewPublisher) PublishSellerRatingRequested(ctx context.Context, review *domain.Review) error {
	if review == nil || p == nil || p.broker == nil {
		return nil
	}
	payload, err := json.Marshal(review)
	if err != nil {
		return err
	}
	return p.broker.Publish(ctx, p.sellerRating, payload)
}
