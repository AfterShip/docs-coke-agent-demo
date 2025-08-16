package types

import (
	"encoding/json"
	"time"

	"eino/pkg/langfuse/api/resources/commons/types"
)

// TraceEvent represents a trace event in the ingestion system
type TraceEvent struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	UserID         *string               `json:"userId,omitempty"`
	SessionID      *string               `json:"sessionId,omitempty"`
	Input          interface{}           `json:"input,omitempty"`
	Output         interface{}           `json:"output,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Tags           []string              `json:"tags,omitempty"`
	Environment    string                `json:"environment,omitempty"`
	Release        *string               `json:"release,omitempty"`
	Version        *string               `json:"version,omitempty"`
	Public         *bool                 `json:"public,omitempty"`
	Timestamp      time.Time             `json:"timestamp"`
	ParentTraceID  *string               `json:"parentTraceId,omitempty"`
}

// TraceCreateEvent represents a trace creation event
type TraceCreateEvent struct {
	TraceEvent
	Type string `json:"type"` // "trace-create"
}

// TraceUpdateEvent represents a trace update event
type TraceUpdateEvent struct {
	TraceEvent
	Type   string     `json:"type"`   // "trace-update"
	EndTime *time.Time `json:"endTime,omitempty"`
}

// NewTraceEvent creates a new trace event from a Trace struct
func NewTraceEvent(trace *types.Trace) *TraceEvent {
	var name string
	if trace.Name != nil {
		name = *trace.Name
	}
	
	return &TraceEvent{
		ID:          trace.ID,
		Name:        name,
		UserID:      trace.UserID,
		SessionID:   trace.SessionID,
		Input:       trace.Input,
		Output:      trace.Output,
		Metadata:    trace.Metadata,
		Tags:        trace.Tags,
		Release:     trace.Release,
		Version:     trace.Version,
		Public:      trace.Public,
		Timestamp:   trace.Timestamp,
	}
}

// NewTraceCreateEvent creates a new trace creation event
func NewTraceCreateEvent(trace *types.Trace) *TraceCreateEvent {
	return &TraceCreateEvent{
		TraceEvent: *NewTraceEvent(trace),
		Type:       "trace-create",
	}
}

// NewTraceUpdateEvent creates a new trace update event
func NewTraceUpdateEvent(trace *types.Trace, endTime *time.Time) *TraceUpdateEvent {
	return &TraceUpdateEvent{
		TraceEvent: *NewTraceEvent(trace),
		Type:       "trace-update",
		EndTime:    endTime,
	}
}

// ToIngestionEvent converts TraceCreateEvent to IngestionEvent
func (e *TraceCreateEvent) ToIngestionEvent() IngestionEvent {
	return IngestionEvent{
		ID:        e.ID,
		Type:      EventTypeTraceCreate,
		Timestamp: e.Timestamp,
		Body:      e,
	}
}

// ToIngestionEvent converts TraceUpdateEvent to IngestionEvent
func (e *TraceUpdateEvent) ToIngestionEvent() IngestionEvent {
	return IngestionEvent{
		ID:        e.ID,
		Type:      EventTypeTraceUpdate,
		Timestamp: e.Timestamp,
		Body:      e,
	}
}

// MarshalJSON implements json.Marshaler for TraceEvent
func (e *TraceEvent) MarshalJSON() ([]byte, error) {
	type Alias TraceEvent
	return json.Marshal(&struct {
		*Alias
		Timestamp string `json:"timestamp"`
	}{
		Alias:     (*Alias)(e),
		Timestamp: e.Timestamp.UTC().Format(time.RFC3339Nano),
	})
}

// UnmarshalJSON implements json.Unmarshaler for TraceEvent
func (e *TraceEvent) UnmarshalJSON(data []byte) error {
	type Alias TraceEvent
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

// MarshalJSON implements json.Marshaler for TraceUpdateEvent
func (e *TraceUpdateEvent) MarshalJSON() ([]byte, error) {
	type Alias TraceUpdateEvent
	aux := &struct {
		*Alias
		EndTime *string `json:"endTime,omitempty"`
	}{
		Alias: (*Alias)(e),
	}
	
	if e.EndTime != nil {
		endTimeStr := e.EndTime.UTC().Format(time.RFC3339Nano)
		aux.EndTime = &endTimeStr
	}
	
	return json.Marshal(aux)
}

// UnmarshalJSON implements json.Unmarshaler for TraceUpdateEvent
func (e *TraceUpdateEvent) UnmarshalJSON(data []byte) error {
	type Alias TraceUpdateEvent
	aux := &struct {
		*Alias
		EndTime *string `json:"endTime,omitempty"`
	}{
		Alias: (*Alias)(e),
	}
	
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	
	if aux.EndTime != nil {
		endTime, err := time.Parse(time.RFC3339Nano, *aux.EndTime)
		if err != nil {
			return err
		}
		e.EndTime = &endTime
	}
	
	return nil
}

// Validate performs validation on TraceEvent
func (e *TraceEvent) Validate() error {
	if e.ID == "" {
		return &ValidationError{Field: "id", Message: "trace id is required"}
	}
	
	if e.Name == "" {
		return &ValidationError{Field: "name", Message: "trace name is required"}
	}
	
	if e.Timestamp.IsZero() {
		return &ValidationError{Field: "timestamp", Message: "trace timestamp is required"}
	}
	
	return nil
}

// Validate performs validation on TraceCreateEvent
func (e *TraceCreateEvent) Validate() error {
	if err := e.TraceEvent.Validate(); err != nil {
		return err
	}
	
	if e.Type != "trace-create" {
		return &ValidationError{Field: "type", Message: "event type must be 'trace-create'"}
	}
	
	return nil
}

// Validate performs validation on TraceUpdateEvent
func (e *TraceUpdateEvent) Validate() error {
	if err := e.TraceEvent.Validate(); err != nil {
		return err
	}
	
	if e.Type != "trace-update" {
		return &ValidationError{Field: "type", Message: "event type must be 'trace-update'"}
	}
	
	// EndTime validation if provided
	if e.EndTime != nil && e.EndTime.Before(e.Timestamp) {
		return &ValidationError{Field: "endTime", Message: "end time cannot be before start time"}
	}
	
	return nil
}

// WithUserID sets the user ID for the trace event
func (e *TraceEvent) WithUserID(userID string) *TraceEvent {
	e.UserID = &userID
	return e
}

// WithSessionID sets the session ID for the trace event
func (e *TraceEvent) WithSessionID(sessionID string) *TraceEvent {
	e.SessionID = &sessionID
	return e
}

// WithInput sets the input for the trace event
func (e *TraceEvent) WithInput(input interface{}) *TraceEvent {
	e.Input = input
	return e
}

// WithOutput sets the output for the trace event
func (e *TraceEvent) WithOutput(output interface{}) *TraceEvent {
	e.Output = output
	return e
}

// WithMetadata sets the metadata for the trace event
func (e *TraceEvent) WithMetadata(metadata map[string]interface{}) *TraceEvent {
	e.Metadata = metadata
	return e
}

// WithTags sets the tags for the trace event
func (e *TraceEvent) WithTags(tags ...string) *TraceEvent {
	e.Tags = tags
	return e
}

// WithEnvironment sets the environment for the trace event
func (e *TraceEvent) WithEnvironment(environment string) *TraceEvent {
	e.Environment = environment
	return e
}

// WithRelease sets the release for the trace event
func (e *TraceEvent) WithRelease(release string) *TraceEvent {
	e.Release = &release
	return e
}

// WithVersion sets the version for the trace event
func (e *TraceEvent) WithVersion(version string) *TraceEvent {
	e.Version = &version
	return e
}

// WithPublic sets the public flag for the trace event
func (e *TraceEvent) WithPublic(public bool) *TraceEvent {
	e.Public = &public
	return e
}