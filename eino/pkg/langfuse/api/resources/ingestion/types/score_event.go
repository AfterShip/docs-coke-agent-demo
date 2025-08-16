package types

import (
	"encoding/json"
	"time"

	"eino/pkg/langfuse/api/resources/commons/types"
)

// ScoreEvent represents a score event in the ingestion system
type ScoreEvent struct {
	ID            string                 `json:"id"`
	TraceID       string                 `json:"traceId"`
	ObservationID *string               `json:"observationId,omitempty"`
	Name          string                `json:"name"`
	Value         interface{}           `json:"value"`
	DataType      types.ScoreDataType   `json:"dataType"`
	Comment       *string               `json:"comment,omitempty"`
	ConfigID      *string               `json:"configId,omitempty"`
	Timestamp     time.Time             `json:"timestamp"`
	Source        ScoreSource           `json:"source,omitempty"`
	AuthorUserID  *string               `json:"authorUserId,omitempty"`
	QueueID       *string               `json:"queueId,omitempty"`
}

// ScoreCreateEvent represents a score creation event
type ScoreCreateEvent struct {
	ScoreEvent
	EventType string `json:"type"` // "score-create"
}

// ScoreSource represents the source of a score
type ScoreSource string

const (
	ScoreSourceAPI       ScoreSource = "API"
	ScoreSourceSDK       ScoreSource = "SDK"
	ScoreSourceUI        ScoreSource = "UI"
	ScoreSourceWorkflow  ScoreSource = "WORKFLOW"
	ScoreSourceEval      ScoreSource = "EVAL"
	ScoreSourceAnnotation ScoreSource = "ANNOTATION"
)

// NewScoreEvent creates a new score event from a Score struct
func NewScoreEvent(score *types.Score) *ScoreEvent {
	return &ScoreEvent{
		ID:            score.ID,
		TraceID:       score.TraceID,
		ObservationID: score.ObservationID,
		Name:          score.Name,
		Value:         score.Value,
		DataType:      score.DataType,
		Comment:       score.Comment,
		ConfigID:      score.ConfigID,
		Timestamp:     score.Timestamp,
		Source:        ScoreSourceSDK, // Default to SDK for events created from SDK
	}
}

// NewScoreCreateEvent creates a new score creation event
func NewScoreCreateEvent(score *types.Score) *ScoreCreateEvent {
	return &ScoreCreateEvent{
		ScoreEvent: *NewScoreEvent(score),
		EventType:  "score-create",
	}
}

// ToIngestionEvent converts ScoreCreateEvent to IngestionEvent
func (e *ScoreCreateEvent) ToIngestionEvent() IngestionEvent {
	return IngestionEvent{
		ID:        e.ID,
		Type:      EventTypeScoreCreate,
		Timestamp: e.Timestamp,
		Body:      e,
	}
}

// MarshalJSON implements json.Marshaler for ScoreEvent
func (e *ScoreEvent) MarshalJSON() ([]byte, error) {
	type Alias ScoreEvent
	return json.Marshal(&struct {
		*Alias
		Timestamp string `json:"timestamp"`
	}{
		Alias:     (*Alias)(e),
		Timestamp: e.Timestamp.UTC().Format(time.RFC3339Nano),
	})
}

// UnmarshalJSON implements json.Unmarshaler for ScoreEvent
func (e *ScoreEvent) UnmarshalJSON(data []byte) error {
	type Alias ScoreEvent
	aux := &struct {
		*Alias
		Timestamp string `json:"timestamp"`
	}{
		Alias: (*Alias)(e),
	}
	
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	
	var err error
	e.Timestamp, err = time.Parse(time.RFC3339Nano, aux.Timestamp)
	return err
}

// Validate performs validation on ScoreEvent
func (e *ScoreEvent) Validate() error {
	if e.ID == "" {
		return &ValidationError{Field: "id", Message: "score id is required"}
	}
	
	if e.TraceID == "" {
		return &ValidationError{Field: "traceId", Message: "trace id is required"}
	}
	
	if e.Name == "" {
		return &ValidationError{Field: "name", Message: "score name is required"}
	}
	
	if e.Value == nil {
		return &ValidationError{Field: "value", Message: "score value is required"}
	}
	
	if e.Timestamp.IsZero() {
		return &ValidationError{Field: "timestamp", Message: "timestamp is required"}
	}
	
	// Validate data type and value consistency
	if err := e.validateValueAndDataType(); err != nil {
		return err
	}
	
	return nil
}

