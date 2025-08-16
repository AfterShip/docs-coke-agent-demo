package api

import (
	"testing"
	"time"

	"eino/pkg/langfuse/config"
)

func TestNewAPIClientWithRestyClient(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
	}{
		{
			name: "valid configuration",
			config: &config.Config{
				Host:           "https://api.example.com",
				PublicKey:      "test-public-key",
				SecretKey:      "test-secret-key",
				RequestTimeout: 30 * time.Second,
				RetryCount:     3,
				SampleRate:     1.0,
				HTTPUserAgent:  "test-agent",
			},
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "invalid config - missing host",
			config: &config.Config{
				PublicKey:      "test-public-key",
				SecretKey:      "test-secret-key",
				RequestTimeout: 30 * time.Second,
				SampleRate:     1.0,
			},
			wantErr: true,
		},
		{
			name: "invalid config - missing public key",
			config: &config.Config{
				Host:           "https://api.example.com",
				SecretKey:      "test-secret-key",
				RequestTimeout: 30 * time.Second,
				SampleRate:     1.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip health check for tests
			if tt.config != nil {
				tt.config.SkipInitialHealthCheck = true
			}

			client, err := NewAPIClient(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewAPIClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if client == nil {
					t.Error("NewAPIClient() returned nil client without error")
					return
				}

				// Test that resty client is accessible
				restyClient := client.GetRestyClient()
				if restyClient == nil {
					t.Error("GetRestyClient() returned nil")
				}

				// Test basic properties
				if client.GetHost() != tt.config.Host {
					t.Errorf("GetHost() = %v, want %v", client.GetHost(), tt.config.Host)
				}

				if client.IsClosed() {
					t.Error("New client should not be closed")
				}

				// Clean up
				client.Close()
			}
		})
	}
}

func TestAPIClientRestyMethods(t *testing.T) {
	config := &config.Config{
		Host:                   "https://api.example.com",
		PublicKey:              "test-public-key",
		SecretKey:              "test-secret-key",
		RequestTimeout:         30 * time.Second,
		RetryCount:             3,
		SampleRate:             1.0,
		HTTPUserAgent:          "test-agent",
		SkipInitialHealthCheck: true,
	}

	client, err := NewAPIClient(config)
	if err != nil {
		t.Fatalf("NewAPIClient() failed: %v", err)
	}
	defer client.Close()

	t.Run("SetUserAgent", func(t *testing.T) {
		newAgent := "new-test-agent"
		client.SetUserAgent(newAgent)

		restyClient := client.GetRestyClient()
		if restyClient == nil {
			t.Fatal("GetRestyClient() returned nil")
		}

		// Check that the header was updated (access via Header field)
		userAgent := restyClient.Header.Get("User-Agent")
		if userAgent != newAgent {
			t.Errorf("SetUserAgent() = %v, want %v", userAgent, newAgent)
		}
	})

	t.Run("SetTimeout", func(t *testing.T) {
		newTimeout := 60 * time.Second
		client.SetTimeout(newTimeout)

		if client.GetConfig().RequestTimeout != newTimeout {
			t.Errorf("SetTimeout() config not updated, got %v, want %v", 
				client.GetConfig().RequestTimeout, newTimeout)
		}
	})

	t.Run("GetRestyClient after close", func(t *testing.T) {
		client.Close()
		restyClient := client.GetRestyClient()
		if restyClient != nil {
			t.Error("GetRestyClient() should return nil after close")
		}
	})
}

func TestAPIClientRestyClientConfiguration(t *testing.T) {
	config := &config.Config{
		Host:                   "https://test.langfuse.com",
		PublicKey:              "pk-test",
		SecretKey:              "sk-test",
		RequestTimeout:         15 * time.Second,
		RetryCount:             2,
		SampleRate:             1.0,
		HTTPUserAgent:          "langfuse-test",
		Debug:                  true,
		SkipInitialHealthCheck: true,
	}

	client, err := NewAPIClient(config)
	if err != nil {
		t.Fatalf("NewAPIClient() failed: %v", err)
	}
	defer client.Close()

	restyClient := client.GetRestyClient()
	if restyClient == nil {
		t.Fatal("GetRestyClient() returned nil")
	}

	// Verify basic configuration
	if restyClient.BaseURL != config.Host {
		t.Errorf("BaseURL = %v, want %v", restyClient.BaseURL, config.Host)
	}

	if restyClient.Debug != config.Debug {
		t.Errorf("Debug = %v, want %v", restyClient.Debug, config.Debug)
	}

	// Verify headers
	userAgent := restyClient.Header.Get("User-Agent")
	if userAgent != config.HTTPUserAgent {
		t.Errorf("User-Agent = %v, want %v", userAgent, config.HTTPUserAgent)
	}

	contentType := restyClient.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %v, want application/json", contentType)
	}

	accept := restyClient.Header.Get("Accept")
	if accept != "application/json" {
		t.Errorf("Accept = %v, want application/json", accept)
	}
}

func TestAPIClientBackwardCompatibility(t *testing.T) {
	config := &config.Config{
		Host:                   "https://api.example.com",
		PublicKey:              "test-public-key",
		SecretKey:              "test-secret-key",
		RequestTimeout:         30 * time.Second,
		RetryCount:             3,
		SampleRate:             1.0,
		HTTPUserAgent:          "test-agent",
		SkipInitialHealthCheck: true,
	}

	client, err := NewAPIClient(config)
	if err != nil {
		t.Fatalf("NewAPIClient() failed: %v", err)
	}
	defer client.Close()

	// Test that existing methods still work
	t.Run("existing methods work", func(t *testing.T) {
		if client.Health == nil {
			t.Error("Health client should not be nil")
		}

		if client.Ingestion == nil {
			t.Error("Ingestion client should not be nil")
		}

		if client.Traces == nil {
			t.Error("Traces client should not be nil")
		}

		if client.Scores == nil {
			t.Error("Scores client should not be nil")
		}

		if client.Sessions == nil {
			t.Error("Sessions client should not be nil")
		}
	})

	// Test existing utility methods
	t.Run("utility methods", func(t *testing.T) {
		if client.GetHost() != config.Host {
			t.Errorf("GetHost() = %v, want %v", client.GetHost(), config.Host)
		}

		if client.GetSampleRate() != config.SampleRate {
			t.Errorf("GetSampleRate() = %v, want %v", client.GetSampleRate(), config.SampleRate)
		}

		if client.Debug() != config.Debug {
			t.Errorf("Debug() = %v, want %v", client.Debug(), config.Debug)
		}

		if client.IsEnabled() != config.Enabled {
			t.Errorf("IsEnabled() = %v, want %v", client.IsEnabled(), config.Enabled)
		}
	})
}