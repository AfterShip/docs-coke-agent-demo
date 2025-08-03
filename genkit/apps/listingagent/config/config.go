package config

import (
	"sync"
)

// Config is the running configuration structure of the service.
type Config struct {
	*Options
}

var (
	globalConfig *Config
	once         sync.Once
)

// CreateConfigFromOptions creates a running configuration instance based
// on a given IAM pump command line or configuration file option.
func CreateConfigFromOptions(opts *Options) (*Config, error) {
	cfg := &Config{
		Options: opts,
	}
	SetGlobalConfig(cfg)
	return cfg, nil
}

// SetGlobalConfig 设置全局配置 (单例模式)
func SetGlobalConfig(c *Config) {
	once.Do(func() {
		globalConfig = c
	})
}

// GetGlobalConfig 获取全局配置
func GetGlobalConfig() *Config {
	return globalConfig
}
