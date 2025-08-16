package traces

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
	"eino/pkg/langfuse/api/resources/traces/types"
	commonTypes "eino/pkg/langfuse/api/resources/commons/types"
	commonErrors "eino/pkg/langfuse/api/resources/commons/errors"
	paginationTypes "eino/pkg/langfuse/api/resources/utils/pagination/types"
)

func TestNewClient(t *testing.T) {
	restyClient := resty.New()

	client := NewClient(restyClient)

	assert.NotNil(t, client)
	assert.Equal(t, restyClient, client.client)
}

func TestClient_List(t *testing.T) {
	tests := []struct {
		name           string
		request        *types.GetTracesRequest
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
		verifyRequest  func(t *testing.T, r *http.Request)
	}{
		{
			name: "successful list with filters",
			request: &types.GetTracesRequest{
				ProjectID:     "project-123",
				Page:          intPtr(1),
				Limit:         intPtr(10),
				UserID:        stringPtr("user-456"),
				Name:          stringPtr("test-trace"),
				SessionID:     stringPtr("session-789"),
				FromTimestamp: timePtr(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
				ToTimestamp:   timePtr(time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)),
				OrderBy:       stringPtr("timestamp"),
				Tags:          []string{"tag1", "tag2"},
			},
			serverResponse: `{
				"data": [
					{
						"id": "trace-1",
						"name": "test-trace",
						"timestamp": "2024-01-15T12:00:00Z"
					}
				],
				"meta": {
					"page": 1,
					"limit": 10,
					"totalItems": 1,
					"totalPages": 1
				}
			}`,
			serverStatus: http.StatusOK,
			expectError:  false,
			verifyRequest: func(t *testing.T, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/api/public/traces", r.URL.Path)
				
				query := r.URL.Query()
				assert.Equal(t, "project-123", query.Get("projectId"))
				assert.Equal(t, "1", query.Get("page"))
				assert.Equal(t, "10", query.Get("limit"))
				assert.Equal(t, "user-456", query.Get("userId"))
				assert.Equal(t, "test-trace", query.Get("name"))
				assert.Equal(t, "session-789", query.Get("sessionId"))
				assert.Equal(t, "timestamp", query.Get("orderBy"))
				assert.Equal(t, "tag1,tag2", query.Get("tags"))
				assert.Contains(t, query.Get("fromTimestamp"), "2024-01-01T00:00:00")
				assert.Contains(t, query.Get("toTimestamp"), "2024-01-31T23:59:59")
			},
		},
		{
			name:           "successful list with nil request",
			request:        nil,
			serverResponse: `{"data": [], "meta": {"page": 1, "limit": 10, "totalItems": 0, "totalPages": 0}}`,
			serverStatus:   http.StatusOK,
			expectError:    false,
		},
		{
			name:          "server error",
			request:       &types.GetTracesRequest{},
			serverStatus:  http.StatusInternalServerError,
			expectError:   true,
			errorContains: "failed to list traces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.verifyRequest != nil {
					tt.verifyRequest(t, r)
				}

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
			response, err := client.List(ctx, tt.request)

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
				assert.NotNil(t, response.Data)
				assert.NotNil(t, response.Meta)
			}
		})
	}
}

func TestClient_Get(t *testing.T) {
	tests := []struct {
		name           string
		traceID        string
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
	}{
		{
			name:    "successful get",
			traceID: "trace-123",
			serverResponse: `{
				"id": "trace-123",
				"name": "test-trace",
				"timestamp": "2024-01-15T12:00:00Z",
				"userId": "user-456"
			}`,
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:          "empty trace ID",
			traceID:       "",
			expectError:   true,
			errorContains: "trace ID cannot be empty",
		},
		{
			name:          "trace not found",
			traceID:       "nonexistent",
			serverStatus:  http.StatusNotFound,
			expectError:   true,
			errorContains: "failed to get trace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.traceID != "" {
					assert.Equal(t, "GET", r.Method)
					assert.Equal(t, fmt.Sprintf("/api/public/traces/%s", tt.traceID), r.URL.Path)
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
			response, err := client.Get(ctx, tt.traceID)

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
				assert.Equal(t, tt.traceID, response.ID)
			}
		})
	}
}

func TestClient_GetWithObservations(t *testing.T) {
	tests := []struct {
		name           string
		traceID        string
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
	}{
		{
			name:    "successful get with observations",
			traceID: "trace-123",
			serverResponse: `{
				"id": "trace-123",
				"name": "test-trace",
				"timestamp": "2024-01-15T12:00:00Z",
				"observations": [
					{
						"id": "obs-1",
						"type": "GENERATION",
						"name": "test-observation"
					}
				]
			}`,
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:          "empty trace ID",
			traceID:       "",
			expectError:   true,
			errorContains: "trace ID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.traceID != "" {
					assert.Equal(t, "GET", r.Method)
					assert.Equal(t, fmt.Sprintf("/api/public/traces/%s", tt.traceID), r.URL.Path)
					assert.Equal(t, "true", r.URL.Query().Get("includeObservations"))
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
			response, err := client.GetWithObservations(ctx, tt.traceID)

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
				assert.Equal(t, tt.traceID, response.ID)
				assert.NotNil(t, response.Observations)
			}
		})
	}
}

