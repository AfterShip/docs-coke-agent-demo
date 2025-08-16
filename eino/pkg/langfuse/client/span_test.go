package client

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"eino/pkg/langfuse/api/resources/commons/types"
	"eino/pkg/langfuse/internal/queue"
)

func TestSpanBuilder_FluentAPI(t *testing.T) {
	client := createTestClient(t)
	traceID := "test-trace-id"
	
	span := NewSpanBuilder(client, traceID).
		Name("test-span").
		Input("test input").
		Output("test output").
		WithStartTime(time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)).
		WithEndTime(time.Date(2024, 1, 1, 10, 5, 0, 0, time.UTC)).
		WithInput("updated input").
		WithOutput("updated output").
		WithMetadata(map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		}).
		WithLevel("WARNING").
		WithStatusMessage("Span completed with warning").
		Version("2.0.0")
	
	assert.Equal(t, "test-span", span.GetName())
	assert.Equal(t, traceID, span.GetTraceID())
	assert.Equal(t, "updated input", span.input)
	assert.Equal(t, "updated output", span.output)
	assert.Equal(t, types.ObservationLevelWarning, span.level)
	assert.Equal(t, "Span completed with warning", *span.statusMessage)
	assert.Equal(t, "2.0.0", *span.version)
	assert.Equal(t, "value1", span.metadata["key1"])
	assert.Equal(t, 42, span.metadata["key2"])
}

func TestSpanBuilder_BasicProperties(t *testing.T) {
	client := createTestClient(t)
	traceID := "test-trace-id"
	
	span := NewSpanBuilder(client, traceID)
	
	// Test auto-generated ID
	assert.NotEmpty(t, span.GetID())
	
	// Test trace ID assignment
	assert.Equal(t, traceID, span.GetTraceID())
	
	// Test default values
	assert.False(t, span.startTime.IsZero())
	assert.Equal(t, types.ObservationLevelDefault, span.level)
	assert.NotNil(t, span.metadata)
	assert.False(t, span.submitted)
}

func TestSpanBuilder_IDCustomization(t *testing.T) {
	client := createTestClient(t)
	
	// Test auto-generated ID
	span1 := NewSpanBuilder(client, "trace-id")
	assert.NotEmpty(t, span1.GetID())
	
	// Test different spans have different IDs
	span2 := NewSpanBuilder(client, "trace-id")
	assert.NotEqual(t, span1.GetID(), span2.GetID())
	
	// Test custom ID
	customID := "custom-span-id"
	span3 := NewSpanBuilder(client, "trace-id").ID(customID)
	assert.Equal(t, customID, span3.GetID())
}

func TestSpanBuilder_ParentObservationID(t *testing.T) {
	client := createTestClient(t)
	
	parentID := "parent-observation-id"
	span := NewSpanBuilder(client, "trace-id").
		ParentObservationID(parentID)
	
	assert.Equal(t, parentID, *span.parentObservationID)
}

func TestSpanBuilder_TimingHandling(t *testing.T) {
	client := createTestClient(t)
	span := NewSpanBuilder(client, "trace-id")
	
	// Test auto-generated start time
	assert.False(t, span.startTime.IsZero())
	
	// Test custom start time
	startTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	span.StartTime(startTime)
	assert.Equal(t, startTime, span.startTime)
	
	// Test end time
	endTime := time.Date(2024, 1, 1, 10, 5, 0, 0, time.UTC)
	span.EndTime(endTime)
	assert.Equal(t, endTime, *span.endTime)
	
	// Test timezone conversion to UTC
	localTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.Local)
	span.StartTime(localTime)
	assert.Equal(t, time.UTC, span.startTime.Location())
	
	// Test fluent API aliases
	span2 := NewSpanBuilder(client, "trace-id")
	span2.WithStartTime(startTime).WithEndTime(endTime)
	assert.Equal(t, startTime, span2.startTime)
	assert.Equal(t, endTime, *span2.endTime)
}

