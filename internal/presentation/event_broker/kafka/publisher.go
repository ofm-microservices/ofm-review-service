package kafka

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

type reviewPublisher struct {
	broker interface {
		Publish(context.Context, string, []byte) error
	}
	gig, user, gigRating, sellerRating string
}

// NewReviewPublisher constructs the Kafka review projection publisher.
func NewReviewPublisher(broker interface {
	Publish(context.Context, string, []byte) error
}, cfg *config.Config) app.ReviewPublisher {
	return &reviewPublisher{broker: broker, gig: cfg.Kafka.ReviewGigProjectionSubject, user: cfg.Kafka.ReviewUserProjectionSubject, gigRating: cfg.Kafka.ReviewGigRatingSubject, sellerRating: cfg.Kafka.ReviewSellerRatingSubject}
}

func (p *reviewPublisher) publish(ctx context.Context, topic string, value any) error {
	if p == nil || p.broker == nil {
		return nil
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return p.broker.Publish(ctx, topic, payload)
}

func (p *reviewPublisher) PublishGigReviewProjectionRequested(ctx context.Context, review *domain.Review) error {
	if review == nil {
		return nil
	}
	return p.publish(ctx, p.gig, reviewEvent{ID: review.ID, OrderID: review.OrderID, GigID: review.GigID, Content: review.Content, BuyerID: review.BuyerID, Rating: review.Rating, SellerID: review.SellerID, CreatedAt: review.CreatedAt, UpdatedAt: review.UpdatedAt})
}
func (p *reviewPublisher) PublishUserReviewProjectionRequested(ctx context.Context, review *domain.Review) error {
	if review == nil {
		return nil
	}
	return p.publish(ctx, p.user, reviewEvent{ID: review.ID, OrderID: review.OrderID, GigID: review.GigID, Content: review.Content, BuyerID: review.BuyerID, Rating: review.Rating, SellerID: review.SellerID, CreatedAt: review.CreatedAt, UpdatedAt: review.UpdatedAt})
}
func (p *reviewPublisher) PublishGigRatingRequested(ctx context.Context, review *domain.Review) error {
	if review == nil {
		return nil
	}
	return p.publish(ctx, p.gigRating, review)
}
func (p *reviewPublisher) PublishSellerRatingRequested(ctx context.Context, review *domain.Review) error {
	if review == nil {
		return nil
	}
	return p.publish(ctx, p.sellerRating, review)
}
