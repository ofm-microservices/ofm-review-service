package config

// KafkaConfig defines the Kafka broker and consumer group used by review-service.
type KafkaConfig struct {
	NATSConfig
	Brokers                []string `env:"KAFKA_BROKERS" envSeparator:"," envDefault:"127.0.0.1:9092"`
	GroupID                string   `env:"KAFKA_REVIEW_GROUP_ID" envDefault:"review-service"`
	DeadLetterTopic        string   `env:"KAFKA_REVIEW_DLQ_TOPIC" envDefault:"migration.dead-letter"`
	RecoveryTopic          string   `env:"KAFKA_REVIEW_RECOVERY_TOPIC" envDefault:"migration.recovery.commands.review"`
	RecoveryGroup          string   `env:"KAFKA_REVIEW_RECOVERY_GROUP" envDefault:"review-service-recovery"`
	RecoveryCompletedTopic string   `env:"KAFKA_REVIEW_RECOVERY_COMPLETED_TOPIC" envDefault:"migration.recovery.completed"`
}
