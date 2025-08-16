package client

import (
	"context"
	"time"

	"eino/pkg/langfuse/api/resources/commons/types"
	ingestiontypes "eino/pkg/langfuse/api/resources/ingestion/types"
	"eino/pkg/langfuse/internal/utils"
)

// GenerationBuilder provides a fluent API for building LLM generation observations
type GenerationBuilder struct {
	id                   string
	traceID              string
	parentObservationID  *string
	name                 string
	startTime            time.Time
	endTime              *time.Time
	completionStartTime  *time.Time
	model                *string
	modelParameters      map[string]interface{}
	input                interface{}
	output               interface{}
	usage                *types.Usage
	metadata             map[string]interface{}
	level                types.ObservationLevel
	statusMessage        *string
	version              *string
	client               *Langfuse
	submitted            bool
}

// NewGenerationBuilder creates a new GenerationBuilder instance
func NewGenerationBuilder(client *Langfuse, traceID string) *GenerationBuilder {
	return &GenerationBuilder{
		id:              utils.GenerateObservationID(),
		traceID:         traceID,
		startTime:       time.Now().UTC(),
		level:           types.ObservationLevelDefault,
		client:          client,
		metadata:        make(map[string]interface{}),
		modelParameters: make(map[string]interface{}),
	}
}

// ID sets the generation ID
func (gb *GenerationBuilder) ID(id string) *GenerationBuilder {
	if gb.submitted {
		return gb
	}
	gb.id = id
	return gb
}

// ParentObservationID sets the parent observation ID
func (gb *GenerationBuilder) ParentObservationID(parentID string) *GenerationBuilder {
	if gb.submitted {
		return gb
	}
	gb.parentObservationID = &parentID
	return gb
}

// Name sets the generation name
func (gb *GenerationBuilder) Name(name string) *GenerationBuilder {
	if gb.submitted {
		return gb
	}
	gb.name = name
	return gb
}

// StartTime sets the start time
func (gb *GenerationBuilder) StartTime(startTime time.Time) *GenerationBuilder {
	if gb.submitted {
		return gb
	}
	gb.startTime = startTime.UTC()
	return gb
}

// EndTime sets the end time
func (gb *GenerationBuilder) EndTime(endTime time.Time) *GenerationBuilder {
	if gb.submitted {
		return gb
	}
	endTimeUTC := endTime.UTC()
	gb.endTime = &endTimeUTC
	return gb
}

// CompletionStartTime sets the completion start time for streaming responses
func (gb *GenerationBuilder) CompletionStartTime(completionStartTime time.Time) *GenerationBuilder {
	if gb.submitted {
		return gb
	}
	completionStartTimeUTC := completionStartTime.UTC()
	gb.completionStartTime = &completionStartTimeUTC
	return gb
}

// Model sets the model name
func (gb *GenerationBuilder) Model(model string) *GenerationBuilder {
	if gb.submitted {
		return gb
	}
	gb.model = &model
	return gb
}

// ModelParameters sets the model parameters
func (gb *GenerationBuilder) ModelParameters(params map[string]interface{}) *GenerationBuilder {
	if gb.submitted {
		return gb
	}
	gb.modelParameters = params
	return gb
}

// AddModelParameter adds a single model parameter
func (gb *GenerationBuilder) AddModelParameter(key string, value interface{}) *GenerationBuilder {
	if gb.submitted {
		return gb
	}
	if gb.modelParameters == nil {
		gb.modelParameters = make(map[string]interface{})
	}
	gb.modelParameters[key] = value
	return gb
}

// Temperature sets the temperature parameter
func (gb *GenerationBuilder) Temperature(temp float64) *GenerationBuilder {
	return gb.AddModelParameter("temperature", temp)
}

// MaxTokens sets the max_tokens parameter
func (gb *GenerationBuilder) MaxTokens(tokens int) *GenerationBuilder {
	return gb.AddModelParameter("max_tokens", tokens)
}

// TopP sets the top_p parameter
func (gb *GenerationBuilder) TopP(topP float64) *GenerationBuilder {
	return gb.AddModelParameter("top_p", topP)
}

// FrequencyPenalty sets the frequency_penalty parameter
func (gb *GenerationBuilder) FrequencyPenalty(penalty float64) *GenerationBuilder {
	return gb.AddModelParameter("frequency_penalty", penalty)
}

