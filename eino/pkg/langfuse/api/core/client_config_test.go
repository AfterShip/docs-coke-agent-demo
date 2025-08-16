package core

import (
	"testing"
	"time"

	"github.com/go-resty/resty/v2"

	"eino/pkg/langfuse/config"
)

func TestConfigureRestyClient(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Config
		client  *resty.Client
		wantErr bool
		validate func(t *testing.T, c *resty.Client)
	}{
		{
			name:   "basic configuration",
			config: &config.Config{
				Host: "https://api.example.com",
				HTTPUserAgent: "test-agent",
				Timeout: 30 * time.Second,
			},
			client: resty.New(),
			wantErr: false,
			validate: func(t *testing.T, c *resty.Client) {
				if c.BaseURL != "https://api.example.com" {
					t.Errorf("Expected base URL 'https://api.example.com', got '%s'", c.BaseURL)
				}
			},
		},
		{
			name: "with authentication",
			config: &config.Config{
				Host: "https://api.example.com",
				PublicKey: "test-public",
				SecretKey: "test-secret",
				HTTPUserAgent: "test-agent",
				Timeout: 30 * time.Second,
			},
			client: resty.New(),
			wantErr: false,
			validate: func(t *testing.T, c *resty.Client) {
				if c.BaseURL != "https://api.example.com" {
					t.Errorf("Expected base URL 'https://api.example.com', got '%s'", c.BaseURL)
				}
			},
		},
		{
			name: "with retry configuration",
			config: &config.Config{
				Host: "https://api.example.com",
				HTTPUserAgent: "test-agent",
				Timeout: 30 * time.Second,
				RetryCount: 3,
				RetryDelay: 1 * time.Second,
				MaxRetryDelay: 10 * time.Second,
			},
			client: resty.New(),
			wantErr: false,
			validate: func(t *testing.T, c *resty.Client) {
				if c.BaseURL != "https://api.example.com" {
					t.Errorf("Expected base URL 'https://api.example.com', got '%s'", c.BaseURL)
				}
			},
		},
		{
			name: "with debug mode",
			config: &config.Config{
				Host: "https://api.example.com",
				HTTPUserAgent: "test-agent",
				Timeout: 30 * time.Second,
				Debug: true,
			},
			client: resty.New(),
			wantErr: false,
			validate: func(t *testing.T, c *resty.Client) {
				if c.BaseURL != "https://api.example.com" {
					t.Errorf("Expected base URL 'https://api.example.com', got '%s'", c.BaseURL)
				}
			},
		},
		{
			name:    "nil client",
			config:  &config.Config{Host: "https://api.example.com"},
			client:  nil,
			wantErr: true,
		},
		{
			name:    "nil config",
			config:  nil,
			client:  resty.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ConfigureRestyClient(tt.client, tt.config)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ConfigureRestyClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, tt.client)
			}

			// Note: Resty v2 doesn't require explicit cleanup
		})
	}
}

func TestConfigureRestyClientDefaults(t *testing.T) {
	config := config.DefaultConfig()
	client := resty.New()

	err := ConfigureRestyClient(client, config)
	if err != nil {
		t.Fatalf("ConfigureRestyClient() failed: %v", err)
	}

	// Verify basic configuration was applied
	if client.BaseURL != config.Host {
		t.Errorf("Expected base URL '%s', got '%s'", config.Host, client.BaseURL)
	}
}