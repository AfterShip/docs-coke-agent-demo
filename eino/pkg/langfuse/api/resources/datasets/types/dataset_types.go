package types

import (
	"time"

	commonTypes "eino/pkg/langfuse/api/resources/commons/types"
	"eino/pkg/langfuse/api/resources/utils/pagination/types"
)

// GetDatasetsRequest represents a request to list datasets
type GetDatasetsRequest struct {
	ProjectID string `json:"projectId,omitempty"`
	Page      *int   `json:"page,omitempty"`
	Limit     *int   `json:"limit,omitempty"`
	Name      *string `json:"name,omitempty"`
	FromTimestamp *time.Time `json:"fromTimestamp,omitempty"`
	ToTimestamp   *time.Time `json:"toTimestamp,omitempty"`
}

// GetDatasetsResponse represents the response from listing datasets
type GetDatasetsResponse struct {
	Data []commonTypes.Dataset `json:"data"`
	Meta types.MetaResponse    `json:"meta"`
}

// CreateDatasetRequest represents a request to create a dataset
type CreateDatasetRequest struct {
	Name        string                 `json:"name"`
	Description *string                `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CreateDatasetResponse represents the response from creating a dataset
type CreateDatasetResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description *string                `json:"description,omitempty"`
	ProjectID   string                 `json:"projectId"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
}

// UpdateDatasetRequest represents a request to update a dataset
type UpdateDatasetRequest struct {
	Name        *string                `json:"name,omitempty"`
	Description *string                `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// GetDatasetItemsRequest represents a request to list dataset items
type GetDatasetItemsRequest struct {
	DatasetID string `json:"datasetId"`
	Page      *int   `json:"page,omitempty"`
	Limit     *int   `json:"limit,omitempty"`
	FromTimestamp *time.Time `json:"fromTimestamp,omitempty"`
	ToTimestamp   *time.Time `json:"toTimestamp,omitempty"`
}

// GetDatasetItemsResponse represents the response from listing dataset items
type GetDatasetItemsResponse struct {
	Data []commonTypes.DatasetItem `json:"data"`
	Meta types.MetaResponse        `json:"meta"`
}

// CreateDatasetItemRequest represents a request to create a dataset item
type CreateDatasetItemRequest struct {
	ID                  *string                `json:"id,omitempty"`
	Input               interface{}            `json:"input,omitempty"`
	ExpectedOutput      interface{}            `json:"expectedOutput,omitempty"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
	SourceTraceID       *string                `json:"sourceTraceId,omitempty"`
	SourceObservationID *string                `json:"sourceObservationId,omitempty"`
}