// validateValueAndDataType validates that the value matches the specified data type
func (e *ScoreEvent) validateValueAndDataType() error {
	switch e.DataType {
	case types.ScoreDataTypeNumeric:
		switch e.Value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			// Valid numeric types
		case json.Number:
			// JSON numbers are also valid
		default:
			return &ValidationError{Field: "value", Message: "value must be numeric for NUMERIC data type"}
		}
	case types.ScoreDataTypeBoolean:
		if _, ok := e.Value.(bool); !ok {
			return &ValidationError{Field: "value", Message: "value must be boolean for BOOLEAN data type"}
		}
	case types.ScoreDataTypeCategorical:
		if _, ok := e.Value.(string); !ok {
			return &ValidationError{Field: "value", Message: "value must be string for CATEGORICAL data type"}
		}
	default:
		return &ValidationError{Field: "dataType", Message: "invalid score data type"}
	}
	
	return nil
}

// Validate performs validation on ScoreCreateEvent
func (e *ScoreCreateEvent) Validate() error {
	if err := e.ScoreEvent.Validate(); err != nil {
		return err
	}
	
	if e.EventType != "score-create" {
		return &ValidationError{Field: "type", Message: "event type must be 'score-create'"}
	}
	
	return nil
}

// Helper methods for score event building
func (e *ScoreEvent) WithObservationID(observationID string) *ScoreEvent {
	e.ObservationID = &observationID
	return e
}

func (e *ScoreEvent) WithComment(comment string) *ScoreEvent {
	e.Comment = &comment
	return e
}

func (e *ScoreEvent) WithConfigID(configID string) *ScoreEvent {
	e.ConfigID = &configID
	return e
}

func (e *ScoreEvent) WithSource(source ScoreSource) *ScoreEvent {
	e.Source = source
	return e
}

func (e *ScoreEvent) WithAuthorUserID(userID string) *ScoreEvent {
	e.AuthorUserID = &userID
	return e
}

func (e *ScoreEvent) WithQueueID(queueID string) *ScoreEvent {
	e.QueueID = &queueID
	return e
}

// ScoreValue helper functions for type-safe value creation
func NumericScore(value float64) interface{} {
	return value
}

func BooleanScore(value bool) interface{} {
	return value
}

func CategoricalScore(value string) interface{} {
	return value
}

// CreateNumericScoreEvent creates a numeric score event
func CreateNumericScoreEvent(id, traceID, name string, value float64) *ScoreCreateEvent {
	scoreEvent := &ScoreEvent{
		ID:        id,
		TraceID:   traceID,
		Name:      name,
		Value:     value,
		DataType:  types.ScoreDataTypeNumeric,
		Timestamp: time.Now().UTC(),
		Source:    ScoreSourceSDK,
	}
	
	return &ScoreCreateEvent{
		ScoreEvent: *scoreEvent,
		EventType:  "score-create",
	}
}

// CreateBooleanScoreEvent creates a boolean score event
func CreateBooleanScoreEvent(id, traceID, name string, value bool) *ScoreCreateEvent {
	scoreEvent := &ScoreEvent{
		ID:        id,
		TraceID:   traceID,
		Name:      name,
		Value:     value,
		DataType:  types.ScoreDataTypeBoolean,
		Timestamp: time.Now().UTC(),
		Source:    ScoreSourceSDK,
	}
	
	return &ScoreCreateEvent{
		ScoreEvent: *scoreEvent,
		EventType:  "score-create",
	}
}

// CreateCategoricalScoreEvent creates a categorical score event
func CreateCategoricalScoreEvent(id, traceID, name, value string) *ScoreCreateEvent {
	scoreEvent := &ScoreEvent{
		ID:        id,
		TraceID:   traceID,
		Name:      name,
		Value:     value,
		DataType:  types.ScoreDataTypeCategorical,
		Timestamp: time.Now().UTC(),
		Source:    ScoreSourceSDK,
	}
	
	return &ScoreCreateEvent{
		ScoreEvent: *scoreEvent,
		EventType:  "score-create",
	}
}