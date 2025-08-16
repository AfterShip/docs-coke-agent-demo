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

func TestGenerationBuilder_FluentAPI(t *testing.T) {
	client := createTestClient(t)
	traceID := "test-trace-id"
	
	generation := NewGenerationBuilder(client, traceID).
		Name("test-generation").
		Model("gpt-4").
		Temperature(0.7).
		MaxTokens(1000).
		TopP(0.9).
		FrequencyPenalty(0.1).
		PresencePenalty(0.2).
		Input("Hello, world!").
		Output("Hi there!").
		UsageTokens(10, 20).
		AddMetadata("key", "value").
		Debug().
		StatusMessage("Generation completed").
		Version("1.0.0")
	
	assert.Equal(t, "test-generation", generation.GetName())
	assert.Equal(t, traceID, generation.GetTraceID())
	assert.Equal(t, "gpt-4", *generation.GetModel())
	assert.Equal(t, 0.7, generation.modelParameters["temperature"])
	assert.Equal(t, 1000, generation.modelParameters["max_tokens"])
	assert.Equal(t, 0.9, generation.modelParameters["top_p"])
	assert.Equal(t, 0.1, generation.modelParameters["frequency_penalty"])
	assert.Equal(t, 0.2, generation.modelParameters["presence_penalty"])
	assert.Equal(t, "Hello, world!", generation.input)
	assert.Equal(t, "Hi there!", generation.output)
	assert.Equal(t, types.ObservationLevelDebug, generation.level)
	assert.Equal(t, "Generation completed", *generation.statusMessage)
	assert.Equal(t, "1.0.0", *generation.version)
	assert.Equal(t, "value", generation.metadata["key"])
	
	// Test usage
	usage := generation.GetUsage()
	require.NotNil(t, usage)
	assert.Equal(t, 10, *usage.Input)
	assert.Equal(t, 20, *usage.Output)
	assert.Equal(t, 30, *usage.Total)
}

func TestGenerationBuilder_ModelParameters(t *testing.T) {
	client := createTestClient(t)
	generation := NewGenerationBuilder(client, "trace-id")
	
	// Test setting model parameters individually
	generation.
		AddModelParameter("custom_param", "custom_value").
		AddModelParameter("numeric_param", 42).
		AddModelParameter("bool_param", true)
	
	assert.Equal(t, "custom_value", generation.modelParameters["custom_param"])
	assert.Equal(t, 42, generation.modelParameters["numeric_param"])
	assert.Equal(t, true, generation.modelParameters["bool_param"])
	
	// Test setting model parameters as map
	params := map[string]interface{}{
		"temperature": 0.8,
		"top_k": 50,
	}
	generation.ModelParameters(params)
	assert.Equal(t, params, generation.modelParameters)
	assert.NotContains(t, generation.modelParameters, "custom_param") // Should be replaced
}

func TestGenerationBuilder_UsageVariations(t *testing.T) {
	client := createTestClient(t)
	
	t.Run("UsageTokens", func(t *testing.T) {
		generation := NewGenerationBuilder(client, "trace-id")
		generation.UsageTokens(100, 200)
		
		usage := generation.GetUsage()
		require.NotNil(t, usage)
		assert.Equal(t, 100, *usage.Input)
		assert.Equal(t, 200, *usage.Output)
		assert.Equal(t, 300, *usage.Total)
	})
	
	t.Run("UsageWithCost", func(t *testing.T) {
		generation := NewGenerationBuilder(client, "trace-id")
		generation.UsageWithCost(100, 200, 0.01, 0.02)
		
		usage := generation.GetUsage()
		require.NotNil(t, usage)
		assert.Equal(t, 100, *usage.Input)
		assert.Equal(t, 200, *usage.Output)
		assert.Equal(t, 300, *usage.Total)
		assert.Equal(t, 0.01, *usage.InputCost)
		assert.Equal(t, 0.02, *usage.OutputCost)
		assert.Equal(t, 0.03, *usage.TotalCost)
	})
	
	t.Run("CustomUsage", func(t *testing.T) {
		generation := NewGenerationBuilder(client, "trace-id")
		customUsage := &types.Usage{
			Input:      intPtr(50),
			Output:     intPtr(75),
			Total:      intPtr(125),
			InputCost:  floatPtr(0.005),
			OutputCost: floatPtr(0.015),
			TotalCost:  floatPtr(0.02),
			Unit:       stringPtr("tokens"),
		}
		generation.Usage(customUsage)
		
		usage := generation.GetUsage()
		assert.Equal(t, customUsage, usage)
	})
}

