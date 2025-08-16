package ingestion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"eino/pkg/langfuse/api/resources/ingestion/types"
)

func TestNewClient(t *testing.T) {
	restyClient := resty.New()

	client := NewClient(restyClient)

	assert.NotNil(t, client)
	assert.Equal(t, restyClient, client.client)
}

func TestClient_Submit(t *testing.T) {
	tests := []struct {
		name           string
		request        *types.IngestionRequest
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
	}{
		{
			name: "successful submission",
			request: &types.IngestionRequest{
				Batch: []types.IngestionEvent{
					createMockIngestionEvent(),
				},
				Metadata: map[string]interface{}{
					"sdk_version": "test",
				},
			},
			serverResponse: `{"success": true, "timestamp": "2024-01-15T12:00:00Z"}`,
			serverStatus:   http.StatusOK,
			expectError:    false,
		},
		{
			name:          "nil request",
			request:       nil,
			expectError:   true,
			errorContains: "ingestion request cannot be nil",
		},
		{
			name: "server error",
			request: &types.IngestionRequest{
				Batch: []types.IngestionEvent{
					createMockIngestionEvent(),
				},
			},
			serverResponse: `{"error": "internal server error"}`,
			serverStatus:   http.StatusInternalServerError,
			expectError:    true,
			errorContains:  "failed to submit ingestion request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.serverStatus != 0 {
					w.WriteHeader(tt.serverStatus)
				}
				if tt.serverResponse != "" {
					w.Write([]byte(tt.serverResponse))
				}

				// Verify request method and path
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/api/public/ingestion", r.URL.Path)
			}))
			defer server.Close()

			// Setup client
			restyClient := resty.New().SetBaseURL(server.URL)
			client := NewClient(restyClient)

			// Execute test
			ctx := context.Background()
			response, err := client.Submit(ctx, tt.request)

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
				assert.True(t, response.Success)
			}
		})
	}
}

func TestClient_SubmitBatch(t *testing.T) {
	tests := []struct {
		name           string
		events         []types.IngestionEvent
		expectError    bool
		errorContains  string
	}{
		{
			name: "successful batch submission",
			events: []types.IngestionEvent{
				createMockIngestionEvent(),
				createMockIngestionEvent(),
			},
			expectError: false,
		},
		{
			name:          "empty batch",
			events:        []types.IngestionEvent{},
			expectError:   true,
			errorContains: "cannot submit empty batch",
		},
		{
			name:          "nil events",
			events:        nil,
			expectError:   true,
			errorContains: "cannot submit empty batch",
		},
		{
			name:          "batch size exceeds maximum",
			events:        createLargeBatch(types.MaxBatchSize + 1),
			expectError:   true,
			errorContains: "batch size",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"success": true, "timestamp": "2024-01-15T12:00:00Z"}`))
			}))
			defer server.Close()

			// Setup client
			restyClient := resty.New().SetBaseURL(server.URL)
			client := NewClient(restyClient)

			// Execute test
			ctx := context.Background()
			response, err := client.SubmitBatch(ctx, tt.events)

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

func TestClient_SubmitBatchWithMetadata(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "timestamp": "2024-01-15T12:00:00Z"}`))
	}))
	defer server.Close()

	// Setup client
	restyClient := resty.New().SetBaseURL(server.URL)
	client := NewClient(restyClient)

	events := []types.IngestionEvent{createMockIngestionEvent()}
	metadata := &types.IngestionBatchMetadata{
		SDKVersion:     "test-v1.0.0",
		SDKIntegration: "test-integration",
		ClientID:       "test-client",
	}

	ctx := context.Background()
	response, err := client.SubmitBatchWithMetadata(ctx, events, metadata)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.True(t, response.Success)
}

func TestClient_SubmitTrace(t *testing.T) {
	tests := []struct {
		name          string
		event         *types.TraceCreateEvent
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful trace submission",
			event:       createMockTraceCreateEvent(),
			expectError: false,
		},
		{
			name:          "nil trace event",
			event:         nil,
			expectError:   true,
			errorContains: "trace event cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"success": true, "timestamp": "2024-01-15T12:00:00Z"}`))
			}))
			defer server.Close()

			// Setup client
			restyClient := resty.New().SetBaseURL(server.URL)
			client := NewClient(restyClient)

			// Execute test
			ctx := context.Background()
			response, err := client.SubmitTrace(ctx, tt.event)

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

