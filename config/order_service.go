package config

// OrderServiceConfig defines the outbound order-service gRPC endpoint used to
// validate review writes.
type OrderServiceConfig struct {
	Address string `env:"ORDER_SERVICE_ADDRESS" envDefault:"127.0.0.1:9501"`
}
