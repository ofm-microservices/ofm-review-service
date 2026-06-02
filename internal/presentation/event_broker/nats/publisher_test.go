package nats

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"review-service/config"
	domain "review-service/internal/domain"
	eventbroker "review-service/internal/presentation/event_broker"
)

type publisherBrokerStub struct {
	subjects []string
	payloads [][]byte
}

func (b *publisherBrokerStub) Publish(_ context.Context, subject string, payload []byte) error {
	b.subjects = append(b.subjects, subject)
	b.payloads = append(b.payloads, append([]byte(nil), payload...))
	return nil
}

func (b *publisherBrokerStub) RunPullConsumer(context.Context, config.PullConsumerConfig, eventbroker.MessageHandler) error {
	return nil
}

func (b *publisherBrokerStub) Close() {}

func TestPublishGigAndUserProjectionUseNilAuthor(t *testing.T) {
	broker := &publisherBrokerStub{}
	p := &ReviewPublisher{
		broker:       broker,
		gig:          "review.projection.gig",
		user:         "review.projection.user",
		gigRating:    "review.rating.gig",
		sellerRating: "review.rating.seller",
	}
	review := &domain.Review{
		ID:        "review-1",
		OrderID:   "order-1",
		GigID:     "gig-1",
		Content:   "content",
		BuyerID:   "buyer-1",
		Rating:    5,
		SellerID:  "seller-1",
		CreatedAt: time.UnixMicro(1000).UTC(),
	}

	if err := p.PublishGigReviewProjectionRequested(context.Background(), review); err != nil {
		t.Fatalf("PublishGigReviewProjectionRequested returned error: %v", err)
	}
	if err := p.PublishUserReviewProjectionRequested(context.Background(), review); err != nil {
		t.Fatalf("PublishUserReviewProjectionRequested returned error: %v", err)
	}
	if len(broker.payloads) != 2 {
		t.Fatalf("expected two projection payloads, got %d", len(broker.payloads))
	}
	for _, payload := range broker.payloads {
		var evt reviewEvent
		if err := json.Unmarshal(payload, &evt); err != nil {
			t.Fatalf("unexpected payload: %v", err)
		}
		if evt.Author != nil {
			t.Fatalf("expected author to be nil in projection payload, got %#v", evt.Author)
		}
	}
}

func TestPublishRatingRequestsPreservesReviewPayload(t *testing.T) {
	broker := &publisherBrokerStub{}
	p := &ReviewPublisher{
		broker:       broker,
		gig:          "review.projection.gig",
		user:         "review.projection.user",
		gigRating:    "review.rating.gig",
		sellerRating: "review.rating.seller",
	}
	review := &domain.Review{ID: "review-1", GigID: "gig-1", SellerID: "seller-1", Rating: 5, CreatedAt: time.UnixMicro(1000).UTC()}

	if err := p.PublishGigRatingRequested(context.Background(), review); err != nil {
		t.Fatalf("PublishGigRatingRequested returned error: %v", err)
	}
	if err := p.PublishSellerRatingRequested(context.Background(), review); err != nil {
		t.Fatalf("PublishSellerRatingRequested returned error: %v", err)
	}
	if len(broker.payloads) != 2 {
		t.Fatalf("expected two rating payloads, got %d", len(broker.payloads))
	}
	var got domain.Review
	if err := json.Unmarshal(broker.payloads[0], &got); err != nil {
		t.Fatalf("expected review JSON payload, got error %v", err)
	}
	if got.ID != review.ID || got.GigID != review.GigID || got.SellerID != review.SellerID {
		t.Fatalf("unexpected payload: %#v", got)
	}
}

func TestPublishNilBrokerIsNoop(t *testing.T) {
	var p ReviewPublisher
	if err := p.PublishGigReviewProjectionRequested(context.Background(), &domain.Review{}); err != nil {
		t.Fatalf("expected nil broker to be a no-op, got %v", err)
	}
	if err := p.PublishGigRatingRequested(context.Background(), &domain.Review{}); err != nil {
		t.Fatalf("expected nil broker to be a no-op, got %v", err)
	}
}