func TestClient_SubmitTraceUpdate(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "timestamp": "2024-01-15T12:00:00Z"}`))
	}))
	defer server.Close()

	// Setup client
	restyClient := resty.New().SetBaseURL(server.URL)
	client := NewClient(restyClient)

	event := createMockTraceUpdateEvent()

	ctx := context.Background()
	response, err := client.SubmitTraceUpdate(ctx, event)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.True(t, response.Success)
}

func TestClient_SubmitObservation(t *testing.T) {
	tests := []struct {
		name          string
		event         interface{}
		expectError   bool
		errorContains string
	}{
		{
			name:        "observation create event",
			event:       createMockObservationCreateEvent(),
			expectError: false,
		},
		{
			name:        "observation update event",
			event:       createMockObservationUpdateEvent(),
			expectError: false,
		},
		{
			name:        "span create event",
			event:       createMockSpanCreateEvent(),
			expectError: false,
		},
		{
			name:        "generation create event",
			event:       createMockGenerationCreateEvent(),
			expectError: false,
		},
		{
			name:          "nil event",
			event:         nil,
			expectError:   true,
			errorContains: "observation event cannot be nil",
		},
		{
			name:          "unsupported event type",
			event:         "unsupported",
			expectError:   true,
			errorContains: "unsupported observation event type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"success": true, "timestamp": "2024-01-15T12:00:00Z"}`))
			}))
			defer server.Close()

			// Setup client
			restyClient := resty.New().SetBaseURL(server.URL)
			client := NewClient(restyClient)

			// Execute test
			ctx := context.Background()
			response, err := client.SubmitObservation(ctx, tt.event)

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

func TestClient_SubmitScore(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "timestamp": "2024-01-15T12:00:00Z"}`))
	}))
	defer server.Close()

	// Setup client
	restyClient := resty.New().SetBaseURL(server.URL)
	client := NewClient(restyClient)

	event := createMockScoreCreateEvent()

	ctx := context.Background()
	response, err := client.SubmitScore(ctx, event)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.True(t, response.Success)
}

func TestClient_SubmitMultipleEvents(t *testing.T) {
	tests := []struct {
		name          string
		events        []interface{}
		expectError   bool
		errorContains string
	}{
		{
			name: "mixed valid events",
			events: []interface{}{
				createMockTraceCreateEvent(),
				createMockObservationCreateEvent(),
				createMockScoreCreateEvent(),
			},
			expectError: false,
		},
		{
			name:          "empty events list",
			events:        []interface{}{},
			expectError:   true,
			errorContains: "cannot submit empty events list",
		},
		{
			name: "event with nil",
			events: []interface{}{
				createMockTraceCreateEvent(),
				nil,
			},
			expectError:   true,
			errorContains: "event at index 1 cannot be nil",
		},
		{
			name: "unsupported event type",
			events: []interface{}{
				createMockTraceCreateEvent(),
				"unsupported",
			},
			expectError:   true,
			errorContains: "unsupported event type at index 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"success": true, "timestamp": "2024-01-15T12:00:00Z"}`))
			}))
			defer server.Close()

			// Setup client
			restyClient := resty.New().SetBaseURL(server.URL)
			client := NewClient(restyClient)

			// Execute test
			ctx := context.Background()
			response, err := client.SubmitMultipleEvents(ctx, tt.events)

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

func TestClient_SubmitWithRetry(t *testing.T) {
	tests := []struct {
		name           string
		serverBehavior func(callCount *int) http.HandlerFunc
		maxRetries     int
		expectError    bool
		errorContains  string
	}{
		{
			name: "success on first try",
			serverBehavior: func(callCount *int) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					*callCount++
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"success": true, "timestamp": "2024-01-15T12:00:00Z"}`))
				}
			},
			maxRetries:  3,
			expectError: false,
		},
		{
			name: "success on second try",
			serverBehavior: func(callCount *int) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					*callCount++
					if *callCount == 1 {
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"success": true, "timestamp": "2024-01-15T12:00:00Z"}`))
				}
			},
			maxRetries:  3,
			expectError: false,
		},
		{
			name: "max retries exceeded",
			serverBehavior: func(callCount *int) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					*callCount++
					w.WriteHeader(http.StatusInternalServerError)
				}
			},
			maxRetries:    2,
			expectError:   true,
			errorContains: "max retries exceeded",
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

			request := &types.IngestionRequest{
				Batch: []types.IngestionEvent{createMockIngestionEvent()},
			}

			// Execute test
			ctx := context.Background()
			response, err := client.SubmitWithRetry(ctx, request, tt.maxRetries, 10*time.Millisecond)

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

