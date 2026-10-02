package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"github.com/ofm-microservices/ofm-common/pkg/migration/events"
	"github.com/ofm-microservices/ofm-common/pkg/resilience"
	"review-service/config"
	app "review-service/internal/application"
	eventbroker "review-service/internal/presentation/event_broker"
)

// RecoverySubscriber applies review fallback commands through review application use-cases.
type RecoverySubscriber interface{ Subscribe(context.Context) error }
type recoverySubscriber struct {
	broker eventbroker.EventBroker
	svc    app.ReviewService
	cfg    config.KafkaConfig
	log    logging.Logger
}

// NewRecoverySubscriber constructs the review recovery Kafka adapter.
func NewRecoverySubscriber(b eventbroker.EventBroker, svc app.ReviewService, cfg config.KafkaConfig, log logging.Logger) (RecoverySubscriber, error) {
	if b == nil || svc == nil || log == nil {
		return nil, errors.New("invalid review recovery subscriber dependency")
	}
	return &recoverySubscriber{broker: b, svc: svc, cfg: cfg, log: log.With(logging.String("module", "kafka-review-recovery-subscriber"))}, nil
}
func (s *recoverySubscriber) Subscribe(ctx context.Context) error {
	return s.broker.RunPullConsumer(ctx, config.PullConsumerConfig{Subject: s.cfg.RecoveryTopic, GroupID: s.cfg.RecoveryGroup}, s.handle)
}
func (s *recoverySubscriber) handle(ctx context.Context, _ string, raw []byte) error {
	var cmd events.Envelope
	if err := json.Unmarshal(raw, &cmd); err != nil {
		return fmt.Errorf("decode review recovery command: %w", err)
	}
	if !strings.EqualFold(cmd.AggregateType, "review") && !strings.EqualFold(cmd.AggregateType, "reviews") {
		return fmt.Errorf("unsupported review recovery aggregate_type=%q", cmd.AggregateType)
	}
	var p struct {
		ReviewID       string `json:"review_id"`
		OrderID        string `json:"order_id"`
		BuyerID        string `json:"buyer_id"`
		BuyerUsername  string `json:"buyer_username"`
		Content        string `json:"content"`
		Rating         int32  `json:"rating"`
		IdempotencyKey string `json:"idempotency_key"`
		RequestedAt    string `json:"requested_at"`
	}
	if err := json.Unmarshal(cmd.Payload, &p); err != nil {
		return fmt.Errorf("decode review recovery payload: %w", err)
	}
	if strings.TrimSpace(p.ReviewID) == "" {
		return resilience.Permanent(fmt.Errorf("review recovery command %s is missing review_id", cmd.CommandID))
	}
	if p.OrderID == "" {
		const marker = "/orders/"
		if start := strings.Index(cmd.CommandPath, marker); start >= 0 {
			p.OrderID = strings.SplitN(cmd.CommandPath[start+len(marker):], "/", 2)[0]
		}
	}
	if p.BuyerID == "" {
		p.BuyerID = cmd.RecoveryPrincipalID
	}
	if p.IdempotencyKey == "" {
		p.IdempotencyKey = cmd.IdempotencyKey
	}
	if p.RequestedAt == "" {
		p.RequestedAt = cmd.OccurredAt.UTC().Format(time.RFC3339Nano)
	}
	result, err := s.svc.CreateReview(ctx, app.CreateReviewCommand{ReviewID: p.ReviewID, OrderID: p.OrderID, BuyerID: p.BuyerID, BuyerUsername: p.BuyerUsername, Content: p.Content, Rating: p.Rating, IdempotencyKey: p.IdempotencyKey, RequestedAt: p.RequestedAt})
	if err != nil {
		return err
	}
	aggregateID := cmd.AggregateID
	if result != nil && result.ID != "" {
		aggregateID = result.ID
	}
	body, err := json.Marshal(events.Envelope{EventID: cmd.EventID + ".completed", CommandID: cmd.CommandID, CorrelationID: cmd.CorrelationID, CausationID: cmd.EventID, IdempotencyKey: cmd.IdempotencyKey, TestRunID: cmd.TestRunID, EventType: "migration.recovery.completed", Operation: cmd.Operation, SchemaVersion: 1, AggregateType: "review", AggregateID: aggregateID, SourceService: "review-service-recovery", OccurredAt: time.Now().UTC(), Payload: marshal(result)})
	if err != nil {
		return err
	}
	return s.broker.Publish(context.WithoutCancel(ctx), s.cfg.RecoveryCompletedTopic, body)
}
func marshal(v any) []byte { b, _ := json.Marshal(v); return b }
