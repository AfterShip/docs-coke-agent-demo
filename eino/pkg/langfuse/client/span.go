package client

import (
	"context"
	"time"

	"eino/pkg/langfuse/api/resources/commons/types"
	ingestiontypes "eino/pkg/langfuse/api/resources/ingestion/types"
	"eino/pkg/langfuse/internal/utils"
)

// SpanBuilder provides a fluent API for building span observations
type SpanBuilder struct {
	id                   string
	traceID              string
	parentObservationID  *string
	name                 string
	startTime            time.Time
	endTime              *time.Time
	input                interface{}
	output               interface{}
	metadata             map[string]interface{}
	level                types.ObservationLevel
	statusMessage        *string
	version              *string
	client               *Langfuse
	submitted            bool
}

// NewSpanBuilder creates a new SpanBuilder instance
func NewSpanBuilder(client *Langfuse, traceID string) *SpanBuilder {
	return &SpanBuilder{
		id:        utils.GenerateObservationID(),
		traceID:   traceID,
		startTime: time.Now().UTC(),
		level:     types.ObservationLevelDefault,
		client:    client,
		metadata:  make(map[string]interface{}),
	}
}

// ID sets the span ID
func (sb *SpanBuilder) ID(id string) *SpanBuilder {
	if sb.submitted {
		return sb
	}
	sb.id = id
	return sb
}

// ParentObservationID sets the parent observation ID
func (sb *SpanBuilder) ParentObservationID(parentID string) *SpanBuilder {
	if sb.submitted {
		return sb
	}
	sb.parentObservationID = &parentID
	return sb
}

// Name sets the span name
func (sb *SpanBuilder) Name(name string) *SpanBuilder {
	if sb.submitted {
		return sb
	}
	sb.name = name
	return sb
}

// StartTime sets the start time
func (sb *SpanBuilder) StartTime(startTime time.Time) *SpanBuilder {
	if sb.submitted {
		return sb
	}
	sb.startTime = startTime.UTC()
	return sb
}

// EndTime sets the end time
func (sb *SpanBuilder) EndTime(endTime time.Time) *SpanBuilder {
	if sb.submitted {
		return sb
	}
	endTimeUTC := endTime.UTC()
	sb.endTime = &endTimeUTC
	return sb
}

// Input sets the input data
func (sb *SpanBuilder) Input(input interface{}) *SpanBuilder {
	if sb.submitted {
		return sb
	}
	sb.input = input
	return sb
}

// Output sets the output data
func (sb *SpanBuilder) Output(output interface{}) *SpanBuilder {
	if sb.submitted {
		return sb
	}
	sb.output = output
	return sb
}

// Metadata sets the metadata map
func (sb *SpanBuilder) Metadata(metadata map[string]interface{}) *SpanBuilder {
	if sb.submitted {
		return sb
	}
	sb.metadata = metadata
	return sb
}

// AddMetadata adds a single metadata key-value pair
func (sb *SpanBuilder) AddMetadata(key string, value interface{}) *SpanBuilder {
	if sb.submitted {
		return sb
	}
	if sb.metadata == nil {
		sb.metadata = make(map[string]interface{})
	}
	sb.metadata[key] = value
	return sb
}

// Level sets the observation level
func (sb *SpanBuilder) Level(level types.ObservationLevel) *SpanBuilder {
	if sb.submitted {
		return sb
	}
	sb.level = level
	return sb
}

// Debug sets the observation level to DEBUG
func (sb *SpanBuilder) Debug() *SpanBuilder {
	return sb.Level(types.ObservationLevelDebug)
}

// Warning sets the observation level to WARNING
func (sb *SpanBuilder) Warning() *SpanBuilder {
	return sb.Level(types.ObservationLevelWarning)
}

// Error sets the observation level to ERROR
func (sb *SpanBuilder) Error() *SpanBuilder {
	return sb.Level(types.ObservationLevelError)
}

// StatusMessage sets the status message
func (sb *SpanBuilder) StatusMessage(message string) *SpanBuilder {
	if sb.submitted {
		return sb
	}
	sb.statusMessage = &message
	return sb
}

// Version sets the version
func (sb *SpanBuilder) Version(version string) *SpanBuilder {
	if sb.submitted {
		return sb
	}
	sb.version = &version
	return sb
}

// GetID returns the span ID
func (sb *SpanBuilder) GetID() string {
	return sb.id
}

// GetTraceID returns the trace ID
func (sb *SpanBuilder) GetTraceID() string {
	return sb.traceID
}

// GetName returns the span name
func (sb *SpanBuilder) GetName() string {
	return sb.name
}

// WithStartTime is an alias for StartTime for fluent API
func (sb *SpanBuilder) WithStartTime(startTime time.Time) *SpanBuilder {
	return sb.StartTime(startTime)
}

// WithEndTime is an alias for EndTime for fluent API
func (sb *SpanBuilder) WithEndTime(endTime time.Time) *SpanBuilder {
	return sb.EndTime(endTime)
}

