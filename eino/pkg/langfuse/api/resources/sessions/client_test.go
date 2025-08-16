package sessions

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"eino/pkg/langfuse/api/resources/sessions/types"
	commonTypes "eino/pkg/langfuse/api/resources/commons/types"
	commonErrors "eino/pkg/langfuse/api/resources/commons/errors"
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
		request        *types.GetSessionsRequest
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
		verifyRequest  func(t *testing.T, r *http.Request)
	}{
		{
			name: "successful list with filters",
			request: &types.GetSessionsRequest{
				ProjectID:     "project-123",
				Page:          intPtr(1),
				Limit:         intPtr(10),
				UserID:        stringPtr("user-456"),
				FromTimestamp: timePtr(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
				ToTimestamp:   timePtr(time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)),
				OrderBy:       stringPtr("timestamp"),
			},
			serverResponse: `{
				"data": [
					{
						"id": "session-1",
						"userId": "user-456",
						"createdAt": "2024-01-15T12:00:00Z"
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
				assert.Equal(t, "/api/public/sessions", r.URL.Path)
				
				query := r.URL.Query()
				assert.Equal(t, "project-123", query.Get("projectId"))
				assert.Equal(t, "1", query.Get("page"))
				assert.Equal(t, "10", query.Get("limit"))
				assert.Equal(t, "user-456", query.Get("userId"))
				assert.Equal(t, "timestamp", query.Get("orderBy"))
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
			request:       &types.GetSessionsRequest{},
			serverStatus:  http.StatusInternalServerError,
			expectError:   true,
			errorContains: "failed to list sessions",
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
		sessionID      string
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
	}{
		{
			name:      "successful get",
			sessionID: "session-123",
			serverResponse: `{
				"id": "session-123",
				"userId": "user-456",
				"createdAt": "2024-01-15T12:00:00Z"
			}`,
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:          "empty session ID",
			sessionID:     "",
			expectError:   true,
			errorContains: "session ID cannot be empty",
		},
		{
			name:          "session not found",
			sessionID:     "nonexistent",
			serverStatus:  http.StatusNotFound,
			expectError:   true,
			errorContains: "failed to get session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.sessionID != "" {
					assert.Equal(t, "GET", r.Method)
					assert.Equal(t, fmt.Sprintf("/api/public/sessions/%s", tt.sessionID), r.URL.Path)
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
			response, err := client.Get(ctx, tt.sessionID)

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
				assert.Equal(t, tt.sessionID, response.ID)
			}
		})
	}
}

func TestClient_GetWithTraces(t *testing.T) {
	tests := []struct {
		name           string
		sessionID      string
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
	}{
		{
			name:      "successful get with traces",
			sessionID: "session-123",
			serverResponse: `{
				"id": "session-123",
				"userId": "user-456",
				"createdAt": "2024-01-15T12:00:00Z",
				"traces": [
					{
						"id": "trace-1",
						"name": "test-trace",
						"timestamp": "2024-01-15T12:00:00Z"
					}
				]
			}`,
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:          "empty session ID",
			sessionID:     "",
			expectError:   true,
			errorContains: "session ID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.sessionID != "" {
					assert.Equal(t, "GET", r.Method)
					assert.Equal(t, fmt.Sprintf("/api/public/sessions/%s", tt.sessionID), r.URL.Path)
					assert.Equal(t, "true", r.URL.Query().Get("includeTraces"))
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
			response, err := client.GetWithTraces(ctx, tt.sessionID)

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
				assert.Equal(t, tt.sessionID, response.ID)
				assert.NotNil(t, response.Traces)
			}
		})
	}
}

func TestClient_Create(t *testing.T) {
	tests := []struct {
		name           string
		request        *types.CreateSessionRequest
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
	}{
		{
			name: "successful create with user",
			request: &types.CreateSessionRequest{
				UserID: stringPtr("user-123"),
			},
			serverResponse: `{
				"id": "session-123",
				"userId": "user-123",
				"createdAt": "2024-01-15T12:00:00Z"
			}`,
			serverStatus: http.StatusCreated,
			expectError:  false,
		},
		{
			name:    "successful create with nil request",
			request: nil,
			serverResponse: `{
				"id": "session-456",
				"createdAt": "2024-01-15T12:00:00Z"
			}`,
			serverStatus: http.StatusCreated,
			expectError:  false,
		},
		{
			name:          "server error",
			request:       &types.CreateSessionRequest{},
			serverStatus:  http.StatusBadRequest,
			expectError:   true,
			errorContains: "failed to create session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/api/public/sessions", r.URL.Path)

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
			}
		})
	}
}

func TestClient_Update(t *testing.T) {
	tests := []struct {
		name           string
		request        *types.UpdateSessionRequest
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
	}{
		{
			name: "successful update",
			request: &types.UpdateSessionRequest{
				SessionID: "session-123",
				UserID:    stringPtr("user-456"),
			},
			serverResponse: `{
				"id": "session-123",
				"userId": "user-456",
				"updatedAt": "2024-01-15T12:00:00Z"
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
			request:       &types.UpdateSessionRequest{SessionID: "session-123"},
			serverStatus:  http.StatusNotFound,
			expectError:   true,
			errorContains: "failed to update session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.request != nil {
					assert.Equal(t, "PATCH", r.Method)
					assert.Equal(t, fmt.Sprintf("/api/public/sessions/%s", tt.request.SessionID), r.URL.Path)
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
				assert.Equal(t, "session-123", response.ID)
			}
		})
	}
}