func TestClient_Health(t *testing.T) {
	tests := []struct {
		name           string
		serverStatus   int
		serverResponse string
		expectError    bool
		errorContains  string
	}{
		{
			name:           "healthy service",
			serverStatus:   http.StatusOK,
			serverResponse: `{"status": "ok"}`,
			expectError:    false,
		},
		{
			name:          "unhealthy service",
			serverStatus:  http.StatusServiceUnavailable,
			expectError:   true,
			errorContains: "ingestion health check failed",
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

				// Verify request
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/api/public/health", r.URL.Path)
			}))
			defer server.Close()

			// Setup client
			restyClient := resty.New().SetBaseURL(server.URL)
			client := NewClient(restyClient)

			// Execute test
			ctx := context.Background()
			err := client.Health(ctx)

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

func TestClient_ContextPropagation(t *testing.T) {
	// Create test server that verifies context
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify that context headers are passed
		assert.NotEmpty(t, r.Header.Get("User-Agent"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "timestamp": "2024-01-15T12:00:00Z"}`))
	}))
	defer server.Close()

	// Setup client
	restyClient := resty.New().SetBaseURL(server.URL)
	client := NewClient(restyClient)

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	request := &types.IngestionRequest{
		Batch: []types.IngestionEvent{createMockIngestionEvent()},
	}

	_, err := client.Submit(ctx, request)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

// Helper functions for creating mock data

func createMockIngestionEvent() types.IngestionEvent {
	return types.IngestionEvent{
		ID:        "test-event-id",
		Type:      types.EventTypeTraceCreate,
		Timestamp: time.Now(),
		Body: map[string]interface{}{
			"id":   "test-trace-id",
			"name": "test-trace",
		},
	}
}

func createLargeBatch(size int) []types.IngestionEvent {
	events := make([]types.IngestionEvent, size)
	for i := 0; i < size; i++ {
		events[i] = createMockIngestionEvent()
	}
	return events
}

func createMockTraceCreateEvent() *types.TraceCreateEvent {
	return &types.TraceCreateEvent{
		ID:        "test-trace-id",
		Name:      stringPtr("test-trace"),
		Timestamp: time.Now(),
		Input:     json.RawMessage(`{"input": "test"}`),
		Output:    json.RawMessage(`{"output": "result"}`),
	}
}

func createMockTraceUpdateEvent() *types.TraceUpdateEvent {
	return &types.TraceUpdateEvent{
		ID:        "test-trace-id",
		Name:      stringPtr("updated-trace"),
		Timestamp: time.Now(),
		Output:    json.RawMessage(`{"output": "updated-result"}`),
	}
}

func createMockObservationCreateEvent() *types.ObservationCreateEvent {
	return &types.ObservationCreateEvent{
		ID:        "test-observation-id",
		TraceID:   "test-trace-id",
		Type:      types.ObservationTypeGeneration,
		Name:      stringPtr("test-observation"),
		Timestamp: time.Now(),
	}
}

func createMockObservationUpdateEvent() *types.ObservationUpdateEvent {
	return &types.ObservationUpdateEvent{
		ID:        "test-observation-id",
		TraceID:   "test-trace-id",
		Type:      types.ObservationTypeGeneration,
		Name:      stringPtr("updated-observation"),
		Timestamp: time.Now(),
	}
}

func createMockSpanCreateEvent() *types.SpanCreateEvent {
	return &types.SpanCreateEvent{
		ID:        "test-span-id",
		TraceID:   "test-trace-id",
		Name:      stringPtr("test-span"),
		Timestamp: time.Now(),
	}
}

func createMockGenerationCreateEvent() *types.GenerationCreateEvent {
	return &types.GenerationCreateEvent{
		ID:        "test-generation-id",
		TraceID:   "test-trace-id",
		Name:      stringPtr("test-generation"),
		Timestamp: time.Now(),
	}
}

func createMockScoreCreateEvent() *types.ScoreCreateEvent {
	return &types.ScoreCreateEvent{
		ID:          "test-score-id",
		TraceID:     stringPtr("test-trace-id"),
		Name:        "test-score",
		Value:       1.0,
		DataType:    types.ScoreDataTypeNumeric,
		Timestamp:   time.Now(),
	}
}

// Helper function for string pointers
func stringPtr(s string) *string {
	return &s
}