func TestSpanBuilder_MetadataHandling(t *testing.T) {
	client := createTestClient(t)
	span := NewSpanBuilder(client, "trace-id")
	
	// Test adding individual metadata
	span.AddMetadata("key1", "value1")
	span.AddMetadata("key2", 42)
	span.AddMetadata("key3", true)
	
	assert.Equal(t, "value1", span.metadata["key1"])
	assert.Equal(t, 42, span.metadata["key2"])
	assert.Equal(t, true, span.metadata["key3"])
	
	// Test replacing metadata map
	newMetadata := map[string]interface{}{
		"newKey": "newValue",
		"count":  100,
	}
	span.Metadata(newMetadata)
	assert.Equal(t, newMetadata, span.metadata)
	assert.NotContains(t, span.metadata, "key1")
	
	// Test fluent API alias
	span2 := NewSpanBuilder(client, "trace-id")
	span2.WithMetadata(newMetadata)
	assert.Equal(t, newMetadata, span2.metadata)
}

func TestSpanBuilder_LevelHandling(t *testing.T) {
	client := createTestClient(t)
	
	tests := []struct {
		name     string
		setup    func(*SpanBuilder) *SpanBuilder
		expected types.ObservationLevel
	}{
		{
			name:     "default level",
			setup:    func(sb *SpanBuilder) *SpanBuilder { return sb },
			expected: types.ObservationLevelDefault,
		},
		{
			name:     "debug level method",
			setup:    func(sb *SpanBuilder) *SpanBuilder { return sb.Debug() },
			expected: types.ObservationLevelDebug,
		},
		{
			name:     "warning level method",
			setup:    func(sb *SpanBuilder) *SpanBuilder { return sb.Warning() },
			expected: types.ObservationLevelWarning,
		},
		{
			name:     "error level method",
			setup:    func(sb *SpanBuilder) *SpanBuilder { return sb.Error() },
			expected: types.ObservationLevelError,
		},
		{
			name:     "explicit level",
			setup:    func(sb *SpanBuilder) *SpanBuilder { return sb.Level(types.ObservationLevelDebug) },
			expected: types.ObservationLevelDebug,
		},
		{
			name:     "WithLevel DEBUG string",
			setup:    func(sb *SpanBuilder) *SpanBuilder { return sb.WithLevel("DEBUG") },
			expected: types.ObservationLevelDebug,
		},
		{
			name:     "WithLevel WARNING string",
			setup:    func(sb *SpanBuilder) *SpanBuilder { return sb.WithLevel("WARNING") },
			expected: types.ObservationLevelWarning,
		},
		{
			name:     "WithLevel ERROR string",
			setup:    func(sb *SpanBuilder) *SpanBuilder { return sb.WithLevel("ERROR") },
			expected: types.ObservationLevelError,
		},
		{
			name:     "WithLevel unknown string defaults",
			setup:    func(sb *SpanBuilder) *SpanBuilder { return sb.WithLevel("UNKNOWN") },
			expected: types.ObservationLevelDefault,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			span := NewSpanBuilder(client, "trace-id")
			span = tt.setup(span)
			assert.Equal(t, tt.expected, span.level)
		})
	}
}

func TestSpanBuilder_InputOutputHandling(t *testing.T) {
	client := createTestClient(t)
	span := NewSpanBuilder(client, "trace-id")
	
	// Test different input/output types
	inputs := []interface{}{
		"string input",
		42,
		true,
		map[string]interface{}{"key": "value"},
		[]string{"item1", "item2"},
	}
	
	outputs := []interface{}{
		"string output",
		100,
		false,
		map[string]interface{}{"result": "success"},
		[]int{1, 2, 3},
	}
	
	for i, input := range inputs {
		output := outputs[i]
		span.Input(input).Output(output)
		assert.Equal(t, input, span.input)
		assert.Equal(t, output, span.output)
	}
	
	// Test fluent API aliases
	span2 := NewSpanBuilder(client, "trace-id")
	span2.WithInput("fluent input").WithOutput("fluent output")
	assert.Equal(t, "fluent input", span2.input)
	assert.Equal(t, "fluent output", span2.output)
}

func TestSpanBuilder_StatusMessage(t *testing.T) {
	client := createTestClient(t)
	span := NewSpanBuilder(client, "trace-id")
	
	message := "Operation completed successfully"
	span.StatusMessage(message)
	assert.Equal(t, message, *span.statusMessage)
	
	// Test fluent API alias
	span2 := NewSpanBuilder(client, "trace-id")
	span2.WithStatusMessage(message)
	assert.Equal(t, message, *span2.statusMessage)
}