func TestClient_Create(t *testing.T) {
	tests := []struct {
		name           string
		request        *types.CreateTraceRequest
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
	}{
		{
			name: "successful create",
			request: &types.CreateTraceRequest{
				Name:      "test-trace",
				UserID:    stringPtr("user-123"),
				SessionID: stringPtr("session-456"),
				Input:     map[string]interface{}{"query": "test"},
				Output:    map[string]interface{}{"result": "success"},
				Metadata:  map[string]interface{}{"version": "1.0"},
				Tags:      []string{"test", "api"},
			},
			serverResponse: `{
				"id": "trace-123",
				"name": "test-trace",
				"timestamp": "2024-01-15T12:00:00Z"
			}`,
			serverStatus: http.StatusCreated,
			expectError:  false,
		},
		{
			name:          "nil request",
			request:       nil,
			expectError:   true,
			errorContains: "create request cannot be nil",
		},
		{
			name:          "server error",
			request:       &types.CreateTraceRequest{Name: "test"},
			serverStatus:  http.StatusBadRequest,
			expectError:   true,
			errorContains: "failed to create trace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.request != nil {
					assert.Equal(t, "POST", r.Method)
					assert.Equal(t, "/api/public/traces", r.URL.Path)
					assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
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
			response, err := client.Create(ctx, tt.request)

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
				assert.Equal(t, "test-trace", response.Name)
			}
		})
	}
}

func TestClient_Update(t *testing.T) {
	tests := []struct {
		name           string
		request        *types.UpdateTraceRequest
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
	}{
		{
			name: "successful update",
			request: &types.UpdateTraceRequest{
				TraceID:   "trace-123",
				Name:      stringPtr("updated-trace"),
				UserID:    stringPtr("user-123"),
				SessionID: stringPtr("session-456"),
				Output:    map[string]interface{}{"result": "updated"},
				Metadata:  map[string]interface{}{"version": "1.1"},
				Tags:      []string{"updated", "api"},
			},
			serverResponse: `{
				"id": "trace-123",
				"name": "updated-trace",
				"timestamp": "2024-01-15T12:00:00Z"
			}`,
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:          "nil request",
			request:       nil,
			expectError:   true,
			errorContains: "update request cannot be nil",
		},
		{
			name:          "server error",
			request:       &types.UpdateTraceRequest{TraceID: "trace-123", Name: stringPtr("test")},
			serverStatus:  http.StatusNotFound,
			expectError:   true,
			errorContains: "failed to update trace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.request != nil {
					assert.Equal(t, "PATCH", r.Method)
					assert.Equal(t, fmt.Sprintf("/api/public/traces/%s", tt.request.TraceID), r.URL.Path)
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
			response, err := client.Update(ctx, tt.request)

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
				assert.Equal(t, "trace-123", response.ID)
			}
		})
	}
}

