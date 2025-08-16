package types

import (
	"encoding/json"
	"time"
)

// Dataset represents a dataset for storing test cases and evaluation data
type Dataset struct {
	// Unique identifier for the dataset
	ID string `json:"id"`

	// Name of the dataset
	Name string `json:"name"`

	// Description of the dataset
	Description *string `json:"description,omitempty"`

	// Project ID this dataset belongs to
	ProjectID string `json:"projectId"`

	// Timestamp when the dataset was created
	CreatedAt time.Time `json:"createdAt"`

	// Timestamp when the dataset was last updated
	UpdatedAt time.Time `json:"updatedAt"`

	// Metadata associated with the dataset
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Number of items in the dataset
	ItemCount *int `json:"itemCount,omitempty"`

	// Number of runs associated with this dataset
	RunCount *int `json:"runCount,omitempty"`
}

// DatasetItem represents an individual item within a dataset
type DatasetItem struct {
	// Unique identifier for the dataset item
	ID string `json:"id"`

	// Dataset ID this item belongs to
	DatasetID string `json:"datasetId"`

	// Input data for the dataset item
	Input json.RawMessage `json:"input,omitempty"`

	// Expected output data for the dataset item
	ExpectedOutput json.RawMessage `json:"expectedOutput,omitempty"`

	// Metadata associated with the dataset item
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Source trace ID if this item was created from a trace
	SourceTraceID *string `json:"sourceTraceId,omitempty"`

	// Source observation ID if this item was created from an observation
	SourceObservationID *string `json:"sourceObservationId,omitempty"`

	// Timestamp when the item was created
	CreatedAt time.Time `json:"createdAt"`

	// Timestamp when the item was last updated
	UpdatedAt time.Time `json:"updatedAt"`

	// Status of the dataset item
	Status *string `json:"status,omitempty"`
}

// DatasetRun represents a run/evaluation session against a dataset
type DatasetRun struct {
	// Unique identifier for the dataset run
	ID string `json:"id"`

	// Name of the dataset run
	Name string `json:"name"`

	// Description of the dataset run
	Description *string `json:"description,omitempty"`

	// Dataset ID this run is associated with
	DatasetID string `json:"datasetId"`

	// Metadata associated with the dataset run
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Timestamp when the run was created
	CreatedAt time.Time `json:"createdAt"`

	// Timestamp when the run was last updated
	UpdatedAt time.Time `json:"updatedAt"`

	// Number of items processed in this run
	ItemCount *int `json:"itemCount,omitempty"`

	// Status of the dataset run
	Status *string `json:"status,omitempty"`
}

// DatasetRunItem represents an individual run item within a dataset run
type DatasetRunItem struct {
	// Unique identifier for the dataset run item
	ID string `json:"id"`

	// Dataset run ID this item belongs to
	DatasetRunID string `json:"datasetRunId"`

	// Dataset item ID this run item corresponds to
	DatasetItemID string `json:"datasetItemId"`

	// Trace ID generated during this run
	TraceID *string `json:"traceId,omitempty"`

	// Observation ID if specific to an observation
	ObservationID *string `json:"observationId,omitempty"`

	// Input data used for this run item
	Input json.RawMessage `json:"input,omitempty"`

	// Expected output for comparison
	ExpectedOutput json.RawMessage `json:"expectedOutput,omitempty"`

	// Actual output generated during the run
	Output json.RawMessage `json:"output,omitempty"`

	// Metadata associated with the run item
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Timestamp when the run item was created
	CreatedAt time.Time `json:"createdAt"`

	// Timestamp when the run item was completed
	CompletedAt *time.Time `json:"completedAt,omitempty"`

	// Status of the run item
	Status *string `json:"status,omitempty"`
}

// DatasetCreateRequest represents a request to create a new dataset
type DatasetCreateRequest struct {
	// Name of the dataset
	Name string `json:"name"`

	// Description of the dataset
	Description *string `json:"description,omitempty"`

	// Metadata associated with the dataset
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// DatasetItemCreateRequest represents a request to create a new dataset item
type DatasetItemCreateRequest struct {
	// Unique identifier for the dataset item
	ID *string `json:"id,omitempty"`

	// Dataset ID this item belongs to
	DatasetID string `json:"datasetId"`

	// Input data for the dataset item
	Input interface{} `json:"input,omitempty"`

	// Expected output data for the dataset item
	ExpectedOutput interface{} `json:"expectedOutput,omitempty"`

	// Metadata associated with the dataset item
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Source trace ID if this item was created from a trace
	SourceTraceID *string `json:"sourceTraceId,omitempty"`

	// Source observation ID if this item was created from an observation
	SourceObservationID *string `json:"sourceObservationId,omitempty"`
}

// DatasetRunCreateRequest represents a request to create a new dataset run
type DatasetRunCreateRequest struct {
	// Name of the dataset run
	Name string `json:"name"`

	// Description of the dataset run
	Description *string `json:"description,omitempty"`

	// Dataset ID this run is associated with
	DatasetID string `json:"datasetId"`

	// Metadata associated with the dataset run
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// DatasetListResponse represents the response from listing datasets
type DatasetListResponse struct {
	Data []Dataset `json:"data"`
	Meta struct {
		Page       int `json:"page"`
		Limit      int `json:"limit"`
		TotalItems int `json:"totalItems"`
		TotalPages int `json:"totalPages"`
	} `json:"meta"`
}