// CreateDatasetItemResponse represents the response from creating a dataset item
type CreateDatasetItemResponse struct {
	ID                  string                 `json:"id"`
	DatasetID           string                 `json:"datasetId"`
	Input               interface{}            `json:"input,omitempty"`
	ExpectedOutput      interface{}            `json:"expectedOutput,omitempty"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
	SourceTraceID       *string                `json:"sourceTraceId,omitempty"`
	SourceObservationID *string                `json:"sourceObservationId,omitempty"`
	CreatedAt           time.Time              `json:"createdAt"`
	UpdatedAt           time.Time              `json:"updatedAt"`
}

// UpdateDatasetItemRequest represents a request to update a dataset item
type UpdateDatasetItemRequest struct {
	Input          interface{}            `json:"input,omitempty"`
	ExpectedOutput interface{}            `json:"expectedOutput,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// GetDatasetRunsRequest represents a request to list dataset runs
type GetDatasetRunsRequest struct {
	DatasetID string     `json:"datasetId"`
	Page      *int       `json:"page,omitempty"`
	Limit     *int       `json:"limit,omitempty"`
	Name      *string    `json:"name,omitempty"`
	FromTimestamp *time.Time `json:"fromTimestamp,omitempty"`
	ToTimestamp   *time.Time `json:"toTimestamp,omitempty"`
}

// GetDatasetRunsResponse represents the response from listing dataset runs
type GetDatasetRunsResponse struct {
	Data []commonTypes.DatasetRun `json:"data"`
	Meta types.MetaResponse       `json:"meta"`
}

// CreateDatasetRunRequest represents a request to create a dataset run
type CreateDatasetRunRequest struct {
	Name        string                 `json:"name"`
	Description *string                `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CreateDatasetRunResponse represents the response from creating a dataset run
type CreateDatasetRunResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description *string                `json:"description,omitempty"`
	DatasetID   string                 `json:"datasetId"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
}

// GetDatasetRunItemsRequest represents a request to list dataset run items
type GetDatasetRunItemsRequest struct {
	DatasetRunID string `json:"datasetRunId"`
	Page         *int   `json:"page,omitempty"`
	Limit        *int   `json:"limit,omitempty"`
	FromTimestamp *time.Time `json:"fromTimestamp,omitempty"`
	ToTimestamp   *time.Time `json:"toTimestamp,omitempty"`
}

// GetDatasetRunItemsResponse represents the response from listing dataset run items
type GetDatasetRunItemsResponse struct {
	Data []commonTypes.DatasetRunItem `json:"data"`
	Meta types.MetaResponse           `json:"meta"`
}

// CreateDatasetRunItemRequest represents a request to create a dataset run item
type CreateDatasetRunItemRequest struct {
	DatasetItemID  string      `json:"datasetItemId"`
	TraceID        *string     `json:"traceId,omitempty"`
	ObservationID  *string     `json:"observationId,omitempty"`
	Input          interface{} `json:"input,omitempty"`
	ExpectedOutput interface{} `json:"expectedOutput,omitempty"`
	Output         interface{} `json:"output,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// CreateDatasetRunItemResponse represents the response from creating a dataset run item
type CreateDatasetRunItemResponse struct {
	ID             string      `json:"id"`
	DatasetRunID   string      `json:"datasetRunId"`
	DatasetItemID  string      `json:"datasetItemId"`
	TraceID        *string     `json:"traceId,omitempty"`
	ObservationID  *string     `json:"observationId,omitempty"`
	Input          interface{} `json:"input,omitempty"`
	ExpectedOutput interface{} `json:"expectedOutput,omitempty"`
	Output         interface{} `json:"output,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time   `json:"createdAt"`
	CompletedAt    *time.Time  `json:"completedAt,omitempty"`
}

// DatasetStats represents statistics about datasets
type DatasetStats struct {
	TotalDatasets     int            `json:"totalDatasets"`
	TotalItems        int            `json:"totalItems"`
	TotalRuns         int            `json:"totalRuns"`
	DatasetsByProject map[string]int `json:"datasetsByProject"`
	RecentActivity    []RecentDatasetActivity `json:"recentActivity"`
	DateRange         *DateRange     `json:"dateRange,omitempty"`
}

// RecentDatasetActivity represents recent activity on datasets
type RecentDatasetActivity struct {
	DatasetID   string    `json:"datasetId"`
	DatasetName string    `json:"datasetName"`
	Activity    string    `json:"activity"` // "created", "updated", "run_created"
	Timestamp   time.Time `json:"timestamp"`
	UserID      *string   `json:"userId,omitempty"`
}

// DateRange represents a date range
type DateRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// GetDatasetStatsRequest represents a request to get dataset statistics
type GetDatasetStatsRequest struct {
	ProjectID     string     `json:"projectId,omitempty"`
	FromTimestamp *time.Time `json:"fromTimestamp,omitempty"`
	ToTimestamp   *time.Time `json:"toTimestamp,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// Validate validates the GetDatasetsRequest
func (req *GetDatasetsRequest) Validate() error {
	if req.Limit != nil && (*req.Limit < 1 || *req.Limit > 1000) {
		return &ValidationError{Field: "limit", Message: "limit must be between 1 and 1000"}
	}
	
	if req.Page != nil && *req.Page < 1 {
		return &ValidationError{Field: "page", Message: "page must be greater than 0"}
	}
	
	if req.FromTimestamp != nil && req.ToTimestamp != nil && req.FromTimestamp.After(*req.ToTimestamp) {
		return &ValidationError{Field: "timestamps", Message: "fromTimestamp cannot be after toTimestamp"}
	}
	
	return nil
}

// Validate validates the CreateDatasetRequest
func (req *CreateDatasetRequest) Validate() error {
	if req.Name == "" {
		return &ValidationError{Field: "name", Message: "name is required"}
	}
	
	if len(req.Name) > 255 {
		return &ValidationError{Field: "name", Message: "name must be 255 characters or less"}
	}
	
	if req.Description != nil && len(*req.Description) > 2000 {
		return &ValidationError{Field: "description", Message: "description must be 2000 characters or less"}
	}
	
	return nil
}

// Validate validates the CreateDatasetItemRequest
func (req *CreateDatasetItemRequest) Validate() error {
	if req.Input == nil && req.ExpectedOutput == nil {
		return &ValidationError{Field: "input_or_output", Message: "at least one of input or expectedOutput is required"}
	}
	
	return nil
}

// Validate validates the CreateDatasetRunRequest
func (req *CreateDatasetRunRequest) Validate() error {
	if req.Name == "" {
		return &ValidationError{Field: "name", Message: "name is required"}
	}
	
	if len(req.Name) > 255 {
		return &ValidationError{Field: "name", Message: "name must be 255 characters or less"}
	}
	
	if req.Description != nil && len(*req.Description) > 2000 {
		return &ValidationError{Field: "description", Message: "description must be 2000 characters or less"}
	}
	
	return nil
}

// Validate validates the CreateDatasetRunItemRequest
func (req *CreateDatasetRunItemRequest) Validate() error {
	if req.DatasetItemID == "" {
		return &ValidationError{Field: "datasetItemId", Message: "datasetItemId is required"}
	}
	
	return nil
}

// Validate validates the GetDatasetItemsRequest
func (req *GetDatasetItemsRequest) Validate() error {
	if req.DatasetID == "" {
		return &ValidationError{Field: "datasetId", Message: "datasetId is required"}
	}
	
	if req.Limit != nil && (*req.Limit < 1 || *req.Limit > 1000) {
		return &ValidationError{Field: "limit", Message: "limit must be between 1 and 1000"}
	}
	
	if req.Page != nil && *req.Page < 1 {
		return &ValidationError{Field: "page", Message: "page must be greater than 0"}
	}
	
	if req.FromTimestamp != nil && req.ToTimestamp != nil && req.FromTimestamp.After(*req.ToTimestamp) {
		return &ValidationError{Field: "timestamps", Message: "fromTimestamp cannot be after toTimestamp"}
	}
	
	return nil
}

// Validate validates the GetDatasetRunsRequest
func (req *GetDatasetRunsRequest) Validate() error {
	if req.DatasetID == "" {
		return &ValidationError{Field: "datasetId", Message: "datasetId is required"}
	}
	
	if req.Limit != nil && (*req.Limit < 1 || *req.Limit > 1000) {
		return &ValidationError{Field: "limit", Message: "limit must be between 1 and 1000"}
	}
	
	if req.Page != nil && *req.Page < 1 {
		return &ValidationError{Field: "page", Message: "page must be greater than 0"}
	}
	
	if req.FromTimestamp != nil && req.ToTimestamp != nil && req.FromTimestamp.After(*req.ToTimestamp) {
		return &ValidationError{Field: "timestamps", Message: "fromTimestamp cannot be after toTimestamp"}
	}
	
	return nil
}

// Validate validates the GetDatasetRunItemsRequest
func (req *GetDatasetRunItemsRequest) Validate() error {
	if req.DatasetRunID == "" {
		return &ValidationError{Field: "datasetRunId", Message: "datasetRunId is required"}
	}
	
	if req.Limit != nil && (*req.Limit < 1 || *req.Limit > 1000) {
		return &ValidationError{Field: "limit", Message: "limit must be between 1 and 1000"}
	}
	
	if req.Page != nil && *req.Page < 1 {
		return &ValidationError{Field: "page", Message: "page must be greater than 0"}
	}
	
	if req.FromTimestamp != nil && req.ToTimestamp != nil && req.FromTimestamp.After(*req.ToTimestamp) {
		return &ValidationError{Field: "timestamps", Message: "fromTimestamp cannot be after toTimestamp"}
	}
	
	return nil
}