func TestGenerationBuilder_TimingHandling(t *testing.T) {
	client := createTestClient(t)
	generation := NewGenerationBuilder(client, "trace-id")
	
	// Test auto-generated start time
	assert.False(t, generation.startTime.IsZero())
	
	// Test custom start time
	startTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	generation.StartTime(startTime)
	assert.Equal(t, startTime, generation.startTime)
	
	// Test end time
	endTime := time.Date(2024, 1, 1, 10, 5, 0, 0, time.UTC)
	generation.EndTime(endTime)
	assert.Equal(t, endTime, *generation.endTime)
	
	// Test completion start time
	completionStartTime := time.Date(2024, 1, 1, 10, 2, 0, 0, time.UTC)
	generation.CompletionStartTime(completionStartTime)
	assert.Equal(t, completionStartTime, *generation.completionStartTime)
	
	// Test timezone conversion to UTC
	localTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.Local)
	generation.StartTime(localTime)
	assert.Equal(t, time.UTC, generation.startTime.Location())
}

func TestGenerationBuilder_LevelHandling(t *testing.T) {
	client := createTestClient(t)
	
	tests := []struct {
		name     string
		setup    func(*GenerationBuilder) *GenerationBuilder
		expected types.ObservationLevel
	}{
		{
			name:     "default level",
			setup:    func(gb *GenerationBuilder) *GenerationBuilder { return gb },
			expected: types.ObservationLevelDefault,
		},
		{
			name:     "debug level",
			setup:    func(gb *GenerationBuilder) *GenerationBuilder { return gb.Debug() },
			expected: types.ObservationLevelDebug,
		},
		{
			name:     "warning level",
			setup:    func(gb *GenerationBuilder) *GenerationBuilder { return gb.Warning() },
			expected: types.ObservationLevelWarning,
		},
		{
			name:     "error level",
			setup:    func(gb *GenerationBuilder) *GenerationBuilder { return gb.Error() },
			expected: types.ObservationLevelError,
		},
		{
			name:     "explicit level",
			setup:    func(gb *GenerationBuilder) *GenerationBuilder { return gb.Level(types.ObservationLevelWarning) },
			expected: types.ObservationLevelWarning,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generation := NewGenerationBuilder(client, "trace-id")
			generation = tt.setup(generation)
			assert.Equal(t, tt.expected, generation.level)
		})
	}
}

func TestGenerationBuilder_StreamingSupport(t *testing.T) {
	client := createTestClient(t)
	generation := NewGenerationBuilder(client, "trace-id")
	
	// Test Stream() method
	generation.Stream()
	assert.NotNil(t, generation.completionStartTime)
	assert.False(t, generation.completionStartTime.IsZero())
	
	// Test StreamAt() method
	customTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	generation2 := NewGenerationBuilder(client, "trace-id")
	generation2.StreamAt(customTime)
	assert.Equal(t, customTime, *generation2.completionStartTime)
}

