package datasets

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"eino/pkg/langfuse/api/resources/datasets/types"
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
		request        *types.GetDatasetsRequest
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
		verifyRequest  func(t *testing.T, r *http.Request)
	}{
		{
			name: "successful list with filters",
			request: &types.GetDatasetsRequest{
				ProjectID: "project-123",
				Page:      intPtr(1),
				Limit:     intPtr(10),
				Name:      stringPtr("test-dataset"),
			},
			serverResponse: `{
				"data": [
					{
						"id": "dataset-1",
						"name": "test-dataset",
						"description": "A test dataset",
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
				assert.Equal(t, "/api/public/datasets", r.URL.Path)
				
				query := r.URL.Query()
				assert.Equal(t, "project-123", query.Get("projectId"))
				assert.Equal(t, "1", query.Get("page"))
				assert.Equal(t, "10", query.Get("limit"))
				assert.Equal(t, "test-dataset", query.Get("name"))
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
			request:       &types.GetDatasetsRequest{},
			serverStatus:  http.StatusInternalServerError,
			expectError:   true,
			errorContains: "failed to list datasets",
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
		datasetID      string
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
	}{
		{
			name:      "successful get",
			datasetID: "dataset-123",
			serverResponse: `{
				"id": "dataset-123",
				"name": "test-dataset",
				"description": "A test dataset",
				"createdAt": "2024-01-15T12:00:00Z"
			}`,
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:          "empty dataset ID",
			datasetID:     "",
			expectError:   true,
			errorContains: "dataset ID cannot be empty",
		},
		{
			name:          "dataset not found",
			datasetID:     "nonexistent",
			serverStatus:  http.StatusNotFound,
			expectError:   true,
			errorContains: "failed to get dataset",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.datasetID != "" {
					assert.Equal(t, "GET", r.Method)
					assert.Equal(t, fmt.Sprintf("/api/public/datasets/%s", tt.datasetID), r.URL.Path)
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
			response, err := client.Get(ctx, tt.datasetID)

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
				assert.Equal(t, tt.datasetID, response.ID)
			}
		})
	}
}

func TestClient_Create(t *testing.T) {
	tests := []struct {
		name           string
		request        *types.CreateDatasetRequest
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
	}{
		{
			name: "successful create",
			request: &types.CreateDatasetRequest{
				Name:        "new-dataset",
				Description: stringPtr("A new test dataset"),
				Metadata: map[string]interface{}{
					"version": "1.0",
					"author":  "test",
				},
			},
			serverResponse: `{
				"id": "dataset-123",
				"name": "new-dataset",
				"description": "A new test dataset",
				"createdAt": "2024-01-15T12:00:00Z"
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
			request:       &types.CreateDatasetRequest{Name: "test"},
			serverStatus:  http.StatusBadRequest,
			expectError:   true,
			errorContains: "failed to create dataset",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.request != nil {
					assert.Equal(t, "POST", r.Method)
					assert.Equal(t, "/api/public/datasets", r.URL.Path)
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
				assert.Equal(t, "new-dataset", response.Name)
			}
		})
	}
}