// PresencePenalty sets the presence_penalty parameter
func (gb *GenerationBuilder) PresencePenalty(penalty float64) *GenerationBuilder {
	return gb.AddModelParameter("presence_penalty", penalty)
}

// Input sets the input data
func (gb *GenerationBuilder) Input(input interface{}) *GenerationBuilder {
	if gb.submitted {
		return gb
	}
	gb.input = input
	return gb
}

// Output sets the output data
func (gb *GenerationBuilder) Output(output interface{}) *GenerationBuilder {
	if gb.submitted {
		return gb
	}
	gb.output = output
	return gb
}

// Usage sets the usage statistics
func (gb *GenerationBuilder) Usage(usage *types.Usage) *GenerationBuilder {
	if gb.submitted {
		return gb
	}
	gb.usage = usage
	return gb
}

// UsageTokens sets usage with token counts
func (gb *GenerationBuilder) UsageTokens(inputTokens, outputTokens int) *GenerationBuilder {
	if gb.submitted {
		return gb
	}
	gb.usage = types.NewUsage(inputTokens, outputTokens)
	return gb
}

// UsageWithCost sets usage with token counts and cost information
func (gb *GenerationBuilder) UsageWithCost(inputTokens, outputTokens int, inputCost, outputCost float64) *GenerationBuilder {
	if gb.submitted {
		return gb
	}
	gb.usage = types.NewUsageWithCost(inputTokens, outputTokens, inputCost, outputCost)
	return gb
}

// Metadata sets the metadata map
func (gb *GenerationBuilder) Metadata(metadata map[string]interface{}) *GenerationBuilder {
	if gb.submitted {
		return gb
	}
	gb.metadata = metadata
	return gb
}

// AddMetadata adds a single metadata key-value pair
func (gb *GenerationBuilder) AddMetadata(key string, value interface{}) *GenerationBuilder {
	if gb.submitted {
		return gb
	}
	if gb.metadata == nil {
		gb.metadata = make(map[string]interface{})
	}
	gb.metadata[key] = value
	return gb
}

// Level sets the observation level
func (gb *GenerationBuilder) Level(level types.ObservationLevel) *GenerationBuilder {
	if gb.submitted {
		return gb
	}
	gb.level = level
	return gb
}

// Debug sets the observation level to DEBUG
func (gb *GenerationBuilder) Debug() *GenerationBuilder {
	return gb.Level(types.ObservationLevelDebug)
}

// Warning sets the observation level to WARNING
func (gb *GenerationBuilder) Warning() *GenerationBuilder {
	return gb.Level(types.ObservationLevelWarning)
}

// Error sets the observation level to ERROR
func (gb *GenerationBuilder) Error() *GenerationBuilder {
	return gb.Level(types.ObservationLevelError)
}

// StatusMessage sets the status message
func (gb *GenerationBuilder) StatusMessage(message string) *GenerationBuilder {
	if gb.submitted {
		return gb
	}
	gb.statusMessage = &message
	return gb
}

// Version sets the version
func (gb *GenerationBuilder) Version(version string) *GenerationBuilder {
	if gb.submitted {
		return gb
	}
	gb.version = &version
	return gb
}

// GetID returns the generation ID
func (gb *GenerationBuilder) GetID() string {
	return gb.id
}

// GetTraceID returns the trace ID
func (gb *GenerationBuilder) GetTraceID() string {
	return gb.traceID
}

// GetName returns the generation name
func (gb *GenerationBuilder) GetName() string {
	return gb.name
}

// GetModel returns the model name
func (gb *GenerationBuilder) GetModel() *string {
	return gb.model
}

// GetUsage returns the usage statistics
func (gb *GenerationBuilder) GetUsage() *types.Usage {
	return gb.usage
}

