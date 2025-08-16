package types

import (
	"encoding/json"
	"time"
)

// ScoreDataType represents the data type of a score value
type ScoreDataType string

const (
	// ScoreDataTypeNumeric for numeric scores (float64)
	ScoreDataTypeNumeric ScoreDataType = "NUMERIC"

	// ScoreDataTypeCategorical for categorical scores (string)
	ScoreDataTypeCategorical ScoreDataType = "CATEGORICAL"

	// ScoreDataTypeBoolean for boolean scores (bool)
	ScoreDataTypeBoolean ScoreDataType = "BOOLEAN"
)

// ScoreSource represents the source of a score
type ScoreSource string

const (
	// ScoreSourceAPI indicates the score was created via API
	ScoreSourceAPI ScoreSource = "API"

	// ScoreSourceAnnotation indicates the score was created via manual annotation
	ScoreSourceAnnotation ScoreSource = "ANNOTATION"

	// ScoreSourceReview indicates the score was created via review process
	ScoreSourceReview ScoreSource = "REVIEW"
)

// Score represents a score/evaluation for traces or observations
type Score struct {
	// Unique identifier for the score
	ID string `json:"id"`

	// Timestamp when the score was created
	Timestamp time.Time `json:"timestamp"`

	// Name/identifier of the score metric
	Name string `json:"name"`

	// The actual score value
	Value json.RawMessage `json:"value"`

	// Data type of the score value
	DataType ScoreDataType `json:"dataType"`

	// Source of the score
	Source *ScoreSource `json:"source,omitempty"`

	// ID of the trace this score belongs to
	TraceID string `json:"traceId"`

	// ID of the observation this score belongs to (optional)
	ObservationID *string `json:"observationId,omitempty"`

	// Comment or explanation for the score
	Comment *string `json:"comment,omitempty"`

	// Author/creator of the score
	AuthorUserID *string `json:"authorUserId,omitempty"`

	// Configuration ID used for this score
	ConfigID *string `json:"configId,omitempty"`

	// Queue ID if this score was generated from a queue
	QueueID *string `json:"queueId,omitempty"`
}

// ScoreCreateRequest represents a request to create a new score
type ScoreCreateRequest struct {
	// Unique identifier for the score
	ID *string `json:"id,omitempty"`

	// Timestamp when the score was created
	Timestamp *time.Time `json:"timestamp,omitempty"`

	// Name/identifier of the score metric
	Name string `json:"name"`

	// The actual score value
	Value interface{} `json:"value"`

	// Data type of the score value
	DataType *ScoreDataType `json:"dataType,omitempty"`

	// Source of the score
	Source *ScoreSource `json:"source,omitempty"`

	// ID of the trace this score belongs to
	TraceID string `json:"traceId"`

	// ID of the observation this score belongs to (optional)
	ObservationID *string `json:"observationId,omitempty"`

	// Comment or explanation for the score
	Comment *string `json:"comment,omitempty"`

	// Author/creator of the score
	AuthorUserID *string `json:"authorUserId,omitempty"`

	// Configuration ID used for this score
	ConfigID *string `json:"configId,omitempty"`

	// Queue ID if this score was generated from a queue
	QueueID *string `json:"queueId,omitempty"`
}

// ScoreUpdateRequest represents a request to update an existing score
type ScoreUpdateRequest struct {
	// Name/identifier of the score metric
	Name *string `json:"name,omitempty"`

	// The actual score value
	Value interface{} `json:"value,omitempty"`

	// Comment or explanation for the score
	Comment *string `json:"comment,omitempty"`
}

// ScoreConfig represents configuration for automated scoring
type ScoreConfig struct {
	// Unique identifier for the score configuration
	ID string `json:"id"`

	// Name of the score configuration
	Name string `json:"name"`

	// Data type expected for this score
	DataType ScoreDataType `json:"dataType"`

	// Whether this score configuration is active
	IsArchived *bool `json:"isArchived,omitempty"`

	// Minimum value for numeric scores
	MinValue *float64 `json:"minValue,omitempty"`

	// Maximum value for numeric scores
	MaxValue *float64 `json:"maxValue,omitempty"`

	// Valid categories for categorical scores
	Categories []string `json:"categories,omitempty"`

	// Description of what this score measures
	Description *string `json:"description,omitempty"`
}

// NumericScore creates a ScoreCreateRequest for a numeric score
func NumericScore(name string, value float64, traceID string) *ScoreCreateRequest {
	dataType := ScoreDataTypeNumeric
	source := ScoreSourceAPI

	return &ScoreCreateRequest{
		Name:     name,
		Value:    value,
		DataType: &dataType,
		Source:   &source,
		TraceID:  traceID,
	}
}

// CategoricalScore creates a ScoreCreateRequest for a categorical score
func CategoricalScore(name string, value string, traceID string) *ScoreCreateRequest {
	dataType := ScoreDataTypeCategorical
	source := ScoreSourceAPI

	return &ScoreCreateRequest{
		Name:     name,
		Value:    value,
		DataType: &dataType,
		Source:   &source,
		TraceID:  traceID,
	}
}

// BooleanScore creates a ScoreCreateRequest for a boolean score
func BooleanScore(name string, value bool, traceID string) *ScoreCreateRequest {
	dataType := ScoreDataTypeBoolean
	source := ScoreSourceAPI

	return &ScoreCreateRequest{
		Name:     name,
		Value:    value,
		DataType: &dataType,
		Source:   &source,
		TraceID:  traceID,
	}
}
