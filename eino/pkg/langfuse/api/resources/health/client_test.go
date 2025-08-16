package health

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"eino/pkg/langfuse/api/resources/health/types"
)

func TestNewClient(t *testing.T) {
	restyClient := resty.New()

	client := NewClient(restyClient)

	assert.NotNil(t, client)
	assert.Equal(t, restyClient, client.client)
}

func TestClient_Check(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
		expectedStatus types.HealthStatus
	}{
		{
			name: "healthy service",
			serverResponse: `{
				"status": "healthy",
				"version": "1.0.0",
				"timestamp": "2024-01-15T12:00:00Z",
				"environment": "production"
			}`,
			serverStatus:   http.StatusOK,
			expectError:    false,
			expectedStatus: types.HealthStatusHealthy,
		},
		{
			name: "unhealthy service",
			serverResponse: `{
				"status": "unhealthy",
				"timestamp": "2024-01-15T12:00:00Z",
				"services": {
					"database": {
						"status": "unhealthy",
						"message": "Connection timeout"
					}
				}
			}`,
			serverStatus:   http.StatusServiceUnavailable,
			expectError:    false,
			expectedStatus: types.HealthStatusUnhealthy,
		},
		{
			name:          "server error",
			serverStatus:  http.StatusInternalServerError,
			expectError:   true,
			errorContains: "health check request failed",
		},
		{
			name: "degraded service",
			serverResponse: `{
				"status": "degraded",
				"timestamp": "2024-01-15T12:00:00Z",
				"services": {
					"database": {
						"status": "healthy"
					},
					"cache": {
						"status": "unhealthy",
						"message": "Cache miss rate high"
					}
				}
			}`,
			serverStatus:   http.StatusOK,
			expectError:    false,
			expectedStatus: types.HealthStatusDegraded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method and path
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/api/public/health", r.URL.Path)

				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != "" {
					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte(tt.serverResponse))
				}
			}))
			defer server.Close()

			// Setup client
			restyClient := resty.New().SetBaseURL(server.URL)
			client := NewClient(restyClient)

			// Execute test
			ctx := context.Background()
			response, err := client.Check(ctx)

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, tt.expectedStatus, response.Status)
			}
		})
	}
}

func TestClient_CheckWithTimeout(t *testing.T) {
	tests := []struct {
		name          string
		timeout       time.Duration
		serverDelay   time.Duration
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful check within timeout",
			timeout:     100 * time.Millisecond,
			serverDelay: 10 * time.Millisecond,
			expectError: false,
		},
		{
			name:          "timeout exceeded",
			timeout:       50 * time.Millisecond,
			serverDelay:   100 * time.Millisecond,
			expectError:   true,
			errorContains: "context deadline exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server with delay
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(tt.serverDelay)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status": "healthy", "timestamp": "2024-01-15T12:00:00Z"}`))
			}))
			defer server.Close()

			// Setup client
			restyClient := resty.New().SetBaseURL(server.URL)
			client := NewClient(restyClient)

			// Execute test
			response, err := client.CheckWithTimeout(tt.timeout)

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, types.HealthStatusHealthy, response.Status)
			}
		})
	}
}

func TestClient_IsHealthy(t *testing.T) {
	tests := []struct {
		name             string
		serverResponse   string
		serverStatus     int
		expectedHealthy  bool
		expectError      bool
		errorContains    string
	}{
		{
			name: "healthy service",
			serverResponse: `{
				"status": "healthy",
				"timestamp": "2024-01-15T12:00:00Z"
			}`,
			serverStatus:    http.StatusOK,
			expectedHealthy: true,
			expectError:     false,
		},
		{
			name: "unhealthy service",
			serverResponse: `{
				"status": "unhealthy",
				"timestamp": "2024-01-15T12:00:00Z"
			}`,
			serverStatus:    http.StatusServiceUnavailable,
			expectedHealthy: false,
			expectError:     false,
		},
		{
			name: "degraded service",
			serverResponse: `{
				"status": "degraded",
				"timestamp": "2024-01-15T12:00:00Z"
			}`,
			serverStatus:    http.StatusOK,
			expectedHealthy: false,
			expectError:     false,
		},
		{
			name:          "server error",
			serverStatus:  http.StatusInternalServerError,
			expectError:   true,
			errorContains: "health check request failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != "" {
					w.Write([]byte(tt.serverResponse))
				}
			}))
			defer server.Close()

			// Setup client
			restyClient := resty.New().SetBaseURL(server.URL)
			client := NewClient(restyClient)

			// Execute test
			ctx := context.Background()
			healthy, err := client.IsHealthy(ctx)

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedHealthy, healthy)
			}
		})
	}
}

