package nats

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	"review-service/config"
	app "review-service/internal/application"
	domain "review-service/internal/domain"
	eventbroker "review-service/internal/presentation/event_broker"
)

// ReviewProjectionSubscriber runs the review projection and retry consumers.
type ReviewProjectionSubscriber struct {
	broker eventbroker.EventBroker
	svc    app.ReviewService
	read   domain.ReviewReadRepository
	coord  app.ReviewWindowCoordinator
	users  app.UserPreviewClient
	cfg    *config.Config
	log    logging.Logger
}

// NewReviewProjectionSubscriber constructs the projection subscriber used by
// review-service to populate its Redis read model.
func NewReviewProjectionSubscriber(
	broker eventbroker.EventBroker,
	svc app.ReviewService,
	read domain.ReviewReadRepository,
	coord app.ReviewWindowCoordinator,
	users app.UserPreviewClient,
	cfg *config.Config,
	log logging.Logger,
) (*ReviewProjectionSubscriber, error) {
	if broker == nil {
		return nil, ErrNilEventBroker
	}
	if svc == nil {
		return nil, ErrNilReviewService
	}
	if read == nil {
		return nil, ErrNilReviewReadRepository
	}
	if coord == nil {
		return nil, ErrNilReviewWindowCoordinator
	}
	if users == nil {
		return nil, ErrNilUserPreviewClient
	}
	if cfg == nil {
		return nil, ErrNilConfig
	}
	if log == nil {
		return nil, ErrNilLogger
	}

	return &ReviewProjectionSubscriber{
		broker: broker,
		svc:    svc,
		read:   read,
		coord:  coord,
		users:  users,
		cfg:    cfg,
		log:    log.With(logging.String("module", "review-projection-subscriber")),
	}, nil
}

// Start starts all review projection consumers.
func (s *ReviewProjectionSubscriber) Start(ctx context.Context) error {
	s.StartConsumer(ctx, s.cfg.NATS.ReviewGigProjectionSubject, s.cfg.NATS.ReviewGigProjectionDurable, s.HandleGigProjectionRequested, "")
	s.StartConsumer(ctx, s.cfg.NATS.ReviewUserProjectionSubject, s.cfg.NATS.ReviewUserProjectionDurable, s.HandleUserProjectionRequested, "")
	s.StartConsumer(ctx, s.cfg.NATS.ReviewGigRatingSubject, s.cfg.NATS.ReviewGigRatingDurable, s.HandleGigRatingRequested, "new")
	s.StartConsumer(ctx, s.cfg.NATS.ReviewSellerRatingSubject, s.cfg.NATS.ReviewSellerRatingDurable, s.HandleSellerRatingRequested, "new")
	s.StartConsumer(ctx, s.cfg.NATS.ReviewAuthorRetrySubject, s.cfg.NATS.ReviewAuthorRetryDurable, s.HandleAuthorRetryRequested, "")
	s.StartConsumer(ctx, s.cfg.NATS.ReviewAvatarRetrySubject, s.cfg.NATS.ReviewAvatarRetryDurable, s.HandleAvatarRetryRequested, "")
	return nil
}

func (s *ReviewProjectionSubscriber) StartConsumer(ctx context.Context, subject, durable string, handler func(context.Context, *domain.Review) error, deliverPolicy string) {
	cfg := config.PullConsumerConfig{
		Stream:        s.cfg.NATS.ReviewEventsStream,
		Subject:       subject,
		Durable:       durable,
		BatchSize:     s.cfg.NATS.ReviewBatchSize,
		MaxWait:       s.cfg.NATS.ReviewMaxWait,
		Workers:       s.cfg.NATS.ReviewWorkers,
		QueueSize:     s.cfg.NATS.ReviewQueueSize,
		AckWait:       s.cfg.NATS.ReviewAckWait,
		MaxDeliver:    s.cfg.NATS.ReviewMaxDeliver,
		DeliverPolicy: deliverPolicy,
		Adaptive:      config.PullAdaptiveConfig{},
	}

	go func() {
		if err := s.broker.RunPullConsumer(ctx, cfg, func(msgCtx context.Context, subject string, payload []byte) error {
			var review domain.Review
			if err := json.Unmarshal(payload, &review); err != nil {
				return err
			}
			review = SanitizeReview(review)
			return handler(msgCtx, &review)
		}); err != nil {
			s.log.Error("review projection consumer failed",
				logging.String("subject", subject),
				logging.String("durable", durable),
				logging.Err(err),
			)
		}
	}()
}

// HandleGigProjectionRequested enqueues the gig-side review window job.
func (s *ReviewProjectionSubscriber) HandleGigProjectionRequested(ctx context.Context, review *domain.Review) error {
	projected, err := s.svc.ProjectReview(ctx, review, app.ProjectionPolicy{EnrichAuthor: true, EnrichAvatar: true})
	if err != nil {
		return err
	}
	if projected == nil || projected.Review == nil {
		return nil
	}
	if err := s.coord.EnqueueGigReview(ctx, projected.Review); err != nil {
		return err
	}
	if projected.AuthorRetry {
		if err := s.PublishRetry(ctx, s.cfg.NATS.ReviewAuthorRetrySubject, projected.Review); err != nil {
			s.log.Error("review author retry publish failed",
				logging.Operation("review.retry.publish"),
				logging.String("review_id", projected.Review.ID),
				logging.Err(err),
			)
		}
	}
	if projected.AvatarRetry {
		if err := s.PublishRetry(ctx, s.cfg.NATS.ReviewAvatarRetrySubject, projected.Review); err != nil {
			s.log.Error("review avatar retry publish failed",
				logging.Operation("review.retry.publish"),
				logging.String("review_id", projected.Review.ID),
				logging.Err(err),
			)
		}
	}
	return nil
}

