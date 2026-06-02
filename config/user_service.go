package config

// UserServiceConfig defines the outbound user-service gRPC endpoint used to
// enrich review authors.
type UserServiceConfig struct {
	Address string `env:"USER_SERVICE_ADDRESS" envDefault:"127.0.0.1:9502"`
}
