package models

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"eino/pkg/langfuse/api/resources/models/types"
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
		request        *types.GetModelsRequest
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
		verifyRequest  func(t *testing.T, r *http.Request)
	}{
		{
			name: "successful list with filters",
			request: &types.GetModelsRequest{
				ProjectID:          "project-123",
				Page:               intPtr(1),
				Limit:              intPtr(10),
				ModelName:          stringPtr("gpt-4"),
				Provider:           stringPtr("openai"),
				ModelFamily:        stringPtr("gpt"),
				Unit:               &types.ModelUnitTokens,
				FromDate:           timePtr(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
				ToDate:             timePtr(time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)),
				IncludeDeprecated:  boolPtr(false),
			},
			serverResponse: `{
				"data": [
					{
						"id": "model-1",
						"modelName": "gpt-4",
						"provider": "openai",
						"unit": "TOKENS",
						"inputPrice": 0.03,
						"outputPrice": 0.06
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
				assert.Equal(t, "/api/public/models", r.URL.Path)
				
				query := r.URL.Query()
				assert.Equal(t, "project-123", query.Get("projectId"))
				assert.Equal(t, "1", query.Get("page"))
				assert.Equal(t, "10", query.Get("limit"))
				assert.Equal(t, "gpt-4", query.Get("modelName"))
				assert.Equal(t, "openai", query.Get("provider"))
				assert.Equal(t, "gpt", query.Get("modelFamily"))
				assert.Equal(t, "TOKENS", query.Get("unit"))
				assert.Equal(t, "false", query.Get("includeDeprecated"))
				assert.Contains(t, query.Get("fromDate"), "2024-01-01T00:00:00")
				assert.Contains(t, query.Get("toDate"), "2024-01-31T23:59:59")
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
			request:       &types.GetModelsRequest{},
			serverStatus:  http.StatusInternalServerError,
			expectError:   true,
			errorContains: "failed to list models",
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
		modelID        string
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
	}{
		{
			name:    "successful get",
			modelID: "model-123",
			serverResponse: `{
				"id": "model-123",
				"modelName": "gpt-4",
				"provider": "openai",
				"unit": "TOKENS",
				"inputPrice": 0.03,
				"outputPrice": 0.06
			}`,
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:          "empty model ID",
			modelID:       "",
			expectError:   true,
			errorContains: "model ID cannot be empty",
		},
		{
			name:          "model not found",
			modelID:       "nonexistent",
			serverStatus:  http.StatusNotFound,
			expectError:   true,
			errorContains: "failed to get model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.modelID != "" {
					assert.Equal(t, "GET", r.Method)
					assert.Equal(t, fmt.Sprintf("/api/public/models/%s", tt.modelID), r.URL.Path)
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
			response, err := client.Get(ctx, tt.modelID)

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
				assert.Equal(t, tt.modelID, response.ID)
			}
		})
	}
}

func TestClient_Create(t *testing.T) {
	tests := []struct {
		name           string
		request        *types.CreateModelRequest
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
	}{
		{
			name: "successful create",
			request: &types.CreateModelRequest{
				ModelName:   "custom-model",
				Provider:    "custom",
				Unit:        types.ModelUnitTokens,
				InputPrice:  floatPtr(0.01),
				OutputPrice: floatPtr(0.02),
			},
			serverResponse: `{
				"id": "model-123",
				"modelName": "custom-model",
				"provider": "custom",
				"unit": "TOKENS",
				"inputPrice": 0.01,
				"outputPrice": 0.02
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
			request:       &types.CreateModelRequest{ModelName: "test", Provider: "test"},
			serverStatus:  http.StatusBadRequest,
			expectError:   true,
			errorContains: "failed to create model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.request != nil {
					assert.Equal(t, "POST", r.Method)
					assert.Equal(t, "/api/public/models", r.URL.Path)
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
				assert.Equal(t, "custom-model", response.ModelName)
			}
		})
	}
}

