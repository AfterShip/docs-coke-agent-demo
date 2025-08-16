package types

import (
	"encoding/json"
	"fmt"
	"time"
)

// IngestionRequest represents a batch request to the Langfuse ingestion API
type IngestionRequest struct {
	Batch    []IngestionEvent       `json:"batch"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// IngestionBatchMetadata contains metadata about the batch
type IngestionBatchMetadata struct {
	SDKVersion     string                 `json:"sdk_version,omitempty"`
	SDKIntegration string                 `json:"sdk_integration,omitempty"`
	Timestamp      int64                  `json:"timestamp,omitempty"`
	BatchSize      int                    `json:"batch_size,omitempty"`
	ClientID       string                 `json:"client_id,omitempty"`
	Additional     map[string]interface{} `json:"additional,omitempty"`
}

// NewIngestionRequest creates a new ingestion request with the given events
func NewIngestionRequest(events []IngestionEvent) *IngestionRequest {
	return &IngestionRequest{
		Batch: events,
		Metadata: map[string]interface{}{
			"sdk_version":     "langfuse-go/1.0.0",
			"sdk_integration": "go",
			"timestamp":       time.Now().Unix(),
			"batch_size":      len(events),
		},
	}
}

// NewIngestionRequestWithMetadata creates a new ingestion request with events and custom metadata
func NewIngestionRequestWithMetadata(events []IngestionEvent, metadata *IngestionBatchMetadata) *IngestionRequest {
	req := &IngestionRequest{
		Batch:    events,
		Metadata: make(map[string]interface{}),
	}

	// Set default metadata
	req.Metadata["sdk_version"] = "langfuse-go/1.0.0"
	req.Metadata["sdk_integration"] = "go"
	req.Metadata["timestamp"] = time.Now().Unix()
	req.Metadata["batch_size"] = len(events)

	// Override with custom metadata
	if metadata != nil {
		if metadata.SDKVersion != "" {
			req.Metadata["sdk_version"] = metadata.SDKVersion
		}
		if metadata.SDKIntegration != "" {
			req.Metadata["sdk_integration"] = metadata.SDKIntegration
		}
		if metadata.Timestamp != 0 {
			req.Metadata["timestamp"] = metadata.Timestamp
		}
		if metadata.BatchSize != 0 {
			req.Metadata["batch_size"] = metadata.BatchSize
		}
		if metadata.ClientID != "" {
			req.Metadata["client_id"] = metadata.ClientID
		}
		
		// Merge additional metadata
		if metadata.Additional != nil {
			for k, v := range metadata.Additional {
				req.Metadata[k] = v
			}
		}
	}

	return req
}

// AddEvent adds an event to the batch
func (r *IngestionRequest) AddEvent(event IngestionEvent) {
	r.Batch = append(r.Batch, event)
	
	// Update batch size in metadata
	if r.Metadata != nil {
		r.Metadata["batch_size"] = len(r.Batch)
	}
}

// Size returns the number of events in the batch
func (r *IngestionRequest) Size() int {
	return len(r.Batch)
}

// IsEmpty returns true if the batch is empty
func (r *IngestionRequest) IsEmpty() bool {
	return len(r.Batch) == 0
}

// Validate performs validation on the ingestion request
func (r *IngestionRequest) Validate() error {
	if r.Batch == nil {
		return &RequestValidationError{Field: "batch", Message: "batch is required"}
	}

	if len(r.Batch) == 0 {
		return &RequestValidationError{Field: "batch", Message: "batch cannot be empty"}
	}

	if len(r.Batch) > MaxBatchSize {
		return &RequestValidationError{
			Field:   "batch", 
			Message: fmt.Sprintf("batch size %d exceeds maximum allowed size %d", len(r.Batch), MaxBatchSize),
		}
	}

	// Validate each event in the batch
	for i, event := range r.Batch {
		if err := event.Validate(); err != nil {
			return &RequestValidationError{
				Field:   fmt.Sprintf("batch[%d]", i),
				Message: fmt.Sprintf("event validation failed: %v", err),
			}
		}
	}

	return nil
}

// MarshalJSON implements json.Marshaler for IngestionRequest
func (r *IngestionRequest) MarshalJSON() ([]byte, error) {
	type Alias IngestionRequest
	return json.Marshal((*Alias)(r))
}

// UnmarshalJSON implements json.Unmarshaler for IngestionRequest
func (r *IngestionRequest) UnmarshalJSON(data []byte) error {
	type Alias IngestionRequest
	aux := (*Alias)(r)
	return json.Unmarshal(data, aux)
}

// RequestValidationError represents a validation error for ingestion requests
type RequestValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *RequestValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// Constants for batch processing
const (
	MaxBatchSize     = 100  // Maximum number of events per batch
	DefaultBatchSize = 15   // Default batch size
	MinBatchSize     = 1    // Minimum batch size
)