func TestSpanBuilder_ChildSpan(t *testing.T) {
	client := createTestClient(t)
	parentSpan := NewSpanBuilder(client, "trace-id").Name("parent-span")
	
	childSpan := parentSpan.ChildSpan("child-span")
	
	assert.Equal(t, "child-span", childSpan.GetName())
	assert.Equal(t, parentSpan.GetTraceID(), childSpan.GetTraceID())
	assert.Equal(t, parentSpan.GetID(), *childSpan.parentObservationID)
	assert.NotEqual(t, parentSpan.GetID(), childSpan.GetID())
}

func TestSpanBuilder_Validation(t *testing.T) {
	client := createTestClient(t)
	
	tests := []struct {
		name        string
		setupSpan   func() *SpanBuilder
		expectError bool
		errorField  string
	}{
		{
			name: "valid span",
			setupSpan: func() *SpanBuilder {
				return NewSpanBuilder(client, "trace-id").Name("test-span")
			},
			expectError: false,
		},
		{
			name: "missing trace ID",
			setupSpan: func() *SpanBuilder {
				span := NewSpanBuilder(client, "")
				span.Name("test-span")
				return span
			},
			expectError: true,
			errorField:  "traceId",
		},
		{
			name: "missing name",
			setupSpan: func() *SpanBuilder {
				return NewSpanBuilder(client, "trace-id").Name("")
			},
			expectError: true,
			errorField:  "name",
		},
		{
			name: "missing ID",
			setupSpan: func() *SpanBuilder {
				span := NewSpanBuilder(client, "trace-id").Name("test-span")
				span.id = ""
				return span
			},
			expectError: true,
			errorField:  "id",
		},
		{
			name: "zero start time",
			setupSpan: func() *SpanBuilder {
				span := NewSpanBuilder(client, "trace-id").Name("test-span")
				span.startTime = time.Time{}
				return span
			},
			expectError: true,
			errorField:  "startTime",
		},
		{
			name: "end time before start time",
			setupSpan: func() *SpanBuilder {
				span := NewSpanBuilder(client, "trace-id").Name("test-span")
				span.StartTime(time.Now())
				span.EndTime(time.Now().Add(-time.Hour))
				return span
			},
			expectError: true,
			errorField:  "endTime",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			span := tt.setupSpan()
			err := span.validate()
			
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorField != "" {
					validationErr, ok := err.(*ValidationError)
					require.True(t, ok, "expected ValidationError")
					assert.Equal(t, tt.errorField, validationErr.Field)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSpanBuilder_Submit(t *testing.T) {
	client := createTestClient(t)
	
	span := NewSpanBuilder(client, "trace-id").
		Name("test-span").
		Input("test input").
		Output("test output")
	
	err := span.Submit(context.Background())
	assert.NoError(t, err)
	assert.True(t, span.submitted)
	
	// Test submitting again should fail
	err = span.Submit(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already submitted")
}

func TestSpanBuilder_Update(t *testing.T) {
	client := createTestClient(t)
	
	span := NewSpanBuilder(client, "trace-id").
		Name("test-span").
		Input("initial input")
	
	// Update with different data
	span.Output("final output")
	
	err := span.Update(context.Background())
	assert.NoError(t, err)
	assert.True(t, span.submitted)
}

func TestSpanBuilder_End(t *testing.T) {
	client := createTestClient(t)
	
	span := NewSpanBuilder(client, "trace-id").
		Name("test-span")
	
	err := span.End(context.Background())
	assert.NoError(t, err)
	assert.True(t, span.submitted)
	assert.NotNil(t, span.endTime)
}

func TestSpanBuilder_EndAt(t *testing.T) {
	client := createTestClient(t)
	
	span := NewSpanBuilder(client, "trace-id").
		Name("test-span")
	
	endTime := time.Now().UTC().Add(time.Hour)
	err := span.EndAt(context.Background(), endTime)
	assert.NoError(t, err)
	assert.True(t, span.submitted)
	assert.Equal(t, endTime, *span.endTime)
}

func TestSpanBuilder_ImmutabilityAfterSubmit(t *testing.T) {
	client := createTestClient(t)
	
	span := NewSpanBuilder(client, "trace-id").
		Name("test-span")
	
	originalID := span.GetID()
	originalName := span.GetName()
	
	// Submit the span
	err := span.Submit(context.Background())
	require.NoError(t, err)
	
	// Try to modify after submission - should not change anything
	span.ID("new-id")
	span.Name("new-name")
	span.Input("new-input")
	span.AddMetadata("key", "value")
	
	assert.Equal(t, originalID, span.GetID())
	assert.Equal(t, originalName, span.GetName())
	assert.Nil(t, span.input)
	assert.Empty(t, span.metadata)
}

func TestSpanBuilder_EventConversion(t *testing.T) {
	client := createTestClient(t)
	
	span := NewSpanBuilder(client, "trace-id").
		Name("test-span").
		Input("test input").
		Output("test output").
		AddMetadata("key", "value")
	
	// Test observation event conversion
	obsEvent := span.toObservationEvent()
	assert.Equal(t, span.id, obsEvent.ID)
	assert.Equal(t, span.traceID, obsEvent.TraceID)
	assert.Equal(t, types.ObservationTypeSpan, obsEvent.Type)
	assert.Equal(t, span.name, obsEvent.Name)
	assert.Equal(t, span.input, obsEvent.Input)
	assert.Equal(t, span.output, obsEvent.Output)
	assert.Equal(t, span.metadata, obsEvent.Metadata)
	assert.Equal(t, span.level, obsEvent.Level)
	
	// Test span create event conversion
	createEvent := span.toSpanCreateEvent()
	assert.Equal(t, "span-create", createEvent.EventType)
	assert.Equal(t, span.id, createEvent.ObservationEvent.ID)
	
	// Test span update event conversion
	updateEvent := span.toSpanUpdateEvent()
	assert.Equal(t, "span-update", updateEvent.EventType)
	assert.Equal(t, span.id, updateEvent.ObservationEvent.ID)
}

func TestSpanBuilder_ConcurrentAccess(t *testing.T) {
	client := createTestClient(t)
	span := NewSpanBuilder(client, "trace-id")
	
	// Run concurrent operations
	done := make(chan bool, 10)
	
	// Start 10 goroutines modifying the span
	for i := 0; i < 10; i++ {
		go func(i int) {
			defer func() { done <- true }()
			
			span.AddMetadata(string(rune('a'+i)), i)
		}(i)
	}
	
	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// The span should have all metadata
	assert.Len(t, span.metadata, 10)
}

func TestSpanBuilder_ChainedOperations(t *testing.T) {
	client := createTestClient(t)
	
	// Test complex chained operations
	span := NewSpanBuilder(client, "trace-id").
		Name("complex-span").
		Input(map[string]interface{}{
			"operation": "data processing",
			"parameters": map[string]interface{}{
				"batch_size": 100,
				"timeout":    30,
			},
		}).
		AddMetadata("environment", "production").
		AddMetadata("service", "data-processor").
		AddMetadata("version", "2.1.0").
		Warning().
		StatusMessage("Processing with elevated memory usage")
	
	// Verify all chained operations worked
	assert.Equal(t, "complex-span", span.GetName())
	assert.NotNil(t, span.input)
	assert.Equal(t, "production", span.metadata["environment"])
	assert.Equal(t, "data-processor", span.metadata["service"])
	assert.Equal(t, "2.1.0", span.metadata["version"])
	assert.Equal(t, types.ObservationLevelWarning, span.level)
	assert.Equal(t, "Processing with elevated memory usage", *span.statusMessage)
}

func TestSpanBuilder_NestedSpans(t *testing.T) {
	client := createTestClient(t)
	
	// Create parent span
	parent := NewSpanBuilder(client, "trace-id").
		Name("parent-operation")
	
	// Create child span
	child1 := parent.ChildSpan("child-operation-1")
	child2 := parent.ChildSpan("child-operation-2")
	
	// Create grandchild span
	grandchild := child1.ChildSpan("grandchild-operation")
	
	// Verify hierarchy
	assert.Equal(t, parent.GetID(), *child1.parentObservationID)
	assert.Equal(t, parent.GetID(), *child2.parentObservationID)
	assert.Equal(t, child1.GetID(), *grandchild.parentObservationID)
	
	// All spans should share the same trace ID
	assert.Equal(t, parent.GetTraceID(), child1.GetTraceID())
	assert.Equal(t, parent.GetTraceID(), child2.GetTraceID())
	assert.Equal(t, parent.GetTraceID(), grandchild.GetTraceID())
	
	// All spans should have unique IDs
	ids := []string{parent.GetID(), child1.GetID(), child2.GetID(), grandchild.GetID()}
	uniqueIds := make(map[string]bool)
	for _, id := range ids {
		uniqueIds[id] = true
	}
	assert.Len(t, uniqueIds, 4, "All span IDs should be unique")
}