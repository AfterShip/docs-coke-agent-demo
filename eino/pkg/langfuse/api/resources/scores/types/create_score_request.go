package types

import (
	"encoding/json"
	"time"

	commonTypes "eino/pkg/langfuse/api/resources/commons/types"
)

// CreateScoreRequest represents a request to create a score
type CreateScoreRequest struct {
	ID            *string                   `json:"id,omitempty"`
	TraceID       string                    `json:"traceId"`
	ObservationID *string                   `json:"observationId,omitempty"`
	Name          string                    `json:"name"`
	Value         interface{}               `json:"value"`
	DataType      commonTypes.ScoreDataType `json:"dataType"`
	Comment       *string                   `json:"comment,omitempty"`
	ConfigID      *string                   `json:"configId,omitempty"`
}

// CreateScoreResponse represents the response from creating a score
type CreateScoreResponse struct {
	ID            string                    `json:"id"`
	TraceID       string                    `json:"traceId"`
	ObservationID *string                   `json:"observationId,omitempty"`
	Name          string                    `json:"name"`
	Value         interface{}               `json:"value"`
	DataType      commonTypes.ScoreDataType `json:"dataType"`
	Comment       *string                   `json:"comment,omitempty"`
	ConfigID      *string                   `json:"configId,omitempty"`
	Timestamp     time.Time                 `json:"timestamp"`
	CreatedAt     time.Time                 `json:"createdAt"`
	UpdatedAt     time.Time                 `json:"updatedAt"`
}

// Validate validates the create score request
func (req *CreateScoreRequest) Validate() error {
	if req.TraceID == "" {
		return &ValidationError{Field: "traceId", Message: "traceId is required"}
	}

	if req.Name == "" {
		return &ValidationError{Field: "name", Message: "name is required"}
	}

	if req.Value == nil {
		return &ValidationError{Field: "value", Message: "value is required"}
	}

	// Validate data type and value consistency
	if err := validateValueAndDataType(req.Value, req.DataType); err != nil {
		return err
	}

	return nil
}

// ToCommonScore converts the request to a common Score type
func (req *CreateScoreRequest) ToCommonScore() *commonTypes.Score {
	score := &commonTypes.Score{
		TraceID:       req.TraceID,
		ObservationID: req.ObservationID,
		Name:          req.Name,
		DataType:      req.DataType,
		Comment:       req.Comment,
		ConfigID:      req.ConfigID,
		Timestamp:     time.Now().UTC(),
	}

	// Convert interface{} to json.RawMessage for Value
	if req.Value != nil {
		if valueBytes, err := json.Marshal(req.Value); err == nil {
			score.Value = json.RawMessage(valueBytes)
		}
	}

	if req.ID != nil {
		score.ID = *req.ID
	}

	return score
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// validateValueAndDataType validates that the value matches the specified data type
func validateValueAndDataType(value interface{}, dataType commonTypes.ScoreDataType) error {
	switch dataType {
	case commonTypes.ScoreDataTypeNumeric:
		switch value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			// Valid numeric types
		default:
			return &ValidationError{Field: "value", Message: "value must be numeric for NUMERIC data type"}
		}
	case commonTypes.ScoreDataTypeBoolean:
		if _, ok := value.(bool); !ok {
			return &ValidationError{Field: "value", Message: "value must be boolean for BOOLEAN data type"}
		}
	case commonTypes.ScoreDataTypeCategorical:
		if _, ok := value.(string); !ok {
			return &ValidationError{Field: "value", Message: "value must be string for CATEGORICAL data type"}
		}
	default:
		return &ValidationError{Field: "dataType", Message: "invalid score data type"}
	}

	return nil
}

// Helper functions for creating typed score requests

// NewNumericScoreRequest creates a request for a numeric score
func NewNumericScoreRequest(traceID, name string, value float64) *CreateScoreRequest {
	return &CreateScoreRequest{
		TraceID:  traceID,
		Name:     name,
		Value:    value,
		DataType: commonTypes.ScoreDataTypeNumeric,
	}
}

// NewBooleanScoreRequest creates a request for a boolean score
func NewBooleanScoreRequest(traceID, name string, value bool) *CreateScoreRequest {
	return &CreateScoreRequest{
		TraceID:  traceID,
		Name:     name,
		Value:    value,
		DataType: commonTypes.ScoreDataTypeBoolean,
	}
}

// NewCategoricalScoreRequest creates a request for a categorical score
func NewCategoricalScoreRequest(traceID, name, value string) *CreateScoreRequest {
	return &CreateScoreRequest{
		TraceID:  traceID,
		Name:     name,
		Value:    value,
		DataType: commonTypes.ScoreDataTypeCategorical,
	}
}

// WithObservation adds an observation ID to the score request
func (req *CreateScoreRequest) WithObservation(observationID string) *CreateScoreRequest {
	req.ObservationID = &observationID
	return req
}

// WithComment adds a comment to the score request
func (req *CreateScoreRequest) WithComment(comment string) *CreateScoreRequest {
	req.Comment = &comment
	return req
}

// WithConfig adds a config ID to the score request
func (req *CreateScoreRequest) WithConfig(configID string) *CreateScoreRequest {
	req.ConfigID = &configID
	return req
}

// WithID sets the ID for the score request
func (req *CreateScoreRequest) WithID(id string) *CreateScoreRequest {
	req.ID = &id
	return req
}