// HandleUserProjectionRequested enqueues the seller-side review window job.
func (s *ReviewProjectionSubscriber) HandleUserProjectionRequested(ctx context.Context, review *domain.Review) error {
	projected, err := s.svc.ProjectReview(ctx, review, app.ProjectionPolicy{EnrichAuthor: true, EnrichAvatar: true})
	if err != nil {
		return err
	}
	if projected == nil || projected.Review == nil {
		return nil
	}
	if err := s.coord.EnqueueSellerReview(ctx, projected.Review); err != nil {
		return err
	}
	if projected.AuthorRetry {
		if err := s.PublishRetry(ctx, s.cfg.NATS.ReviewAuthorRetrySubject, projected.Review); err != nil {
			s.log.Error("review author retry publish failed",
				logging.Operation("review.retry.publish"),
				logging.String("review_id", projected.Review.ID),
				logging.Err(err),
			)
		}
	}
	if projected.AvatarRetry {
		if err := s.PublishRetry(ctx, s.cfg.NATS.ReviewAvatarRetrySubject, projected.Review); err != nil {
			s.log.Error("review avatar retry publish failed",
				logging.Operation("review.retry.publish"),
				logging.String("review_id", projected.Review.ID),
				logging.Err(err),
			)
		}
	}
	return nil
}

// HandleGigRatingRequested updates the gig rating summary projection.
func (s *ReviewProjectionSubscriber) HandleGigRatingRequested(ctx context.Context, review *domain.Review) error {
	return s.WriteGigRating(ctx, review)
}

// HandleSellerRatingRequested updates the seller rating summary projection.
func (s *ReviewProjectionSubscriber) HandleSellerRatingRequested(ctx context.Context, review *domain.Review) error {
	return s.WriteSellerRating(ctx, review)
}

// HandleAuthorRetryRequested retries seller-side projection enrichment.
func (s *ReviewProjectionSubscriber) HandleAuthorRetryRequested(ctx context.Context, review *domain.Review) error {
	projected, err := s.svc.ProjectReview(ctx, review, app.ProjectionPolicy{EnrichAuthor: true, EnrichAvatar: true})
	if err != nil {
		return err
	}
	if projected == nil || projected.Review == nil {
		return nil
	}
	if err := s.coord.EnqueueSellerReview(ctx, projected.Review); err != nil {
		return err
	}
	if projected.AuthorRetry {
		s.log.Warn("review author still unavailable after retry",
			logging.Operation("review.retry.enqueue"),
			logging.String("review_id", projected.Review.ID),
		)
	}
	return nil
}

// HandleAvatarRetryRequested retries seller-side avatar enrichment.
func (s *ReviewProjectionSubscriber) HandleAvatarRetryRequested(ctx context.Context, review *domain.Review) error {
	projected, err := s.svc.ProjectReview(ctx, review, app.ProjectionPolicy{EnrichAvatar: true})
	if err != nil {
		return err
	}
	if projected == nil || projected.Review == nil {
		return nil
	}
	if err := s.coord.EnqueueSellerReview(ctx, projected.Review); err != nil {
		return err
	}
	if projected.AvatarRetry {
		s.log.Warn("review avatar still unavailable after retry",
			logging.Operation("review.retry.enqueue"),
			logging.String("review_id", projected.Review.ID),
		)
	}
	return nil
}

// WriteGigRating updates the cached gig rating summary.
func (s *ReviewProjectionSubscriber) WriteGigRating(ctx context.Context, review *domain.Review) error {
	summary, err := s.read.GetGigRatingSummary(ctx, review.GigID)
	if err == nil {
		if err := s.read.UpsertGigRating(ctx, review.GigID, review.Rating); err != nil {
			return err
		}
		return nil
	}
	if !errors.Is(err, domain.ErrReviewNotFound) {
		return err
	}
	if summary, err = s.svc.GetGigRatingSummary(ctx, review.GigID); err != nil {
		return err
	} else if summary != nil {
		return nil
	}
	return nil
}

// WriteSellerRating updates the cached seller rating summary.
func (s *ReviewProjectionSubscriber) WriteSellerRating(ctx context.Context, review *domain.Review) error {
	sellerID := review.SellerID
	if sellerID == "" && review.Author != nil {
		sellerID = review.Author.UserID
	}
	if strings.TrimSpace(sellerID) == "" {
		return domain.ErrInvalidSellerID
	}
	preview, err := s.users.GetUserPreviewByID(ctx, sellerID)
	if err != nil {
		return err
	}
	if preview == nil {
		return domain.ErrReviewNotFound
	}
	username := strings.TrimSpace(preview.Username)
	if username == "" {
		return domain.ErrInvalidUsername
	}
	summary, err := s.read.GetSellerRatingSummaryByUsername(ctx, username)
	if err == nil {
		if err := s.read.UpsertSellerRatingByUsername(ctx, username, review.Rating); err != nil {
			return err
		}
		return nil
	}
	if !errors.Is(err, domain.ErrReviewNotFound) {
		return err
	}
	if summary, err = s.svc.GetUserRatingSummaryByUsername(ctx, username); err != nil {
		return err
	} else if summary != nil {
		return nil
	}
	return nil
}

// PublishRetry emits a best-effort retry event.
func (s *ReviewProjectionSubscriber) PublishRetry(ctx context.Context, subject string, review *domain.Review) error {
	payload, err := json.Marshal(review)
	if err != nil {
		return err
	}
	return s.broker.Publish(ctx, subject, payload)
}

// SanitizeReview normalizes the review payload before processing.
func SanitizeReview(review domain.Review) domain.Review {
	if review.Author != nil {
		review.Author.AvatarURL = strings.TrimSpace(review.Author.AvatarURL)
	}
	return review
}
