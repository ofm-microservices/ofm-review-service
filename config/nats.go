package config

import "time"

// NATSConfig defines NATS streams, subjects, and pull-consumer settings used
// by review-service.
type NATSConfig struct {
	// URL remains only for compatibility adapters and is not required by the
	// Kafka production path.
	URL                           string        `env:"NATS_URL"`
	Review                        string        `env:"NATS_USER"`
	Password                      string        `env:"NATS_PASSWORD"`
	ReviewEventsStream            string        `env:"NATS_STREAM_REVIEW_EVENTS" envDefault:"REVIEW_EVENTS"`
	ReviewGigProjectionSubject    string        `env:"NATS_SUBJECT_REVIEW_GIG_PROJECTION_REQUESTED" envDefault:"review.projection.gig"`
	ReviewUserProjectionSubject   string        `env:"NATS_SUBJECT_REVIEW_USER_PROJECTION_REQUESTED" envDefault:"review.projection.user"`
	ReviewGigRatingSubject        string        `env:"NATS_SUBJECT_REVIEW_GIG_RATING_REQUESTED" envDefault:"review.rating.gig"`
	ReviewSellerRatingSubject     string        `env:"NATS_SUBJECT_REVIEW_SELLER_RATING_REQUESTED" envDefault:"review.rating.seller"`
	ReviewAuthorRetrySubject      string        `env:"NATS_SUBJECT_REVIEW_AUTHOR_RETRY_REQUESTED" envDefault:"review.author.retry.requested"`
	ReviewAvatarRetrySubject      string        `env:"NATS_SUBJECT_REVIEW_AVATAR_RETRY_REQUESTED" envDefault:"review.avatar.retry.requested"`
	ReviewGigProjectionDurable    string        `env:"NATS_DURABLE_REVIEW_GIG_PROJECTION" envDefault:"review_service_gig_projection"`
	ReviewUserProjectionDurable   string        `env:"NATS_DURABLE_REVIEW_USER_PROJECTION" envDefault:"review_service_user_projection"`
	ReviewGigRatingDurable        string        `env:"NATS_DURABLE_REVIEW_GIG_RATING" envDefault:"review_service_gig_rating"`
	ReviewSellerRatingDurable     string        `env:"NATS_DURABLE_REVIEW_SELLER_RATING" envDefault:"review_service_seller_rating"`
	ReviewAuthorRetryDurable      string        `env:"NATS_DURABLE_REVIEW_AUTHOR_RETRY" envDefault:"review_service_author_retry"`
	ReviewAvatarRetryDurable      string        `env:"NATS_DURABLE_REVIEW_AVATAR_RETRY" envDefault:"review_service_avatar_retry"`
	ReviewBatchSize               int           `env:"NATS_REVIEW_BATCH_SIZE" envDefault:"100"`
	ReviewMaxWait                 time.Duration `env:"NATS_REVIEW_MAX_WAIT" envDefault:"500ms"`
	ReviewWorkers                 int           `env:"NATS_REVIEW_WORKERS" envDefault:"4"`
	ReviewQueueSize               int           `env:"NATS_REVIEW_QUEUE_SIZE" envDefault:"500"`
	ReviewAckWait                 time.Duration `env:"NATS_REVIEW_ACK_WAIT" envDefault:"30s"`
	ReviewMaxDeliver              int           `env:"NATS_REVIEW_MAX_DELIVER" envDefault:"5"`
	SagaCommandsStream            string        `env:"NATS_STREAM_SAGA_COMMANDS" envDefault:"SAGA_REVIEW_COMMANDS"`
	SagaCreateReviewSubject       string        `env:"NATS_SUBJECT_SAGA_CREATE_USER" envDefault:"saga.review.create"`
	SagaDeleteReviewSubject       string        `env:"NATS_SUBJECT_SAGA_DELETE_USER" envDefault:"saga.review.delete"`
	SagaCreateReviewResultSubject string        `env:"NATS_SUBJECT_SAGA_CREATE_USER_RESULT" envDefault:"saga.review.create.result"`
	SagaDeleteReviewResultSubject string        `env:"NATS_SUBJECT_SAGA_DELETE_USER_RESULT" envDefault:"saga.review.delete.result"`
	SagaCreateReviewDurable       string        `env:"NATS_DURABLE_SAGA_CREATE_USER" envDefault:"review_service_saga_create"`
	SagaDeleteReviewDurable       string        `env:"NATS_DURABLE_SAGA_DELETE_USER" envDefault:"review_service_saga_delete"`
	SagaBatchSize                 int           `env:"NATS_SAGA_BATCH_SIZE" envDefault:"100"`
	SagaMaxWait                   time.Duration `env:"NATS_SAGA_MAX_WAIT" envDefault:"500ms"`
	SagaWorkers                   int           `env:"NATS_SAGA_WORKERS" envDefault:"8"`
	SagaQueueSize                 int           `env:"NATS_SAGA_QUEUE_SIZE" envDefault:"500"`
	SagaAckWait                   time.Duration `env:"NATS_SAGA_ACK_WAIT" envDefault:"30s"`
	SagaMaxDeliver                int           `env:"NATS_SAGA_MAX_DELIVER" envDefault:"5"`
	SagaAdaptiveEnabled           bool          `env:"NATS_SAGA_ADAPTIVE_ENABLED" envDefault:"false"`
	SagaAdaptiveCheckInterval     time.Duration `env:"NATS_SAGA_ADAPTIVE_CHECK_INTERVAL" envDefault:"2s"`
	SagaAdaptiveMediumPending     int           `env:"NATS_SAGA_ADAPTIVE_MEDIUM_PENDING" envDefault:"200"`
	SagaAdaptiveHighPending       int           `env:"NATS_SAGA_ADAPTIVE_HIGH_PENDING" envDefault:"1000"`
	SagaAdaptiveLowBatchSize      int           `env:"NATS_SAGA_ADAPTIVE_LOW_BATCH_SIZE" envDefault:"25"`
	SagaAdaptiveLowMaxWait        time.Duration `env:"NATS_SAGA_ADAPTIVE_LOW_MAX_WAIT" envDefault:"1s"`
	SagaAdaptiveMediumBatchSize   int           `env:"NATS_SAGA_ADAPTIVE_MEDIUM_BATCH_SIZE" envDefault:"100"`
	SagaAdaptiveMediumMaxWait     time.Duration `env:"NATS_SAGA_ADAPTIVE_MEDIUM_MAX_WAIT" envDefault:"500ms"`
	SagaAdaptiveHighBatchSize     int           `env:"NATS_SAGA_ADAPTIVE_HIGH_BATCH_SIZE" envDefault:"300"`
	SagaAdaptiveHighMaxWait       time.Duration `env:"NATS_SAGA_ADAPTIVE_HIGH_MAX_WAIT" envDefault:"100ms"`
}
