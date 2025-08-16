package client

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"eino/pkg/langfuse/api/resources/commons/types"
)

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{
		Field:   "testField",
		Message: "test error message",
	}
	
	expected := "testField: test error message"
	assert.Equal(t, expected, err.Error())
}

func TestTraceBuilder_ComprehensiveValidation(t *testing.T) {
	client := createTestClient(t)
	
	tests := []struct {
		name          string
		setupTrace    func() *TraceBuilder
		expectError   bool
		expectedField string
		expectedMsg   string
	}{
		{
			name: "valid minimal trace",
			setupTrace: func() *TraceBuilder {
				return client.Trace("minimal-trace")
			},
			expectError: false,
		},
		{
			name: "valid full trace",
			setupTrace: func() *TraceBuilder {
				return client.Trace("full-trace").
					UserID("user123").
					SessionID("session456").
					Input("test input").
					Output("test output").
					AddMetadata("env", "test").
					AddTag("integration").
					Version("1.0.0").
					Release("v1.0").
					Public(true)
			},
			expectError: false,
		},
		{
			name: "empty trace name",
			setupTrace: func() *TraceBuilder {
				return client.Trace("")
			},
			expectError:   true,
			expectedField: "name",
			expectedMsg:   "trace name is required",
		},
		{
			name: "whitespace-only trace name",
			setupTrace: func() *TraceBuilder {
				return client.Trace("   ")
			},
			expectError: false, // Whitespace names are allowed
		},
		{
			name: "missing ID",
			setupTrace: func() *TraceBuilder {
				trace := client.Trace("test")
				trace.id = ""
				return trace
			},
			expectError:   true,
			expectedField: "id",
			expectedMsg:   "trace id is required",
		},
		{
			name: "zero timestamp",
			setupTrace: func() *TraceBuilder {
				trace := client.Trace("test")
				trace.timestamp = time.Time{}
				return trace
			},
			expectError:   true,
			expectedField: "timestamp",
			expectedMsg:   "trace timestamp is required",
		},
		{
			name: "very long trace name",
			setupTrace: func() *TraceBuilder {
				longName := strings.Repeat("a", 10000)
				return client.Trace(longName)
			},
			expectError: false, // Long names are allowed
		},
		{
			name: "special characters in name",
			setupTrace: func() *TraceBuilder {
				return client.Trace("trace-with-special@chars!#$%^&*()")
			},
			expectError: false,
		},
		{
			name: "unicode characters in name",
			setupTrace: func() *TraceBuilder {
				return client.Trace("测试跟踪-🚀-trace")
			},
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trace := tt.setupTrace()
			err := trace.validate()
			
			if tt.expectError {
				require.Error(t, err)
				validationErr, ok := err.(*ValidationError)
				require.True(t, ok, "expected ValidationError")
				
				if tt.expectedField != "" {
					assert.Equal(t, tt.expectedField, validationErr.Field)
				}
				if tt.expectedMsg != "" {
					assert.Equal(t, tt.expectedMsg, validationErr.Message)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGenerationBuilder_ComprehensiveValidation(t *testing.T) {
	client := createTestClient(t)
	
	tests := []struct {
		name          string
		setupGen      func() *GenerationBuilder
		expectError   bool
		expectedField string
		expectedMsg   string
	}{
		{
			name: "valid minimal generation",
			setupGen: func() *GenerationBuilder {
				return NewGenerationBuilder(client, "trace-id").Name("test-gen")
			},
			expectError: false,
		},
		{
			name: "valid full generation",
			setupGen: func() *GenerationBuilder {
				return NewGenerationBuilder(client, "trace-id").
					Name("full-gen").
					Model("gpt-4").
					Temperature(0.7).
					Input("test").
					Output("result").
					UsageTokens(10, 20).
					AddMetadata("key", "value")
			},
			expectError: false,
		},
		{
			name: "empty trace ID",
			setupGen: func() *GenerationBuilder {
				gen := NewGenerationBuilder(client, "")
				return gen.Name("test-gen")
			},
			expectError:   true,
			expectedField: "traceId",
			expectedMsg:   "trace id is required",
		},
		{
			name: "empty generation name",
			setupGen: func() *GenerationBuilder {
				return NewGenerationBuilder(client, "trace-id").Name("")
			},
			expectError:   true,
			expectedField: "name",
			expectedMsg:   "generation name is required",
		},
		{
			name: "empty generation ID",
			setupGen: func() *GenerationBuilder {
				gen := NewGenerationBuilder(client, "trace-id").Name("test")
				gen.id = ""
				return gen
			},
			expectError:   true,
			expectedField: "id",
			expectedMsg:   "generation id is required",
		},
		{
			name: "zero start time",
			setupGen: func() *GenerationBuilder {
				gen := NewGenerationBuilder(client, "trace-id").Name("test")
				gen.startTime = time.Time{}
				return gen
			},
			expectError:   true,
			expectedField: "startTime",
			expectedMsg:   "start time is required",
		},
		{
			name: "end time before start time",
			setupGen: func() *GenerationBuilder {
				gen := NewGenerationBuilder(client, "trace-id").Name("test")
				now := time.Now()
				gen.startTime = now
				gen.endTime = &[]time.Time{now.Add(-time.Hour)}[0]
				return gen
			},
			expectError:   true,
			expectedField: "endTime",
			expectedMsg:   "end time cannot be before start time",
		},
		{
			name: "completion start time before start time",
			setupGen: func() *GenerationBuilder {
				gen := NewGenerationBuilder(client, "trace-id").Name("test")
				now := time.Now()
				gen.startTime = now
				gen.completionStartTime = &[]time.Time{now.Add(-time.Hour)}[0]
				return gen
			},
			expectError:   true,
			expectedField: "completionStartTime",
			expectedMsg:   "completion start time cannot be before start time",
		},
		{
			name: "completion start time after end time",
			setupGen: func() *GenerationBuilder {
				gen := NewGenerationBuilder(client, "trace-id").Name("test")
				now := time.Now()
				gen.startTime = now
				gen.endTime = &[]time.Time{now.Add(time.Hour)}[0]
				gen.completionStartTime = &[]time.Time{now.Add(2 * time.Hour)}[0]
				return gen
			},
			expectError:   true,
			expectedField: "completionStartTime",
			expectedMsg:   "completion start time cannot be after end time",
		},
		{
			name: "valid completion timing",
			setupGen: func() *GenerationBuilder {
				gen := NewGenerationBuilder(client, "trace-id").Name("test")
				now := time.Now()
				gen.startTime = now
				gen.completionStartTime = &[]time.Time{now.Add(time.Minute)}[0]
				gen.endTime = &[]time.Time{now.Add(2 * time.Minute)}[0]
				return gen
			},
			expectError: false,
		},
		{
			name: "negative input tokens",
			setupGen: func() *GenerationBuilder {
				gen := NewGenerationBuilder(client, "trace-id").Name("test")
				gen.usage = &types.Usage{Input: intPtr(-1)}
				return gen
			},
			expectError:   true,
			expectedField: "usage.input",
			expectedMsg:   "input token count cannot be negative",
		},
		{
			name: "negative output tokens",
			setupGen: func() *GenerationBuilder {
				gen := NewGenerationBuilder(client, "trace-id").Name("test")
				gen.usage = &types.Usage{Output: intPtr(-10)}
				return gen
			},
			expectError:   true,
			expectedField: "usage.output",
			expectedMsg:   "output token count cannot be negative",
		},
		{
			name: "negative total tokens",
			setupGen: func() *GenerationBuilder {
				gen := NewGenerationBuilder(client, "trace-id").Name("test")
				gen.usage = &types.Usage{Total: intPtr(-5)}
				return gen
			},
			expectError:   true,
			expectedField: "usage.total",
			expectedMsg:   "total token count cannot be negative",
		},
		{
			name: "zero token counts are valid",
			setupGen: func() *GenerationBuilder {
				gen := NewGenerationBuilder(client, "trace-id").Name("test")
				gen.usage = &types.Usage{
					Input:  intPtr(0),
					Output: intPtr(0),
					Total:  intPtr(0),
				}
				return gen
			},
			expectError: false,
		},
		{
			name: "valid token counts",
			setupGen: func() *GenerationBuilder {
				gen := NewGenerationBuilder(client, "trace-id").Name("test")
				gen.usage = &types.Usage{
					Input:  intPtr(100),
					Output: intPtr(150),
					Total:  intPtr(250),
				}
				return gen
			},
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := tt.setupGen()
			err := gen.validate()
			
			if tt.expectError {
				require.Error(t, err)
				validationErr, ok := err.(*ValidationError)
				require.True(t, ok, "expected ValidationError")
				
				if tt.expectedField != "" {
					assert.Equal(t, tt.expectedField, validationErr.Field)
				}
				if tt.expectedMsg != "" {
					assert.Equal(t, tt.expectedMsg, validationErr.Message)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSpanBuilder_ComprehensiveValidation(t *testing.T) {
	client := createTestClient(t)
	
	tests := []struct {
		name          string
		setupSpan     func() *SpanBuilder
		expectError   bool
		expectedField string
		expectedMsg   string
	}{
		{
			name: "valid minimal span",
			setupSpan: func() *SpanBuilder {
				return NewSpanBuilder(client, "trace-id").Name("test-span")
			},
			expectError: false,
		},
		{
			name: "valid full span",
			setupSpan: func() *SpanBuilder {
				return NewSpanBuilder(client, "trace-id").
					Name("full-span").
					Input("test input").
					Output("test output").
					AddMetadata("key", "value").
					Warning().
					StatusMessage("completed").
					Version("1.0")
			},
			expectError: false,
		},
		{
			name: "empty trace ID",
			setupSpan: func() *SpanBuilder {
				span := NewSpanBuilder(client, "")
				return span.Name("test-span")
			},
			expectError:   true,
			expectedField: "traceId",
			expectedMsg:   "trace id is required",
		},
		{
			name: "empty span name",
			setupSpan: func() *SpanBuilder {
				return NewSpanBuilder(client, "trace-id").Name("")
			},
			expectError:   true,
			expectedField: "name",
			expectedMsg:   "span name is required",
		},
		{
			name: "empty span ID",
			setupSpan: func() *SpanBuilder {
				span := NewSpanBuilder(client, "trace-id").Name("test")
				span.id = ""
				return span
			},
			expectError:   true,
			expectedField: "id",
			expectedMsg:   "span id is required",
		},
		{
			name: "zero start time",
			setupSpan: func() *SpanBuilder {
				span := NewSpanBuilder(client, "trace-id").Name("test")
				span.startTime = time.Time{}
				return span
			},
			expectError:   true,
			expectedField: "startTime",
			expectedMsg:   "start time is required",
		},
		{
			name: "end time before start time",
			setupSpan: func() *SpanBuilder {
				span := NewSpanBuilder(client, "trace-id").Name("test")
				now := time.Now()
				span.startTime = now
				span.endTime = &[]time.Time{now.Add(-time.Hour)}[0]
				return span
			},
			expectError:   true,
			expectedField: "endTime",
			expectedMsg:   "end time cannot be before start time",
		},
		{
			name: "valid timing",
			setupSpan: func() *SpanBuilder {
				span := NewSpanBuilder(client, "trace-id").Name("test")
				now := time.Now()
				span.startTime = now
				span.endTime = &[]time.Time{now.Add(time.Hour)}[0]
				return span
			},
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			span := tt.setupSpan()
			err := span.validate()
			
			if tt.expectError {
				require.Error(t, err)
				validationErr, ok := err.(*ValidationError)
				require.True(t, ok, "expected ValidationError")
				
				if tt.expectedField != "" {
					assert.Equal(t, tt.expectedField, validationErr.Field)
				}
				if tt.expectedMsg != "" {
					assert.Equal(t, tt.expectedMsg, validationErr.Message)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuilder_SubmissionStateValidation(t *testing.T) {
	client := createTestClient(t)
	
	t.Run("trace double submission", func(t *testing.T) {
		trace := client.Trace("test-trace")
		
		// First submission should succeed
		err := trace.Submit(context.Background())
		assert.NoError(t, err)
		assert.True(t, trace.submitted)
		
		// Second submission should fail
		err = trace.Submit(context.Background())
		require.Error(t, err)
		validationErr, ok := err.(*ValidationError)
		require.True(t, ok)
		assert.Equal(t, "state", validationErr.Field)
		assert.Contains(t, validationErr.Message, "already submitted")
	})
	
	t.Run("generation double submission", func(t *testing.T) {
		gen := NewGenerationBuilder(client, "trace-id").Name("test-gen")
		
		// First submission should succeed
		err := gen.Submit(context.Background())
		assert.NoError(t, err)
		assert.True(t, gen.submitted)
		
		// Second submission should fail
		err = gen.Submit(context.Background())
		require.Error(t, err)
		validationErr, ok := err.(*ValidationError)
		require.True(t, ok)
		assert.Equal(t, "state", validationErr.Field)
		assert.Contains(t, validationErr.Message, "already submitted")
	})
	
	t.Run("span double submission", func(t *testing.T) {
		span := NewSpanBuilder(client, "trace-id").Name("test-span")
		
		// First submission should succeed
		err := span.Submit(context.Background())
		assert.NoError(t, err)
		assert.True(t, span.submitted)
		
		// Second submission should fail
		err = span.Submit(context.Background())
		require.Error(t, err)
		validationErr, ok := err.(*ValidationError)
		require.True(t, ok)
		assert.Equal(t, "state", validationErr.Field)
		assert.Contains(t, validationErr.Message, "already submitted")
	})
}

func TestBuilder_ValidationBeforeSubmission(t *testing.T) {
	client := createTestClient(t)
	
	t.Run("invalid trace submission", func(t *testing.T) {
		trace := client.Trace("") // Empty name
		
		err := trace.Submit(context.Background())
		require.Error(t, err)
		assert.False(t, trace.submitted)
		
		validationErr, ok := err.(*ValidationError)
		require.True(t, ok)
		assert.Equal(t, "name", validationErr.Field)
	})
	
	t.Run("invalid generation submission", func(t *testing.T) {
		gen := NewGenerationBuilder(client, "trace-id").Name("") // Empty name
		
		err := gen.Submit(context.Background())
		require.Error(t, err)
		assert.False(t, gen.submitted)
		
		validationErr, ok := err.(*ValidationError)
		require.True(t, ok)
		assert.Equal(t, "name", validationErr.Field)
	})
	
	t.Run("invalid span submission", func(t *testing.T) {
		span := NewSpanBuilder(client, "trace-id").Name("") // Empty name
		
		err := span.Submit(context.Background())
		require.Error(t, err)
		assert.False(t, span.submitted)
		
		validationErr, ok := err.(*ValidationError)
		require.True(t, ok)
		assert.Equal(t, "name", validationErr.Field)
	})
}

func TestBuilder_ComplexValidationScenarios(t *testing.T) {
	client := createTestClient(t)
	
	t.Run("generation with complex timing", func(t *testing.T) {
		baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		
		gen := NewGenerationBuilder(client, "trace-id").
			Name("complex-timing").
			StartTime(baseTime).
			CompletionStartTime(baseTime.Add(30 * time.Second)).
			EndTime(baseTime.Add(60 * time.Second))
		
		err := gen.validate()
		assert.NoError(t, err)
	})
	
	t.Run("generation with invalid timing sequence", func(t *testing.T) {
		baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		
		gen := NewGenerationBuilder(client, "trace-id").
			Name("invalid-timing").
			StartTime(baseTime).
			CompletionStartTime(baseTime.Add(60 * time.Second)).
			EndTime(baseTime.Add(30 * time.Second)) // End before completion
		
		err := gen.validate()
		require.Error(t, err)
		validationErr, ok := err.(*ValidationError)
		require.True(t, ok)
		assert.Equal(t, "completionStartTime", validationErr.Field)
	})
	
	t.Run("usage validation edge cases", func(t *testing.T) {
		tests := []struct {
			name   string
			usage  *types.Usage
			valid  bool
			field  string
		}{
			{
				name:  "nil usage",
				usage: nil,
				valid: true,
			},
			{
				name: "empty usage",
				usage: &types.Usage{},
				valid: true,
			},
			{
				name: "zero values",
				usage: &types.Usage{
					Input:  intPtr(0),
					Output: intPtr(0),
					Total:  intPtr(0),
				},
				valid: true,
			},
			{
				name: "negative input only",
				usage: &types.Usage{
					Input: intPtr(-1),
				},
				valid: false,
				field: "usage.input",
			},
			{
				name: "negative output only",
				usage: &types.Usage{
					Output: intPtr(-1),
				},
				valid: false,
				field: "usage.output",
			},
			{
				name: "negative total only",
				usage: &types.Usage{
					Total: intPtr(-1),
				},
				valid: false,
				field: "usage.total",
			},
			{
				name: "mixed positive and nil",
				usage: &types.Usage{
					Input:  intPtr(100),
					Output: nil,
					Total:  intPtr(100),
				},
				valid: true,
			},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				gen := NewGenerationBuilder(client, "trace-id").Name("test")
				gen.usage = tt.usage
				
				err := gen.validate()
				if tt.valid {
					assert.NoError(t, err)
				} else {
					require.Error(t, err)
					validationErr, ok := err.(*ValidationError)
					require.True(t, ok)
					if tt.field != "" {
						assert.Equal(t, tt.field, validationErr.Field)
					}
				}
			})
		}
	})
}