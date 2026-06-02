package config

// FileServiceConfig defines the outbound file-service gRPC endpoint used to
// resolve review author avatar URLs.
type FileServiceConfig struct {
	Address string `env:"FILE_SERVICE_ADDRESS" envDefault:"127.0.0.1:9503"`
}