// WithInput is an alias for Input for fluent API
func (sb *SpanBuilder) WithInput(input interface{}) *SpanBuilder {
	return sb.Input(input)
}

// WithOutput is an alias for Output for fluent API
func (sb *SpanBuilder) WithOutput(output interface{}) *SpanBuilder {
	return sb.Output(output)
}

// WithMetadata is an alias for Metadata for fluent API
func (sb *SpanBuilder) WithMetadata(metadata map[string]interface{}) *SpanBuilder {
	return sb.Metadata(metadata)
}

// WithLevel is an alias for Level for fluent API (accepts string)
func (sb *SpanBuilder) WithLevel(level string) *SpanBuilder {
	switch level {
	case "DEBUG":
		return sb.Level(types.ObservationLevelDebug)
	case "WARNING":
		return sb.Level(types.ObservationLevelWarning)
	case "ERROR":
		return sb.Level(types.ObservationLevelError)
	default:
		return sb.Level(types.ObservationLevelDefault)
	}
}

// WithStatusMessage is an alias for StatusMessage for fluent API
func (sb *SpanBuilder) WithStatusMessage(message string) *SpanBuilder {
	return sb.StatusMessage(message)
}

// ChildSpan creates a child span (placeholder - needs full implementation)
func (sb *SpanBuilder) ChildSpan(name string) *SpanBuilder {
	childSpan := NewSpanBuilder(sb.client, sb.traceID)
	childSpan.ParentObservationID(sb.id)
	return childSpan.Name(name)
}

// validate performs validation on the span builder
func (sb *SpanBuilder) validate() error {
	if sb.id == "" {
		return &ValidationError{Field: "id", Message: "span id is required"}
	}
	
	if sb.traceID == "" {
		return &ValidationError{Field: "traceId", Message: "trace id is required"}
	}
	
	if sb.name == "" {
		return &ValidationError{Field: "name", Message: "span name is required"}
	}
	
	if sb.startTime.IsZero() {
		return &ValidationError{Field: "startTime", Message: "start time is required"}
	}
	
	// Validate end time if set
	if sb.endTime != nil && sb.endTime.Before(sb.startTime) {
		return &ValidationError{Field: "endTime", Message: "end time cannot be before start time"}
	}
	
	return nil
}

// toObservationEvent converts the builder to an ObservationEvent
func (sb *SpanBuilder) toObservationEvent() *ingestiontypes.ObservationEvent {
	return &ingestiontypes.ObservationEvent{
		ID:                  sb.id,
		TraceID:             sb.traceID,
		ParentObservationID: sb.parentObservationID,
		Type:                types.ObservationTypeSpan,
		Name:                sb.name,
		StartTime:           sb.startTime,
		EndTime:             sb.endTime,
		Input:               sb.input,
		Output:              sb.output,
		Metadata:            sb.metadata,
		Level:               sb.level,
		StatusMessage:       sb.statusMessage,
		Version:             sb.version,
	}
}

// toSpanCreateEvent converts the builder to a SpanCreateEvent
func (sb *SpanBuilder) toSpanCreateEvent() *ingestiontypes.SpanCreateEvent {
	return &ingestiontypes.SpanCreateEvent{
		ObservationEvent: *sb.toObservationEvent(),
		EventType:        "span-create",
	}
}

// toSpanUpdateEvent converts the builder to a SpanUpdateEvent
func (sb *SpanBuilder) toSpanUpdateEvent() *ingestiontypes.SpanUpdateEvent {
	return &ingestiontypes.SpanUpdateEvent{
		ObservationEvent: *sb.toObservationEvent(),
		EventType:        "span-update",
	}
}

// Submit submits the span to the ingestion queue
func (sb *SpanBuilder) Submit(ctx context.Context) error {
	if sb.submitted {
		return &ValidationError{Field: "state", Message: "span already submitted"}
	}
	
	if err := sb.validate(); err != nil {
		return err
	}
	
	event := sb.toSpanCreateEvent()
	ingestionEvent := event.ToIngestionEvent()
	
	if err := sb.client.queue.Enqueue(ingestionEvent); err != nil {
		return err
	}
	
	sb.submitted = true
	return nil
}

// Update updates an existing span
func (sb *SpanBuilder) Update(ctx context.Context) error {
	if sb.submitted {
		return &ValidationError{Field: "state", Message: "span already submitted"}
	}
	
	if err := sb.validate(); err != nil {
		return err
	}
	
	event := sb.toSpanUpdateEvent()
	ingestionEvent := event.ToIngestionEvent()
	
	if err := sb.client.queue.Enqueue(ingestionEvent); err != nil {
		return err
	}
	
	sb.submitted = true
	return nil
}

// End ends the span with the current timestamp and submits it
func (sb *SpanBuilder) End(ctx context.Context) error {
	return sb.EndAt(ctx, time.Now().UTC())
}

// EndAt ends the span with a specific timestamp and submits it
func (sb *SpanBuilder) EndAt(ctx context.Context, endTime time.Time) error {
	sb.EndTime(endTime)
	return sb.Update(ctx)
}