func TestClient_WaitForHealthy(t *testing.T) {
	tests := []struct {
		name               string
		serverBehavior     func(callCount *int) http.HandlerFunc
		checkInterval      time.Duration
		contextTimeout     time.Duration
		expectError        bool
		errorContains      string
		expectedCallCount  int
	}{
		{
			name: "healthy on first check",
			serverBehavior: func(callCount *int) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					*callCount++
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"status": "healthy", "timestamp": "2024-01-15T12:00:00Z"}`))
				}
			},
			checkInterval:     10 * time.Millisecond,
			contextTimeout:    100 * time.Millisecond,
			expectError:       false,
			expectedCallCount: 1,
		},
		{
			name: "healthy on second check",
			serverBehavior: func(callCount *int) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					*callCount++
					if *callCount == 1 {
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{"status": "unhealthy", "timestamp": "2024-01-15T12:00:00Z"}`))
						return
					}
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"status": "healthy", "timestamp": "2024-01-15T12:00:00Z"}`))
				}
			},
			checkInterval:     20 * time.Millisecond,
			contextTimeout:    100 * time.Millisecond,
			expectError:       false,
			expectedCallCount: 2,
		},
		{
			name: "timeout before healthy",
			serverBehavior: func(callCount *int) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					*callCount++
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"status": "unhealthy", "timestamp": "2024-01-15T12:00:00Z"}`))
				}
			},
			checkInterval:  20 * time.Millisecond,
			contextTimeout: 50 * time.Millisecond,
			expectError:    true,
			errorContains:  "context deadline exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			server := httptest.NewServer(tt.serverBehavior(&callCount))
			defer server.Close()

			// Setup client
			restyClient := resty.New().SetBaseURL(server.URL)
			client := NewClient(restyClient)

			// Execute test with timeout context
			ctx, cancel := context.WithTimeout(context.Background(), tt.contextTimeout)
			defer cancel()

			err := client.WaitForHealthy(ctx, tt.checkInterval)

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.expectedCallCount > 0 {
					assert.Equal(t, tt.expectedCallCount, callCount)
				}
			}
		})
	}
}

