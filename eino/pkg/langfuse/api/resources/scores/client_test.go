package scores

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"eino/pkg/langfuse/api/resources/scores/types"
	commonTypes "eino/pkg/langfuse/api/resources/commons/types"
	commonErrors "eino/pkg/langfuse/api/resources/commons/errors"
)

func TestNewClient(t *testing.T) {
	restyClient := resty.New()

	client := NewClient(restyClient)

	assert.NotNil(t, client)
	assert.Equal(t, restyClient, client.client)
}

func TestClient_Create(t *testing.T) {
	tests := []struct {
		name           string
		request        *types.CreateScoreRequest
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
	}{
		{
			name: "successful create numeric score",
			request: &types.CreateScoreRequest{
				Name:     "accuracy",
				TraceID:  stringPtr("trace-123"),
				Value:    0.95,
				DataType: commonTypes.ScoreDataTypeNumeric,
				Source:   stringPtr("manual"),
				Comment:  stringPtr("Excellent accuracy"),
			},
			serverResponse: `{
				"id": "score-123",
				"name": "accuracy",
				"value": 0.95,
				"traceId": "trace-123",
				"dataType": "NUMERIC",
				"timestamp": "2024-01-15T12:00:00Z"
			}`,
			serverStatus: http.StatusCreated,
			expectError:  false,
		},
		{
			name: "successful create boolean score",
			request: &types.CreateScoreRequest{
				Name:          "is_correct",
				ObservationID: stringPtr("obs-456"),
				Value:         true,
				DataType:      commonTypes.ScoreDataTypeBoolean,
			},
			serverResponse: `{
				"id": "score-456",
				"name": "is_correct",
				"value": true,
				"observationId": "obs-456",
				"dataType": "BOOLEAN"
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
			request:       &types.CreateScoreRequest{Name: "test", Value: 1.0, DataType: commonTypes.ScoreDataTypeNumeric},
			serverStatus:  http.StatusBadRequest,
			expectError:   true,
			errorContains: "failed to create score",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.request != nil {
					assert.Equal(t, "POST", r.Method)
					assert.Equal(t, "/api/public/scores", r.URL.Path)
					assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
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
				assert.Equal(t, tt.request.Name, response.Name)
			}
		})
	}
}

func TestClient_List(t *testing.T) {
	tests := []struct {
		name           string
		request        *types.GetScoresRequest
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
		verifyRequest  func(t *testing.T, r *http.Request)
	}{
		{
			name: "successful list with filters",
			request: &types.GetScoresRequest{
				ProjectID:     "project-123",
				Page:          intPtr(1),
				Limit:         intPtr(10),
				TraceID:       stringPtr("trace-456"),
				ObservationID: stringPtr("obs-789"),
				Name:          stringPtr("accuracy"),
				DataType:      &commonTypes.ScoreDataTypeNumeric,
				ConfigID:      stringPtr("config-123"),
				UserID:        stringPtr("user-456"),
				Source:        stringPtr("manual"),
				FromTimestamp: timePtr(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
				ToTimestamp:   timePtr(time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)),
			},
			serverResponse: `{
				"data": [
					{
						"id": "score-1",
						"name": "accuracy",
						"value": 0.95,
						"dataType": "NUMERIC",
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
				assert.Equal(t, "/api/public/scores", r.URL.Path)
				
				query := r.URL.Query()
				assert.Equal(t, "project-123", query.Get("projectId"))
				assert.Equal(t, "1", query.Get("page"))
				assert.Equal(t, "10", query.Get("limit"))
				assert.Equal(t, "trace-456", query.Get("traceId"))
				assert.Equal(t, "obs-789", query.Get("observationId"))
				assert.Equal(t, "accuracy", query.Get("name"))
				assert.Equal(t, "NUMERIC", query.Get("dataType"))
				assert.Equal(t, "config-123", query.Get("configId"))
				assert.Equal(t, "user-456", query.Get("userId"))
				assert.Equal(t, "manual", query.Get("source"))
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
			request:       &types.GetScoresRequest{},
			serverStatus:  http.StatusInternalServerError,
			expectError:   true,
			errorContains: "failed to list scores",
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
		scoreID        string
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
	}{
		{
			name:    "successful get",
			scoreID: "score-123",
			serverResponse: `{
				"id": "score-123",
				"name": "accuracy",
				"value": 0.95,
				"dataType": "NUMERIC",
				"timestamp": "2024-01-15T12:00:00Z"
			}`,
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:          "empty score ID",
			scoreID:       "",
			expectError:   true,
			errorContains: "score ID cannot be empty",
		},
		{
			name:          "score not found",
			scoreID:       "nonexistent",
			serverStatus:  http.StatusNotFound,
			expectError:   true,
			errorContains: "failed to get score",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.scoreID != "" {
					assert.Equal(t, "GET", r.Method)
					assert.Equal(t, fmt.Sprintf("/api/public/scores/%s", tt.scoreID), r.URL.Path)
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
			response, err := client.Get(ctx, tt.scoreID)

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
				assert.Equal(t, tt.scoreID, response.ID)
			}
		})
	}
}

func TestClient_Delete(t *testing.T) {
	tests := []struct {
		name          string
		scoreID       string
		serverStatus  int
		expectError   bool
		errorContains string
	}{
		{
			name:         "successful delete",
			scoreID:      "score-123",
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:          "empty score ID",
			scoreID:       "",
			expectError:   true,
			errorContains: "score ID cannot be empty",
		},
		{
			name:          "score not found",
			scoreID:       "nonexistent",
			serverStatus:  http.StatusNotFound,
			expectError:   true,
			errorContains: "failed to delete score",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.scoreID != "" {
					assert.Equal(t, "DELETE", r.Method)
					assert.Equal(t, fmt.Sprintf("/api/public/scores/%s", tt.scoreID), r.URL.Path)
				}

				w.WriteHeader(tt.serverStatus)
			}))
			defer server.Close()

			// Setup client
			restyClient := resty.New().SetBaseURL(server.URL)
			client := NewClient(restyClient)

			// Execute test
			ctx := context.Background()
			err := client.Delete(ctx, tt.scoreID)

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

func TestClient_GetAggregation(t *testing.T) {
	tests := []struct {
		name           string
		request        *types.GetScoreAggregationRequest
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
	}{
		{
			name: "successful aggregation",
			request: &types.GetScoreAggregationRequest{
				ProjectID: "project-123",
				TraceID:   stringPtr("trace-456"),
				Name:      stringPtr("accuracy"),
				GroupBy:   []string{"name", "dataType"},
			},
			serverResponse: `{
				"aggregations": [
					{
						"groupBy": {"name": "accuracy", "dataType": "NUMERIC"},
						"count": 10,
						"average": 0.85,
						"min": 0.6,
						"max": 0.95
					}
				]
			}`,
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:           "successful aggregation with nil request",
			request:        nil,
			serverResponse: `{"aggregations": []}`,
			serverStatus:   http.StatusOK,
			expectError:    false,
		},
		{
			name:          "server error",
			request:       &types.GetScoreAggregationRequest{},
			serverStatus:  http.StatusInternalServerError,
			expectError:   true,
			errorContains: "failed to get score aggregation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/api/public/scores/aggregation", r.URL.Path)

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
			response, err := client.GetAggregation(ctx, tt.request)

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
		request        *types.GetScoreStatsRequest
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
	}{
		{
			name: "successful stats retrieval",
			request: &types.GetScoreStatsRequest{
				ProjectID: "project-123",
				TraceID:   stringPtr("trace-456"),
			},
			serverResponse: `{
				"totalCount": 100,
				"averageValue": 0.85,
				"minValue": 0.1,
				"maxValue": 1.0,
				"distribution": {"NUMERIC": 80, "BOOLEAN": 20}
			}`,
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:           "successful stats with nil request",
			request:        nil,
			serverResponse: `{"totalCount": 0, "averageValue": 0}`,
			serverStatus:   http.StatusOK,
			expectError:    false,
		},
		{
			name:          "server error",
			request:       &types.GetScoreStatsRequest{},
			serverStatus:  http.StatusInternalServerError,
			expectError:   true,
			errorContains: "failed to get score stats",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/api/public/scores/stats", r.URL.Path)

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
		request       *types.PaginatedScoresRequest
		expectError   bool
		errorContains string
	}{
		{
			name: "successful paginated request",
			request: &types.PaginatedScoresRequest{
				ProjectID: "project-123",
				Page:      1,
				Limit:     10,
				Filter: &types.ScoreFilter{
					TraceIDs:      []string{"trace-123"},
					ObservationIDs: []string{"obs-456"},
					Names:         []string{"accuracy"},
					DataTypes:     []commonTypes.ScoreDataType{commonTypes.ScoreDataTypeNumeric},
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

func TestClient_ListByTrace(t *testing.T) {
	tests := []struct {
		name          string
		traceID       string
		limit         int
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful list by trace",
			traceID:     "trace-123",
			limit:       10,
			expectError: false,
		},
		{
			name:          "empty trace ID",
			traceID:       "",
			limit:         10,
			expectError:   true,
			errorContains: "trace ID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.traceID != "" {
					assert.Equal(t, tt.traceID, r.URL.Query().Get("traceId"))
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
			response, err := client.ListByTrace(ctx, tt.traceID, tt.limit)

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

func TestClient_ListByObservation(t *testing.T) {
	tests := []struct {
		name           string
		observationID  string
		limit          int
		expectError    bool
		errorContains  string
	}{
		{
			name:          "successful list by observation",
			observationID: "obs-123",
			limit:         10,
			expectError:   false,
		},
		{
			name:          "empty observation ID",
			observationID: "",
			limit:         10,
			expectError:   true,
			errorContains: "observation ID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.observationID != "" {
					assert.Equal(t, tt.observationID, r.URL.Query().Get("observationId"))
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
			response, err := client.ListByObservation(ctx, tt.observationID, tt.limit)

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

func TestClient_CreateNumeric(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/public/scores", r.URL.Path)

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": "score-123", "name": "accuracy", "value": 0.95}`))
	}))
	defer server.Close()

	// Setup client
	restyClient := resty.New().SetBaseURL(server.URL)
	client := NewClient(restyClient)

	// Execute test
	ctx := context.Background()
	response, err := client.CreateNumeric(ctx, "trace-123", "accuracy", 0.95)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "accuracy", response.Name)
}

func TestClient_CreateBoolean(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/public/scores", r.URL.Path)

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": "score-456", "name": "is_correct", "value": true}`))
	}))
	defer server.Close()

	// Setup client
	restyClient := resty.New().SetBaseURL(server.URL)
	client := NewClient(restyClient)

	// Execute test
	ctx := context.Background()
	response, err := client.CreateBoolean(ctx, "trace-123", "is_correct", true)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "is_correct", response.Name)
}

func TestClient_CreateCategorical(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/public/scores", r.URL.Path)

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": "score-789", "name": "quality", "value": "excellent"}`))
	}))
	defer server.Close()

	// Setup client
	restyClient := resty.New().SetBaseURL(server.URL)
	client := NewClient(restyClient)

	// Execute test
	ctx := context.Background()
	response, err := client.CreateCategorical(ctx, "trace-123", "quality", "excellent")

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "quality", response.Name)
}

func TestClient_Exists(t *testing.T) {
	tests := []struct {
		name          string
		scoreID       string
		serverStatus  int
		expectedExist bool
		expectError   bool
		errorContains string
	}{
		{
			name:          "score exists",
			scoreID:       "existing-score",
			serverStatus:  http.StatusOK,
			expectedExist: true,
			expectError:   false,
		},
		{
			name:          "score does not exist",
			scoreID:       "nonexistent-score",
			serverStatus:  http.StatusNotFound,
			expectedExist: false,
			expectError:   false,
		},
		{
			name:          "empty score ID",
			scoreID:       "",
			expectError:   true,
			errorContains: "score ID cannot be empty",
		},
		{
			name:          "server error",
			scoreID:       "score-123",
			serverStatus:  http.StatusInternalServerError,
			expectError:   true,
			errorContains: "failed to get score",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
				if tt.serverStatus == http.StatusOK {
					w.Write([]byte(`{"id": "` + tt.scoreID + `", "name": "test"}`))
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
			exists, err := client.Exists(ctx, tt.scoreID)

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

func TestClient_ContextPropagation(t *testing.T) {
	// Create test server that verifies context
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify that context headers are passed
		assert.NotEmpty(t, r.Header.Get("User-Agent"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "score-123", "name": "test"}`))
	}))
	defer server.Close()

	// Setup client
	restyClient := resty.New().SetBaseURL(server.URL)
	client := NewClient(restyClient)

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.Get(ctx, "score-123")
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