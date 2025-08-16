package types

import (
	"encoding/json"
	"time"

	"eino/pkg/langfuse/api/resources/commons/types"
)

// ObservationEvent represents an observation event in the ingestion system
type ObservationEvent struct {
	ID                   string                 `json:"id"`
	TraceID              string                 `json:"traceId"`
	ParentObservationID  *string               `json:"parentObservationId,omitempty"`
	Type                 types.ObservationType `json:"type"`
	Name                 string                `json:"name"`
	StartTime            time.Time             `json:"startTime"`
	EndTime              *time.Time            `json:"endTime,omitempty"`
	CompletionStartTime  *time.Time            `json:"completionStartTime,omitempty"`
	Model                *string               `json:"model,omitempty"`
	ModelParameters      map[string]interface{} `json:"modelParameters,omitempty"`
	Input                interface{}           `json:"input,omitempty"`
	Output               interface{}           `json:"output,omitempty"`
	Usage                *types.Usage          `json:"usage,omitempty"`
	Level                types.ObservationLevel `json:"level,omitempty"`
	StatusMessage        *string               `json:"statusMessage,omitempty"`
	Version              *string               `json:"version,omitempty"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
	Environment          string                `json:"environment,omitempty"`
}

// ObservationCreateEvent represents an observation creation event
type ObservationCreateEvent struct {
	ObservationEvent
	EventType string `json:"type"` // "observation-create"
}

// ObservationUpdateEvent represents an observation update event
type ObservationUpdateEvent struct {
	ObservationEvent
	EventType string `json:"type"` // "observation-update"
}

// SpanCreateEvent represents a span creation event
type SpanCreateEvent struct {
	ObservationEvent
	EventType string `json:"type"` // "span-create"
}

// SpanUpdateEvent represents a span update event
type SpanUpdateEvent struct {
	ObservationEvent
	EventType string `json:"type"` // "span-update"
}

// GenerationCreateEvent represents a generation creation event
type GenerationCreateEvent struct {
	ObservationEvent
	EventType string `json:"type"` // "generation-create"
}

// GenerationUpdateEvent represents a generation update event
type GenerationUpdateEvent struct {
	ObservationEvent
	EventType string `json:"type"` // "generation-update"
}

// EventCreateEvent represents an event creation
type EventCreateEvent struct {
	ObservationEvent
	EventType string `json:"type"` // "event-create"
}

// NewObservationEvent creates a new observation event from an Observation struct
func NewObservationEvent(observation *types.Observation) *ObservationEvent {
	var name string
	if observation.Name != nil {
		name = *observation.Name
	}
	
	var level types.ObservationLevel
	if observation.Level != nil {
		level = *observation.Level
	}
	
	return &ObservationEvent{
		ID:                   observation.ID,
		TraceID:              observation.TraceID,
		ParentObservationID:  observation.ParentObservationID,
		Type:                 observation.Type,
		Name:                 name,
		StartTime:            observation.StartTime,
		EndTime:              observation.EndTime,
		CompletionStartTime:  observation.CompletionStartTime,
		Model:                observation.Model,
		ModelParameters:      observation.ModelParameters,
		Input:                observation.Input,
		Output:               observation.Output,
		Usage:                observation.Usage,
		Level:                level,
		StatusMessage:        observation.StatusMessage,
		Version:              observation.Version,
		Metadata:             observation.Metadata,
	}
}

// NewObservationCreateEvent creates a new observation creation event
func NewObservationCreateEvent(observation *types.Observation) *ObservationCreateEvent {
	return &ObservationCreateEvent{
		ObservationEvent: *NewObservationEvent(observation),
		EventType:        "observation-create",
	}
}

// NewObservationUpdateEvent creates a new observation update event
func NewObservationUpdateEvent(observation *types.Observation) *ObservationUpdateEvent {
	return &ObservationUpdateEvent{
		ObservationEvent: *NewObservationEvent(observation),
		EventType:        "observation-update",
	}
}

// NewSpanCreateEvent creates a new span creation event
func NewSpanCreateEvent(observation *types.Observation) *SpanCreateEvent {
	return &SpanCreateEvent{
		ObservationEvent: *NewObservationEvent(observation),
		EventType:        "span-create",
	}
}

// NewSpanUpdateEvent creates a new span update event
func NewSpanUpdateEvent(observation *types.Observation) *SpanUpdateEvent {
	return &SpanUpdateEvent{
		ObservationEvent: *NewObservationEvent(observation),
		EventType:        "span-update",
	}
}

// NewGenerationCreateEvent creates a new generation creation event
func NewGenerationCreateEvent(observation *types.Observation) *GenerationCreateEvent {
	return &GenerationCreateEvent{
		ObservationEvent: *NewObservationEvent(observation),
		EventType:        "generation-create",
	}
}

// NewGenerationUpdateEvent creates a new generation update event
func NewGenerationUpdateEvent(observation *types.Observation) *GenerationUpdateEvent {
	return &GenerationUpdateEvent{
		ObservationEvent: *NewObservationEvent(observation),
		EventType:        "generation-update",
	}
}

// NewEventCreateEvent creates a new event creation
func NewEventCreateEvent(observation *types.Observation) *EventCreateEvent {
	return &EventCreateEvent{
		ObservationEvent: *NewObservationEvent(observation),
		EventType:        "event-create",
	}
}

// ToIngestionEvent implementations
func (e *ObservationCreateEvent) ToIngestionEvent() IngestionEvent {
	return IngestionEvent{
		ID:        e.ID,
		Type:      EventTypeObservationCreate,
		Timestamp: e.StartTime,
		Body:      e,
	}
}

func (e *ObservationUpdateEvent) ToIngestionEvent() IngestionEvent {
	return IngestionEvent{
		ID:        e.ID,
		Type:      EventTypeObservationUpdate,
		Timestamp: e.StartTime,
		Body:      e,
	}
}

func (e *SpanCreateEvent) ToIngestionEvent() IngestionEvent {
	return IngestionEvent{
		ID:        e.ID,
		Type:      EventTypeSpanCreate,
		Timestamp: e.StartTime,
		Body:      e,
	}
}

func (e *SpanUpdateEvent) ToIngestionEvent() IngestionEvent {
	return IngestionEvent{
		ID:        e.ID,
		Type:      EventTypeSpanUpdate,
		Timestamp: e.StartTime,
		Body:      e,
	}
}

func (e *GenerationCreateEvent) ToIngestionEvent() IngestionEvent {
	return IngestionEvent{
		ID:        e.ID,
		Type:      EventTypeGenerationCreate,
		Timestamp: e.StartTime,
		Body:      e,
	}
}

func (e *GenerationUpdateEvent) ToIngestionEvent() IngestionEvent {
	return IngestionEvent{
		ID:        e.ID,
		Type:      EventTypeGenerationUpdate,
		Timestamp: e.StartTime,
		Body:      e,
	}
}

func (e *EventCreateEvent) ToIngestionEvent() IngestionEvent {
	return IngestionEvent{
		ID:        e.ID,
		Type:      EventTypeEventCreate,
		Timestamp: e.StartTime,
		Body:      e,
	}
}

// JSON marshalling with proper time format
func (e *ObservationEvent) MarshalJSON() ([]byte, error) {
	type Alias ObservationEvent
	aux := &struct {
		*Alias
		StartTime           string  `json:"startTime"`
		EndTime             *string `json:"endTime,omitempty"`
		CompletionStartTime *string `json:"completionStartTime,omitempty"`
	}{
		Alias:     (*Alias)(e),
		StartTime: e.StartTime.UTC().Format(time.RFC3339Nano),
	}
	
	if e.EndTime != nil {
		endTimeStr := e.EndTime.UTC().Format(time.RFC3339Nano)
		aux.EndTime = &endTimeStr
	}
	
	if e.CompletionStartTime != nil {
		completionStartTimeStr := e.CompletionStartTime.UTC().Format(time.RFC3339Nano)
		aux.CompletionStartTime = &completionStartTimeStr
	}
	
	return json.Marshal(aux)
}

func (e *ObservationEvent) UnmarshalJSON(data []byte) error {
	type Alias ObservationEvent
	aux := &struct {
		*Alias
		StartTime           string  `json:"startTime"`
		EndTime             *string `json:"endTime,omitempty"`
		CompletionStartTime *string `json:"completionStartTime,omitempty"`
	}{
		Alias: (*Alias)(e),
	}
	
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	
	// Parse start time
	startTime, err := time.Parse(time.RFC3339Nano, aux.StartTime)
	if err != nil {
		return err
	}
	e.StartTime = startTime
	
	// Parse end time if present
	if aux.EndTime != nil {
		endTime, err := time.Parse(time.RFC3339Nano, *aux.EndTime)
		if err != nil {
			return err
		}
		e.EndTime = &endTime
	}
	
	// Parse completion start time if present
	if aux.CompletionStartTime != nil {
		completionStartTime, err := time.Parse(time.RFC3339Nano, *aux.CompletionStartTime)
		if err != nil {
			return err
		}
		e.CompletionStartTime = &completionStartTime
	}
	
	return nil
}

// Validation methods
func (e *ObservationEvent) Validate() error {
	if e.ID == "" {
		return &ValidationError{Field: "id", Message: "observation id is required"}
	}
	
	if e.TraceID == "" {
		return &ValidationError{Field: "traceId", Message: "trace id is required"}
	}
	
	if e.Name == "" {
		return &ValidationError{Field: "name", Message: "observation name is required"}
	}
	
	if e.StartTime.IsZero() {
		return &ValidationError{Field: "startTime", Message: "start time is required"}
	}
	
	// Validate observation type
	switch e.Type {
	case types.ObservationTypeSpan, types.ObservationTypeGeneration, types.ObservationTypeEvent:
		// Valid types
	default:
		return &ValidationError{Field: "type", Message: "invalid observation type"}
	}
	
	// Time validation
	if e.EndTime != nil && e.EndTime.Before(e.StartTime) {
		return &ValidationError{Field: "endTime", Message: "end time cannot be before start time"}
	}
	
	if e.CompletionStartTime != nil && e.CompletionStartTime.Before(e.StartTime) {
		return &ValidationError{Field: "completionStartTime", Message: "completion start time cannot be before start time"}
	}
	
	return nil
}

func (e *ObservationCreateEvent) Validate() error {
	if err := e.ObservationEvent.Validate(); err != nil {
		return err
	}
	
	if e.EventType != "observation-create" {
		return &ValidationError{Field: "type", Message: "event type must be 'observation-create'"}
	}
	
	return nil
}

func (e *ObservationUpdateEvent) Validate() error {
	if err := e.ObservationEvent.Validate(); err != nil {
		return err
	}
	
	if e.EventType != "observation-update" {
		return &ValidationError{Field: "type", Message: "event type must be 'observation-update'"}
	}
	
	return nil
}

// Helper methods for observation event
func (e *ObservationEvent) WithParent(parentID string) *ObservationEvent {
	e.ParentObservationID = &parentID
	return e
}

func (e *ObservationEvent) WithModel(model string) *ObservationEvent {
	e.Model = &model
	return e
}

func (e *ObservationEvent) WithModelParameters(params map[string]interface{}) *ObservationEvent {
	e.ModelParameters = params
	return e
}

func (e *ObservationEvent) WithInput(input interface{}) *ObservationEvent {
	e.Input = input
	return e
}

func (e *ObservationEvent) WithOutput(output interface{}) *ObservationEvent {
	e.Output = output
	return e
}

func (e *ObservationEvent) WithUsage(usage *types.Usage) *ObservationEvent {
	e.Usage = usage
	return e
}

func (e *ObservationEvent) WithLevel(level types.ObservationLevel) *ObservationEvent {
	e.Level = level
	return e
}

func (e *ObservationEvent) WithMetadata(metadata map[string]interface{}) *ObservationEvent {
	e.Metadata = metadata
	return e
}

func (e *ObservationEvent) WithEndTime(endTime time.Time) *ObservationEvent {
	e.EndTime = &endTime
	return e
}

func (e *ObservationEvent) WithCompletionStartTime(completionStartTime time.Time) *ObservationEvent {
	e.CompletionStartTime = &completionStartTime
	return e
}