func TestGenerationBuilder_Validation(t *testing.T) {
	client := createTestClient(t)
	
	tests := []struct {
		name        string
		setupGen    func() *GenerationBuilder
		expectError bool
		errorField  string
	}{
		{
			name: "valid generation",
			setupGen: func() *GenerationBuilder {
				return NewGenerationBuilder(client, "trace-id").Name("test-gen")
			},
			expectError: false,
		},
		{
			name: "missing trace ID",
			setupGen: func() *GenerationBuilder {
				gen := NewGenerationBuilder(client, "")
				gen.Name("test-gen")
				return gen
			},
			expectError: true,
			errorField:  "traceId",
		},
		{
			name: "missing name",
			setupGen: func() *GenerationBuilder {
				return NewGenerationBuilder(client, "trace-id").Name("")
			},
			expectError: true,
			errorField:  "name",
		},
		{
			name: "end time before start time",
			setupGen: func() *GenerationBuilder {
				gen := NewGenerationBuilder(client, "trace-id").Name("test-gen")
				gen.StartTime(time.Now())
				gen.EndTime(time.Now().Add(-time.Hour))
				return gen
			},
			expectError: true,
			errorField:  "endTime",
		},
		{
			name: "completion start time before start time",
			setupGen: func() *GenerationBuilder {
				gen := NewGenerationBuilder(client, "trace-id").Name("test-gen")
				gen.StartTime(time.Now())
				gen.CompletionStartTime(time.Now().Add(-time.Hour))
				return gen
			},
			expectError: true,
			errorField:  "completionStartTime",
		},
		{
			name: "completion start time after end time",
			setupGen: func() *GenerationBuilder {
				gen := NewGenerationBuilder(client, "trace-id").Name("test-gen")
				now := time.Now()
				gen.StartTime(now)
				gen.EndTime(now.Add(time.Hour))
				gen.CompletionStartTime(now.Add(2 * time.Hour))
				return gen
			},
			expectError: true,
			errorField:  "completionStartTime",
		},
		{
			name: "negative input tokens",
			setupGen: func() *GenerationBuilder {
				gen := NewGenerationBuilder(client, "trace-id").Name("test-gen")
				gen.usage = &types.Usage{Input: intPtr(-1)}
				return gen
			},
			expectError: true,
			errorField:  "usage.input",
		},
		{
			name: "negative output tokens",
			setupGen: func() *GenerationBuilder {
				gen := NewGenerationBuilder(client, "trace-id").Name("test-gen")
				gen.usage = &types.Usage{Output: intPtr(-1)}
				return gen
			},
			expectError: true,
			errorField:  "usage.output",
		},
		{
			name: "negative total tokens",
			setupGen: func() *GenerationBuilder {
				gen := NewGenerationBuilder(client, "trace-id").Name("test-gen")
				gen.usage = &types.Usage{Total: intPtr(-1)}
				return gen
			},
			expectError: true,
			errorField:  "usage.total",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := tt.setupGen()
			err := gen.validate()
			
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

func TestGenerationBuilder_Submit(t *testing.T) {
	client := createTestClient(t)
	
	generation := NewGenerationBuilder(client, "trace-id").
		Name("test-generation").
		Model("gpt-4").
		Input("test input").
		Output("test output")
	
	err := generation.Submit(context.Background())
	assert.NoError(t, err)
	assert.True(t, generation.submitted)
	
	// Test submitting again should fail
	err = generation.Submit(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already submitted")
}

func TestGenerationBuilder_Update(t *testing.T) {
	client := createTestClient(t)
	
	generation := NewGenerationBuilder(client, "trace-id").
		Name("test-generation").
		Input("initial input")
	
	// Update with different data
	generation.Output("final output")
	
	err := generation.Update(context.Background())
	assert.NoError(t, err)
	assert.True(t, generation.submitted)
}

func TestGenerationBuilder_End(t *testing.T) {
	client := createTestClient(t)
	
	generation := NewGenerationBuilder(client, "trace-id").
		Name("test-generation")
	
	err := generation.End(context.Background())
	assert.NoError(t, err)
	assert.True(t, generation.submitted)
	assert.NotNil(t, generation.endTime)
}

func TestGenerationBuilder_EndAt(t *testing.T) {
	client := createTestClient(t)
	
	generation := NewGenerationBuilder(client, "trace-id").
		Name("test-generation")
	
	endTime := time.Now().UTC().Add(time.Hour)
	err := generation.EndAt(context.Background(), endTime)
	assert.NoError(t, err)
	assert.True(t, generation.submitted)
	assert.Equal(t, endTime, *generation.endTime)
}

func TestGenerationBuilder_ParentObservationID(t *testing.T) {
	client := createTestClient(t)
	
	parentID := "parent-observation-id"
	generation := NewGenerationBuilder(client, "trace-id").
		ParentObservationID(parentID)
	
	assert.Equal(t, parentID, *generation.parentObservationID)
}

func TestGenerationBuilder_ImmutabilityAfterSubmit(t *testing.T) {
	client := createTestClient(t)
	
	generation := NewGenerationBuilder(client, "trace-id").
		Name("test-generation")
	
	originalID := generation.GetID()
	originalName := generation.GetName()
	
	// Submit the generation
	err := generation.Submit(context.Background())
	require.NoError(t, err)
	
	// Try to modify after submission - should not change anything
	generation.ID("new-id")
	generation.Name("new-name")
	generation.Model("new-model")
	generation.AddModelParameter("temp", 0.5)
	generation.AddMetadata("key", "value")
	
	assert.Equal(t, originalID, generation.GetID())
	assert.Equal(t, originalName, generation.GetName())
	assert.Nil(t, generation.GetModel())
	assert.Empty(t, generation.modelParameters)
	assert.Empty(t, generation.metadata)
}

func TestGenerationBuilder_EventConversion(t *testing.T) {
	client := createTestClient(t)
	
	generation := NewGenerationBuilder(client, "trace-id").
		Name("test-generation").
		Model("gpt-4").
		Input("test input").
		Output("test output").
		UsageTokens(10, 20)
	
	// Test observation event conversion
	obsEvent := generation.toObservationEvent()
	assert.Equal(t, generation.id, obsEvent.ID)
	assert.Equal(t, generation.traceID, obsEvent.TraceID)
	assert.Equal(t, types.ObservationTypeGeneration, obsEvent.Type)
	assert.Equal(t, generation.name, obsEvent.Name)
	assert.Equal(t, generation.model, obsEvent.Model)
	assert.Equal(t, generation.input, obsEvent.Input)
	assert.Equal(t, generation.output, obsEvent.Output)
	assert.Equal(t, generation.usage, obsEvent.Usage)
	
	// Test generation create event conversion
	createEvent := generation.toGenerationCreateEvent()
	assert.Equal(t, "generation-create", createEvent.EventType)
	assert.Equal(t, generation.id, createEvent.ObservationEvent.ID)
	
	// Test generation update event conversion
	updateEvent := generation.toGenerationUpdateEvent()
	assert.Equal(t, "generation-update", updateEvent.EventType)
	assert.Equal(t, generation.id, updateEvent.ObservationEvent.ID)
}

func TestGenerationBuilder_ConcurrentAccess(t *testing.T) {
	client := createTestClient(t)
	generation := NewGenerationBuilder(client, "trace-id")
	
	// Run concurrent operations
	done := make(chan bool, 10)
	
	// Start 10 goroutines modifying the generation
	for i := 0; i < 10; i++ {
		go func(i int) {
			defer func() { done <- true }()
			
			generation.AddMetadata(string(rune('a'+i)), i)
			generation.AddModelParameter(string(rune('A'+i)), i)
		}(i)
	}
	
	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// The generation should have all metadata and model parameters
	assert.Len(t, generation.metadata, 10)
	assert.Len(t, generation.modelParameters, 10)
}

// Helper functions for tests
func intPtr(i int) *int {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}

func stringPtr(s string) *string {
	return &s
}