func TestClient_Delete(t *testing.T) {
	tests := []struct {
		name          string
		sessionID     string
		serverStatus  int
		expectError   bool
		errorContains string
	}{
		{
			name:         "successful delete",
			sessionID:    "session-123",
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:          "empty session ID",
			sessionID:     "",
			expectError:   true,
			errorContains: "session ID cannot be empty",
		},
		{
			name:          "session not found",
			sessionID:     "nonexistent",
			serverStatus:  http.StatusNotFound,
			expectError:   true,
			errorContains: "failed to delete session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.sessionID != "" {
					assert.Equal(t, "DELETE", r.Method)
					assert.Equal(t, fmt.Sprintf("/api/public/sessions/%s", tt.sessionID), r.URL.Path)
				}

				w.WriteHeader(tt.serverStatus)
			}))
			defer server.Close()

			// Setup client
			restyClient := resty.New().SetBaseURL(server.URL)
			client := NewClient(restyClient)

			// Execute test
			ctx := context.Background()
			err := client.Delete(ctx, tt.sessionID)

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

func TestClient_GetStats(t *testing.T) {
	tests := []struct {
		name           string
		request        *types.GetSessionStatsRequest
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
	}{
		{
			name: "successful stats retrieval",
			request: &types.GetSessionStatsRequest{
				ProjectID: "project-123",
				UserID:    stringPtr("user-456"),
			},
			serverResponse: `{
				"totalCount": 100,
				"uniqueUsers": 10,
				"averageDuration": 3600,
				"userDistribution": {"user-123": 25, "user-456": 75}
			}`,
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:           "successful stats with nil request",
			request:        nil,
			serverResponse: `{"totalCount": 0, "uniqueUsers": 0}`,
			serverStatus:   http.StatusOK,
			expectError:    false,
		},
		{
			name:          "server error",
			request:       &types.GetSessionStatsRequest{},
			serverStatus:  http.StatusInternalServerError,
			expectError:   true,
			errorContains: "failed to get session stats",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/api/public/sessions/stats", r.URL.Path)

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
		request       *types.PaginatedSessionsRequest
		expectError   bool
		errorContains string
	}{
		{
			name: "successful paginated request",
			request: &types.PaginatedSessionsRequest{
				ProjectID: "project-123",
				Page:      1,
				Limit:     10,
				SortOrder: "timestamp",
				Filter: &types.SessionFilter{
					UserIDs: []string{"user-123"},
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

func TestClient_Exists(t *testing.T) {
	tests := []struct {
		name          string
		sessionID     string
		serverStatus  int
		expectedExist bool
		expectError   bool
		errorContains string
	}{
		{
			name:          "session exists",
			sessionID:     "existing-session",
			serverStatus:  http.StatusOK,
			expectedExist: true,
			expectError:   false,
		},
		{
			name:          "session does not exist",
			sessionID:     "nonexistent-session",
			serverStatus:  http.StatusNotFound,
			expectedExist: false,
			expectError:   false,
		},
		{
			name:          "empty session ID",
			sessionID:     "",
			expectError:   true,
			errorContains: "session ID cannot be empty",
		},
		{
			name:          "server error",
			sessionID:     "session-123",
			serverStatus:  http.StatusInternalServerError,
			expectError:   true,
			errorContains: "failed to get session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
				if tt.serverStatus == http.StatusOK {
					w.Write([]byte(`{"id": "` + tt.sessionID + `", "userId": "test"}`))
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
			exists, err := client.Exists(ctx, tt.sessionID)

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

func TestClient_CreateForUser(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful create for user",
			userID:      "user-123",
			expectError: false,
		},
		{
			name:          "empty user ID",
			userID:        "",
			expectError:   true,
			errorContains: "user ID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(`{"id": "session-123", "userId": "` + tt.userID + `"}`))
			}))
			defer server.Close()

			// Setup client
			restyClient := resty.New().SetBaseURL(server.URL)
			client := NewClient(restyClient)

			// Execute test
			ctx := context.Background()
			response, err := client.CreateForUser(ctx, tt.userID)

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

func TestClient_UpdateUserID(t *testing.T) {
	tests := []struct {
		name          string
		sessionID     string
		userID        string
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful update user ID",
			sessionID:   "session-123",
			userID:      "user-456",
			expectError: false,
		},
		{
			name:          "empty session ID",
			sessionID:     "",
			userID:        "user-456",
			expectError:   true,
			errorContains: "session ID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"id": "` + tt.sessionID + `", "userId": "` + tt.userID + `"}`))
			}))
			defer server.Close()

			// Setup client
			restyClient := resty.New().SetBaseURL(server.URL)
			client := NewClient(restyClient)

			// Execute test
			ctx := context.Background()
			response, err := client.UpdateUserID(ctx, tt.sessionID, tt.userID)

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

func TestClient_GetTraceCount(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "true", r.URL.Query().Get("includeTraces"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "session-123",
			"userId": "user-456",
			"traces": [
				{"id": "trace-1"},
				{"id": "trace-2"},
				{"id": "trace-3"}
			]
		}`))
	}))
	defer server.Close()

	// Setup client
	restyClient := resty.New().SetBaseURL(server.URL)
	client := NewClient(restyClient)

	// Execute test
	ctx := context.Background()
	count, err := client.GetTraceCount(ctx, "session-123")

	assert.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestClient_ContextPropagation(t *testing.T) {
	// Create test server that verifies context
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify that context headers are passed
		assert.NotEmpty(t, r.Header.Get("User-Agent"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "session-123", "userId": "test"}`))
	}))
	defer server.Close()

	// Setup client
	restyClient := resty.New().SetBaseURL(server.URL)
	client := NewClient(restyClient)

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.Get(ctx, "session-123")
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