func TestClient_Update(t *testing.T) {
	tests := []struct {
		name           string
		request        *types.UpdateModelRequest
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
	}{
		{
			name: "successful update",
			request: &types.UpdateModelRequest{
				ModelID:     "model-123",
				InputPrice:  floatPtr(0.02),
				OutputPrice: floatPtr(0.04),
			},
			serverResponse: `{
				"id": "model-123",
				"modelName": "gpt-4",
				"provider": "openai",
				"inputPrice": 0.02,
				"outputPrice": 0.04
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
			request:       &types.UpdateModelRequest{ModelID: "model-123"},
			serverStatus:  http.StatusNotFound,
			expectError:   true,
			errorContains: "failed to update model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.request != nil {
					assert.Equal(t, "PATCH", r.Method)
					assert.Equal(t, fmt.Sprintf("/api/public/models/%s", tt.request.ModelID), r.URL.Path)
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
				assert.Equal(t, "model-123", response.ID)
			}
		})
	}
}

func TestClient_Delete(t *testing.T) {
	tests := []struct {
		name          string
		modelID       string
		serverStatus  int
		expectError   bool
		errorContains string
	}{
		{
			name:         "successful delete",
			modelID:      "model-123",
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:          "empty model ID",
			modelID:       "",
			expectError:   true,
			errorContains: "model ID cannot be empty",
		},
		{
			name:          "model not found",
			modelID:       "nonexistent",
			serverStatus:  http.StatusNotFound,
			expectError:   true,
			errorContains: "failed to delete model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.modelID != "" {
					assert.Equal(t, "DELETE", r.Method)
					assert.Equal(t, fmt.Sprintf("/api/public/models/%s", tt.modelID), r.URL.Path)
				}

				w.WriteHeader(tt.serverStatus)
			}))
			defer server.Close()

			// Setup client
			restyClient := resty.New().SetBaseURL(server.URL)
			client := NewClient(restyClient)

			// Execute test
			ctx := context.Background()
			err := client.Delete(ctx, tt.modelID)

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

func TestClient_GetUsage(t *testing.T) {
	tests := []struct {
		name           string
		request        *types.GetModelUsageRequest
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
	}{
		{
			name: "successful usage retrieval",
			request: &types.GetModelUsageRequest{
				ProjectID: "project-123",
				ModelID:   stringPtr("model-456"),
			},
			serverResponse: `{
				"totalUsage": 1000,
				"totalCost": 5.00,
				"usage": [
					{
						"modelId": "model-456",
						"usage": 500,
						"cost": 2.50
					}
				]
			}`,
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:           "successful usage with nil request",
			request:        nil,
			serverResponse: `{"totalUsage": 0, "totalCost": 0, "usage": []}`,
			serverStatus:   http.StatusOK,
			expectError:    false,
		},
		{
			name:          "server error",
			request:       &types.GetModelUsageRequest{},
			serverStatus:  http.StatusInternalServerError,
			expectError:   true,
			errorContains: "failed to get model usage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/api/public/models/usage", r.URL.Path)

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
			response, err := client.GetUsage(ctx, tt.request)

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
		modelID       string
		serverStatus  int
		expectedExist bool
		expectError   bool
		errorContains string
	}{
		{
			name:          "model exists",
			modelID:       "existing-model",
			serverStatus:  http.StatusOK,
			expectedExist: true,
			expectError:   false,
		},
		{
			name:          "model does not exist",
			modelID:       "nonexistent-model",
			serverStatus:  http.StatusNotFound,
			expectedExist: false,
			expectError:   false,
		},
		{
			name:          "empty model ID",
			modelID:       "",
			expectError:   true,
			errorContains: "model ID cannot be empty",
		},
		{
			name:          "server error",
			modelID:       "model-123",
			serverStatus:  http.StatusInternalServerError,
			expectError:   true,
			errorContains: "failed to get model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
				if tt.serverStatus == http.StatusOK {
					w.Write([]byte(`{"id": "` + tt.modelID + `", "modelName": "test"}`))
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
			exists, err := client.Exists(ctx, tt.modelID)

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
		w.Write([]byte(`{"id": "model-123", "modelName": "test"}`))
	}))
	defer server.Close()

	// Setup client
	restyClient := resty.New().SetBaseURL(server.URL)
	client := NewClient(restyClient)

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.Get(ctx, "model-123")
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

func floatPtr(f float64) *float64 {
	return &f
}

func boolPtr(b bool) *bool {
	return &b
}

func timePtr(t time.Time) *time.Time {
	return &t
}