// validate performs validation on the generation builder
func (gb *GenerationBuilder) validate() error {
	if gb.id == "" {
		return &ValidationError{Field: "id", Message: "generation id is required"}
	}
	
	if gb.traceID == "" {
		return &ValidationError{Field: "traceId", Message: "trace id is required"}
	}
	
	if gb.name == "" {
		return &ValidationError{Field: "name", Message: "generation name is required"}
	}
	
	if gb.startTime.IsZero() {
		return &ValidationError{Field: "startTime", Message: "start time is required"}
	}
	
	// Validate end time if set
	if gb.endTime != nil && gb.endTime.Before(gb.startTime) {
		return &ValidationError{Field: "endTime", Message: "end time cannot be before start time"}
	}
	
	// Validate completion start time if set
	if gb.completionStartTime != nil {
		if gb.completionStartTime.Before(gb.startTime) {
			return &ValidationError{Field: "completionStartTime", Message: "completion start time cannot be before start time"}
		}
		if gb.endTime != nil && gb.completionStartTime.After(*gb.endTime) {
			return &ValidationError{Field: "completionStartTime", Message: "completion start time cannot be after end time"}
		}
	}
	
	// Validate usage if present
	if gb.usage != nil {
		if gb.usage.Input != nil && *gb.usage.Input < 0 {
			return &ValidationError{Field: "usage.input", Message: "input token count cannot be negative"}
		}
		if gb.usage.Output != nil && *gb.usage.Output < 0 {
			return &ValidationError{Field: "usage.output", Message: "output token count cannot be negative"}
		}
		if gb.usage.Total != nil && *gb.usage.Total < 0 {
			return &ValidationError{Field: "usage.total", Message: "total token count cannot be negative"}
		}
	}
	
	return nil
}

// toObservationEvent converts the builder to an ObservationEvent
func (gb *GenerationBuilder) toObservationEvent() *ingestiontypes.ObservationEvent {
	return &ingestiontypes.ObservationEvent{
		ID:                   gb.id,
		TraceID:              gb.traceID,
		ParentObservationID:  gb.parentObservationID,
		Type:                 types.ObservationTypeGeneration,
		Name:                 gb.name,
		StartTime:            gb.startTime,
		EndTime:              gb.endTime,
		CompletionStartTime:  gb.completionStartTime,
		Model:                gb.model,
		ModelParameters:      gb.modelParameters,
		Input:                gb.input,
		Output:               gb.output,
		Usage:                gb.usage,
		Metadata:             gb.metadata,
		Level:                gb.level,
		StatusMessage:        gb.statusMessage,
		Version:              gb.version,
	}
}

// toGenerationCreateEvent converts the builder to a GenerationCreateEvent
func (gb *GenerationBuilder) toGenerationCreateEvent() *ingestiontypes.GenerationCreateEvent {
	return &ingestiontypes.GenerationCreateEvent{
		ObservationEvent: *gb.toObservationEvent(),
		EventType:        "generation-create",
	}
}

// toGenerationUpdateEvent converts the builder to a GenerationUpdateEvent
func (gb *GenerationBuilder) toGenerationUpdateEvent() *ingestiontypes.GenerationUpdateEvent {
	return &ingestiontypes.GenerationUpdateEvent{
		ObservationEvent: *gb.toObservationEvent(),
		EventType:        "generation-update",
	}
}

// Submit submits the generation to the ingestion queue
func (gb *GenerationBuilder) Submit(ctx context.Context) error {
	if gb.submitted {
		return &ValidationError{Field: "state", Message: "generation already submitted"}
	}
	
	if err := gb.validate(); err != nil {
		return err
	}
	
	event := gb.toGenerationCreateEvent()
	ingestionEvent := event.ToIngestionEvent()
	
	if err := gb.client.queue.Enqueue(ingestionEvent); err != nil {
		return err
	}
	
	gb.submitted = true
	return nil
}

// Update updates an existing generation
func (gb *GenerationBuilder) Update(ctx context.Context) error {
	if gb.submitted {
		return &ValidationError{Field: "state", Message: "generation already submitted"}
	}
	
	if err := gb.validate(); err != nil {
		return err
	}
	
	event := gb.toGenerationUpdateEvent()
	ingestionEvent := event.ToIngestionEvent()
	
	if err := gb.client.queue.Enqueue(ingestionEvent); err != nil {
		return err
	}
	
	gb.submitted = true
	return nil
}

// End ends the generation with the current timestamp and submits it
func (gb *GenerationBuilder) End(ctx context.Context) error {
	return gb.EndAt(ctx, time.Now().UTC())
}

// EndAt ends the generation with a specific timestamp and submits it
func (gb *GenerationBuilder) EndAt(ctx context.Context, endTime time.Time) error {
	gb.EndTime(endTime)
	return gb.Update(ctx)
}

// Stream starts streaming mode by setting completion start time
func (gb *GenerationBuilder) Stream() *GenerationBuilder {
	return gb.CompletionStartTime(time.Now().UTC())
}

// StreamAt starts streaming mode with a specific completion start time
func (gb *GenerationBuilder) StreamAt(completionStartTime time.Time) *GenerationBuilder {
	return gb.CompletionStartTime(completionStartTime)
}