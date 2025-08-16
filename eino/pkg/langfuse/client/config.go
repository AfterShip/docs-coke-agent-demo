package client

import "eino/pkg/langfuse/config"

// Re-export config types and functions for backward compatibility
type Config = config.Config
type ConfigOption = config.ConfigOption

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	cfg := config.DefaultConfig()
	err := cfg.LoadFromEnvironment()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// NewConfig creates a new configuration with the given options
func NewConfig(opts ...ConfigOption) (*Config, error) {
	return config.NewConfig(opts...)
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return config.DefaultConfig()
}

// Configuration option functions
var (
	WithHost        = config.WithHost
	WithCredentials = config.WithCredentials
	WithPublicKey   = config.WithPublicKey
	WithSecretKey   = config.WithSecretKey
	WithTimeout     = config.WithTimeout
	WithRetryConfig = config.WithRetryConfig
	WithQueueConfig = config.WithQueueConfig
	WithDebug       = config.WithDebug
	WithEnabled     = config.WithEnabled
	WithBatchMode   = config.WithBatchMode
	WithRelease     = config.WithRelease
	WithEnvironment = config.WithEnvironment
	WithUserAgent   = config.WithUserAgent
)