func TestClient_CheckLiveness(t *testing.T) {
	tests := []struct {
		name          string
		serverStatus  int
		expectError   bool
		errorContains string
	}{
		{
			name:         "liveness check successful",
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:          "liveness check failed",
			serverStatus:  http.StatusServiceUnavailable,
			expectError:   true,
			errorContains: "liveness check failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/api/public/health", r.URL.Path)
				w.WriteHeader(tt.serverStatus)
			}))
			defer server.Close()

			// Setup client
			restyClient := resty.New().SetBaseURL(server.URL)
			client := NewClient(restyClient)

			// Execute test
			ctx := context.Background()
			err := client.CheckLiveness(ctx)

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_CheckReadiness(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
	}{
		{
			name: "ready service",
			serverResponse: `{
				"status": "healthy",
				"timestamp": "2024-01-15T12:00:00Z",
				"services": {
					"database": {"status": "healthy"},
					"cache": {"status": "healthy"}
				}
			}`,
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name: "not ready - unhealthy service",
			serverResponse: `{
				"status": "healthy",
				"timestamp": "2024-01-15T12:00:00Z",
				"services": {
					"database": {"status": "healthy"},
					"cache": {"status": "unhealthy", "message": "Cache unavailable"}
				}
			}`,
			serverStatus:  http.StatusOK,
			expectError:   true,
			errorContains: "critical services are unhealthy",
		},
		{
			name: "not ready - service not healthy",
			serverResponse: `{
				"status": "unhealthy",
				"timestamp": "2024-01-15T12:00:00Z"
			}`,
			serverStatus:  http.StatusServiceUnavailable,
			expectError:   true,
			errorContains: "service is not ready",
		},
		{
			name:          "readiness check failed",
			serverStatus:  http.StatusInternalServerError,
			expectError:   true,
			errorContains: "readiness check failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != "" {
					w.Write([]byte(tt.serverResponse))
				}
			}))
			defer server.Close()

			// Setup client
			restyClient := resty.New().SetBaseURL(server.URL)
			client := NewClient(restyClient)

			// Execute test
			ctx := context.Background()
			err := client.CheckReadiness(ctx)

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_GetServiceHealth(t *testing.T) {
	tests := []struct {
		name           string
		serviceName    string
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
		expectedStatus types.HealthStatus
	}{
		{
			name:        "existing service",
			serviceName: "database",
			serverResponse: `{
				"status": "healthy",
				"timestamp": "2024-01-15T12:00:00Z",
				"services": {
					"database": {
						"status": "healthy",
						"message": "All connections active"
					},
					"cache": {
						"status": "degraded"
					}
				}
			}`,
			serverStatus:   http.StatusOK,
			expectError:    false,
			expectedStatus: types.HealthStatusHealthy,
		},
		{
			name:        "non-existing service",
			serviceName: "nonexistent",
			serverResponse: `{
				"status": "healthy",
				"timestamp": "2024-01-15T12:00:00Z",
				"services": {
					"database": {"status": "healthy"}
				}
			}`,
			serverStatus:  http.StatusOK,
			expectError:   true,
			errorContains: "service 'nonexistent' not found",
		},
		{
			name:          "health check failed",
			serviceName:   "database",
			serverStatus:  http.StatusInternalServerError,
			expectError:   true,
			errorContains: "failed to get service health",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != "" {
					w.Write([]byte(tt.serverResponse))
				}
			}))
			defer server.Close()

			// Setup client
			restyClient := resty.New().SetBaseURL(server.URL)
			client := NewClient(restyClient)

			// Execute test
			ctx := context.Background()
			serviceHealth, err := client.GetServiceHealth(ctx, tt.serviceName)

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, serviceHealth)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, serviceHealth)
				assert.Equal(t, tt.expectedStatus, serviceHealth.Status)
			}
		})
	}
}

func TestClient_Monitor(t *testing.T) {
	// Track callback invocations
	callbackCount := 0
	var lastResponse *types.HealthResponse
	var lastError error

	callback := func(response *types.HealthResponse, err error) {
		callbackCount++
		lastResponse = response
		lastError = err
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "timestamp": "2024-01-15T12:00:00Z"}`))
	}))
	defer server.Close()

	// Setup client
	restyClient := resty.New().SetBaseURL(server.URL)
	client := NewClient(restyClient)

	// Execute test with short monitoring period
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	client.Monitor(ctx, 20*time.Millisecond, callback)

	// Verify callback was called at least once (initial check)
	assert.GreaterOrEqual(t, callbackCount, 1)
	assert.NoError(t, lastError)
	assert.NotNil(t, lastResponse)
	assert.Equal(t, types.HealthStatusHealthy, lastResponse.Status)
}

func TestClient_ContextPropagation(t *testing.T) {
	// Create test server that verifies context
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify that context headers are passed
		assert.NotEmpty(t, r.Header.Get("User-Agent"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "timestamp": "2024-01-15T12:00:00Z"}`))
	}))
	defer server.Close()

	// Setup client
	restyClient := resty.New().SetBaseURL(server.URL)
	client := NewClient(restyClient)

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.Check(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestClient_ErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse string
		serverStatus   int
		contentType    string
		expectError    bool
		errorContains  string
	}{
		{
			name:           "invalid JSON response",
			serverResponse: `{"status": "healthy", invalid json`,
			serverStatus:   http.StatusOK,
			contentType:    "application/json",
			expectError:    true,
			errorContains:  "health check request failed",
		},
		{
			name:           "empty response",
			serverResponse: "",
			serverStatus:   http.StatusOK,
			expectError:    false, // Should handle gracefully
		},
		{
			name:          "network timeout simulation",
			serverStatus:  http.StatusRequestTimeout,
			expectError:   true,
			errorContains: "health check request failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.contentType != "" {
					w.Header().Set("Content-Type", tt.contentType)
				}
				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != "" {
					w.Write([]byte(tt.serverResponse))
				}
			}))
			defer server.Close()

			// Setup client
			restyClient := resty.New().SetBaseURL(server.URL)
			client := NewClient(restyClient)

			// Execute test
			ctx := context.Background()
			response, err := client.Check(ctx)

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
			}
		})
	}
}