func TestClient_Delete(t *testing.T) {
	tests := []struct {
		name           string
		traceID        string
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
	}{
		{
			name:           "successful delete",
			traceID:        "trace-123",
			serverResponse: `{"success": true, "message": "Trace deleted successfully"}`,
			serverStatus:   http.StatusOK,
			expectError:    false,
		},
		{
			name:          "empty trace ID",
			traceID:       "",
			expectError:   true,
			errorContains: "trace ID cannot be empty",
		},
		{
			name:          "trace not found",
			traceID:       "nonexistent",
			serverStatus:  http.StatusNotFound,
			expectError:   true,
			errorContains: "failed to delete trace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.traceID != "" {
					assert.Equal(t, "DELETE", r.Method)
					assert.Equal(t, fmt.Sprintf("/api/public/traces/%s", tt.traceID), r.URL.Path)
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
			response, err := client.Delete(ctx, tt.traceID)

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

func TestClient_GetStats(t *testing.T) {
	tests := []struct {
		name           string
		request        *types.GetTraceStatsRequest
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
	}{
		{
			name: "successful stats retrieval",
			request: &types.GetTraceStatsRequest{
				ProjectID: "project-123",
				UserID:    stringPtr("user-456"),
			},
			serverResponse: `{
				"totalCount": 100,
				"uniqueUsers": 10,
				"uniqueSessions": 25,
				"tagDistribution": {"api": 50, "test": 30}
			}`,
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:           "successful stats with nil request",
			request:        nil,
			serverResponse: `{"totalCount": 0, "uniqueUsers": 0, "uniqueSessions": 0}`,
			serverStatus:   http.StatusOK,
			expectError:    false,
		},
		{
			name:          "server error",
			request:       &types.GetTraceStatsRequest{},
			serverStatus:  http.StatusInternalServerError,
			expectError:   true,
			errorContains: "failed to get trace stats",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/api/public/traces/stats", r.URL.Path)

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
			response, err := client.GetStats(ctx, tt.request)

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

func TestClient_ListPaginated(t *testing.T) {
	tests := []struct {
		name          string
		request       *types.PaginatedTracesRequest
		expectError   bool
		errorContains string
	}{
		{
			name: "successful paginated request",
			request: &types.PaginatedTracesRequest{
				ProjectID: "project-123",
				Page:      1,
				Limit:     10,
				SortOrder: "timestamp",
				Filter: &types.TraceFilter{
					UserIDs:    []string{"user-123"},
					SessionIDs: []string{"session-456"},
					Names:      []string{"test-trace"},
					Tags:       []string{"api"},
				},
			},
			expectError: false,
		},
		{
			name:          "nil request",
			request:       nil,
			expectError:   true,
			errorContains: "paginated request cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"data": [], "meta": {"page": 1, "limit": 10, "totalItems": 0, "totalPages": 0}}`))
			}))
			defer server.Close()

			// Setup client
			restyClient := resty.New().SetBaseURL(server.URL)
			client := NewClient(restyClient)

			// Execute test
			ctx := context.Background()
			response, err := client.ListPaginated(ctx, tt.request)

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

func TestClient_Exists(t *testing.T) {
	tests := []struct {
		name          string
		traceID       string
		serverStatus  int
		expectedExist bool
		expectError   bool
		errorContains string
	}{
		{
			name:          "trace exists",
			traceID:       "existing-trace",
			serverStatus:  http.StatusOK,
			expectedExist: true,
			expectError:   false,
		},
		{
			name:          "trace does not exist",
			traceID:       "nonexistent-trace",
			serverStatus:  http.StatusNotFound,
			expectedExist: false,
			expectError:   false,
		},
		{
			name:          "empty trace ID",
			traceID:       "",
			expectError:   true,
			errorContains: "trace ID cannot be empty",
		},
		{
			name:          "server error",
			traceID:       "trace-123",
			serverStatus:  http.StatusInternalServerError,
			expectError:   true,
			errorContains: "failed to get trace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
				if tt.serverStatus == http.StatusOK {
					w.Write([]byte(`{"id": "` + tt.traceID + `", "name": "test"}`))
				} else if tt.serverStatus == http.StatusNotFound {
					w.Write([]byte(`{"error": "not found"}`))
				}
			}))
			defer server.Close()

			// Setup client
			restyClient := resty.New().SetBaseURL(server.URL)
			client := NewClient(restyClient)

			// Execute test
			ctx := context.Background()
			exists, err := client.Exists(ctx, tt.traceID)

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedExist, exists)
			}
		})
	}
}

func TestClient_ListBySession(t *testing.T) {
	tests := []struct {
		name          string
		sessionID     string
		limit         int
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful list by session",
			sessionID:   "session-123",
			limit:       10,
			expectError: false,
		},
		{
			name:          "empty session ID",
			sessionID:     "",
			limit:         10,
			expectError:   true,
			errorContains: "session ID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.sessionID != "" {
					assert.Equal(t, tt.sessionID, r.URL.Query().Get("sessionId"))
					assert.Equal(t, fmt.Sprintf("%d", tt.limit), r.URL.Query().Get("limit"))
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"data": [], "meta": {"page": 1, "limit": 10, "totalItems": 0, "totalPages": 0}}`))
			}))
			defer server.Close()

			// Setup client
			restyClient := resty.New().SetBaseURL(server.URL)
			client := NewClient(restyClient)

			// Execute test
			ctx := context.Background()
			response, err := client.ListBySession(ctx, tt.sessionID, tt.limit)

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

func TestClient_ListByUser(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		limit         int
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful list by user",
			userID:      "user-123",
			limit:       10,
			expectError: false,
		},
		{
			name:          "empty user ID",
			userID:        "",
			limit:         10,
			expectError:   true,
			errorContains: "user ID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.userID != "" {
					assert.Equal(t, tt.userID, r.URL.Query().Get("userId"))
					assert.Equal(t, fmt.Sprintf("%d", tt.limit), r.URL.Query().Get("limit"))
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"data": [], "meta": {"page": 1, "limit": 10, "totalItems": 0, "totalPages": 0}}`))
			}))
			defer server.Close()

			// Setup client
			restyClient := resty.New().SetBaseURL(server.URL)
			client := NewClient(restyClient)

			// Execute test
			ctx := context.Background()
			response, err := client.ListByUser(ctx, tt.userID, tt.limit)

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

func TestClient_ContextPropagation(t *testing.T) {
	// Create test server that verifies context
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify that context headers are passed
		assert.NotEmpty(t, r.Header.Get("User-Agent"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "trace-123", "name": "test"}`))
	}))
	defer server.Close()

	// Setup client
	restyClient := resty.New().SetBaseURL(server.URL)
	client := NewClient(restyClient)

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.Get(ctx, "trace-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

// Helper functions

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func timePtr(t time.Time) *time.Time {
	return &t
}