func TestClient_Update(t *testing.T) {
	tests := []struct {
		name           string
		request        *types.UpdateDatasetRequest
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
	}{
		{
			name: "successful update",
			request: &types.UpdateDatasetRequest{
				DatasetID:   "dataset-123",
				Name:        stringPtr("updated-dataset"),
				Description: stringPtr("Updated description"),
			},
			serverResponse: `{
				"id": "dataset-123",
				"name": "updated-dataset",
				"description": "Updated description",
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
			request:       &types.UpdateDatasetRequest{DatasetID: "dataset-123"},
			serverStatus:  http.StatusNotFound,
			expectError:   true,
			errorContains: "failed to update dataset",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.request != nil {
					assert.Equal(t, "PATCH", r.Method)
					assert.Equal(t, fmt.Sprintf("/api/public/datasets/%s", tt.request.DatasetID), r.URL.Path)
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
				assert.Equal(t, "dataset-123", response.ID)
			}
		})
	}
}

func TestClient_Delete(t *testing.T) {
	tests := []struct {
		name          string
		datasetID     string
		serverStatus  int
		expectError   bool
		errorContains string
	}{
		{
			name:         "successful delete",
			datasetID:    "dataset-123",
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:          "empty dataset ID",
			datasetID:     "",
			expectError:   true,
			errorContains: "dataset ID cannot be empty",
		},
		{
			name:          "dataset not found",
			datasetID:     "nonexistent",
			serverStatus:  http.StatusNotFound,
			expectError:   true,
			errorContains: "failed to delete dataset",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.datasetID != "" {
					assert.Equal(t, "DELETE", r.Method)
					assert.Equal(t, fmt.Sprintf("/api/public/datasets/%s", tt.datasetID), r.URL.Path)
				}

				w.WriteHeader(tt.serverStatus)
			}))
			defer server.Close()

			// Setup client
			restyClient := resty.New().SetBaseURL(server.URL)
			client := NewClient(restyClient)

			// Execute test
			ctx := context.Background()
			err := client.Delete(ctx, tt.datasetID)

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

func TestClient_GetItems(t *testing.T) {
	tests := []struct {
		name           string
		datasetID      string
		request        *types.GetDatasetItemsRequest
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
	}{
		{
			name:      "successful get items",
			datasetID: "dataset-123",
			request: &types.GetDatasetItemsRequest{
				Page:  intPtr(1),
				Limit: intPtr(10),
			},
			serverResponse: `{
				"data": [
					{
						"id": "item-1",
						"datasetId": "dataset-123",
						"input": {"query": "test"},
						"output": {"result": "success"}
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
		},
		{
			name:          "empty dataset ID",
			datasetID:     "",
			request:       &types.GetDatasetItemsRequest{},
			expectError:   true,
			errorContains: "dataset ID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.datasetID != "" {
					assert.Equal(t, "GET", r.Method)
					assert.Equal(t, fmt.Sprintf("/api/public/datasets/%s/items", tt.datasetID), r.URL.Path)
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
			response, err := client.GetItems(ctx, tt.datasetID, tt.request)

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
			}
		})
	}
}

func TestClient_CreateItem(t *testing.T) {
	tests := []struct {
		name           string
		datasetID      string
		request        *types.CreateDatasetItemRequest
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
	}{
		{
			name:      "successful create item",
			datasetID: "dataset-123",
			request: &types.CreateDatasetItemRequest{
				Input:  map[string]interface{}{"query": "test"},
				Output: map[string]interface{}{"result": "success"},
			},
			serverResponse: `{
				"id": "item-123",
				"datasetId": "dataset-123",
				"input": {"query": "test"},
				"output": {"result": "success"}
			}`,
			serverStatus: http.StatusCreated,
			expectError:  false,
		},
		{
			name:          "empty dataset ID",
			datasetID:     "",
			request:       &types.CreateDatasetItemRequest{},
			expectError:   true,
			errorContains: "dataset ID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.datasetID != "" {
					assert.Equal(t, "POST", r.Method)
					assert.Equal(t, fmt.Sprintf("/api/public/datasets/%s/items", tt.datasetID), r.URL.Path)
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
			response, err := client.CreateItem(ctx, tt.datasetID, tt.request)

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
				assert.Equal(t, "dataset-123", response.DatasetID)
			}
		})
	}
}

func TestClient_Exists(t *testing.T) {
	tests := []struct {
		name          string
		datasetID     string
		serverStatus  int
		expectedExist bool
		expectError   bool
		errorContains string
	}{
		{
			name:          "dataset exists",
			datasetID:     "existing-dataset",
			serverStatus:  http.StatusOK,
			expectedExist: true,
			expectError:   false,
		},
		{
			name:          "dataset does not exist",
			datasetID:     "nonexistent-dataset",
			serverStatus:  http.StatusNotFound,
			expectedExist: false,
			expectError:   false,
		},
		{
			name:          "empty dataset ID",
			datasetID:     "",
			expectError:   true,
			errorContains: "dataset ID cannot be empty",
		},
		{
			name:          "server error",
			datasetID:     "dataset-123",
			serverStatus:  http.StatusInternalServerError,
			expectError:   true,
			errorContains: "failed to get dataset",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
				if tt.serverStatus == http.StatusOK {
					w.Write([]byte(`{"id": "` + tt.datasetID + `", "name": "test"}`))
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
			exists, err := client.Exists(ctx, tt.datasetID)

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
		w.Write([]byte(`{"id": "dataset-123", "name": "test"}`))
	}))
	defer server.Close()

	// Setup client
	restyClient := resty.New().SetBaseURL(server.URL)
	client := NewClient(restyClient)

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.Get(ctx, "dataset-123")
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