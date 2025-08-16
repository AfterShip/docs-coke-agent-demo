package client

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"eino/pkg/langfuse/internal/queue"
)

func TestTraceBuilder_FluentAPI(t *testing.T) {
	client := createTestClient(t)
	
	trace := client.Trace("test-trace").
		UserID("user123").
		SessionID("session456").
		AddMetadata("key", "value").
		AddTag("tag1").
		AddTag("tag2").
		Version("1.0.0").
		Release("v1.0").
		Public(true)
	
	assert.Equal(t, "test-trace", trace.GetName())
	assert.Equal(t, "user123", trace.GetUserID())
	assert.Equal(t, "session456", trace.GetSessionID())
	assert.Equal(t, "value", trace.metadata["key"])
	assert.Contains(t, trace.tags, "tag1")
	assert.Contains(t, trace.tags, "tag2")
	assert.Equal(t, "1.0.0", *trace.version)
	assert.Equal(t, "v1.0", *trace.release)
	assert.True(t, *trace.public)
}

func TestTraceBuilder_WithAliases(t *testing.T) {
	client := createTestClient(t)
	metadata := map[string]interface{}{
		"environment": "test",
		"version": "1.0",
	}
	
	trace := client.Trace("test-trace").
		WithUser("user123").
		WithSession("session456").
		WithInput("test input").
		WithOutput("test output").
		WithMetadata(metadata).
		WithTags("tag1", "tag2")
	
	assert.Equal(t, "user123", trace.GetUserID())
	assert.Equal(t, "session456", trace.GetSessionID())
	assert.Equal(t, "test input", trace.input)
	assert.Equal(t, "test output", trace.output)
	assert.Equal(t, metadata, trace.metadata)
	assert.Equal(t, []string{"tag1", "tag2"}, trace.tags)
}

func TestTraceBuilder_IDGeneration(t *testing.T) {
	client := createTestClient(t)
	
	// Test auto-generated ID
	trace1 := client.Trace("trace1")
	assert.NotEmpty(t, trace1.GetID())
	
	// Test different traces have different IDs
	trace2 := client.Trace("trace2")
	assert.NotEqual(t, trace1.GetID(), trace2.GetID())
	
	// Test custom ID
	customID := "custom-trace-id"
	trace3 := client.Trace("trace3").ID(customID)
	assert.Equal(t, customID, trace3.GetID())
}

func TestTraceBuilder_TimestampHandling(t *testing.T) {
	client := createTestClient(t)
	
	// Test auto-generated timestamp
	trace := client.Trace("test-trace")
	assert.False(t, trace.timestamp.IsZero())
	
	// Test custom timestamp
	customTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	trace.Timestamp(customTime)
	assert.Equal(t, customTime, trace.timestamp)
	
	// Test timestamp is converted to UTC
	localTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.Local)
	trace.Timestamp(localTime)
	assert.Equal(t, time.UTC, trace.timestamp.Location())
}

func TestTraceBuilder_MetadataManipulation(t *testing.T) {
	client := createTestClient(t)
	
	trace := client.Trace("test-trace")
	
	// Test adding individual metadata
	trace.AddMetadata("key1", "value1")
	trace.AddMetadata("key2", 42)
	trace.AddMetadata("key3", true)
	
	assert.Equal(t, "value1", trace.metadata["key1"])
	assert.Equal(t, 42, trace.metadata["key2"])
	assert.Equal(t, true, trace.metadata["key3"])
	
	// Test replacing metadata map
	newMetadata := map[string]interface{}{
		"newKey": "newValue",
	}
	trace.Metadata(newMetadata)
	assert.Equal(t, newMetadata, trace.metadata)
	assert.NotContains(t, trace.metadata, "key1")
}

func TestTraceBuilder_TagManipulation(t *testing.T) {
	client := createTestClient(t)
	
	trace := client.Trace("test-trace")
	
	// Test adding individual tags
	trace.AddTag("tag1")
	trace.AddTag("tag2")
	
	assert.Contains(t, trace.tags, "tag1")
	assert.Contains(t, trace.tags, "tag2")
	assert.Len(t, trace.tags, 2)
	
	// Test replacing tags
	trace.Tags("newTag1", "newTag2", "newTag3")
	assert.Equal(t, []string{"newTag1", "newTag2", "newTag3"}, trace.tags)
	assert.NotContains(t, trace.tags, "tag1")
}

func TestTraceBuilder_Validation(t *testing.T) {
	client := createTestClient(t)
	
	tests := []struct {
		name        string
		setupTrace  func() *TraceBuilder
		expectError bool
		errorField  string
	}{
		{
			name: "valid trace",
			setupTrace: func() *TraceBuilder {
				return client.Trace("test-trace")
			},
			expectError: false,
		},
		{
			name: "empty name",
			setupTrace: func() *TraceBuilder {
				trace := client.Trace("")
				return trace
			},
			expectError: true,
			errorField:  "name",
		},
		{
			name: "zero timestamp",
			setupTrace: func() *TraceBuilder {
				trace := client.Trace("test")
				trace.timestamp = time.Time{}
				return trace
			},
			expectError: true,
			errorField:  "timestamp",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trace := tt.setupTrace()
			err := trace.validate()
			
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

func TestTraceBuilder_Submit(t *testing.T) {
	client := createTestClient(t)
	
	trace := client.Trace("test-trace").
		UserID("user123").
		Input("test input").
		Output("test output")
	
	err := trace.Submit(context.Background())
	assert.NoError(t, err)
	assert.True(t, trace.submitted)
	
	// Test submitting again should fail
	err = trace.Submit(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already submitted")
}

func TestTraceBuilder_Update(t *testing.T) {
	client := createTestClient(t)
	
	trace := client.Trace("test-trace").
		UserID("user123").
		Input("initial input")
	
	// Update with different data
	trace.Output("final output")
	
	err := trace.Update(context.Background())
	assert.NoError(t, err)
	assert.True(t, trace.submitted)
	
	// Test updating again should fail
	err = trace.Update(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already submitted")
}

func TestTraceBuilder_End(t *testing.T) {
	client := createTestClient(t)
	
	trace := client.Trace("test-trace")
	
	err := trace.End(context.Background())
	assert.NoError(t, err)
	assert.True(t, trace.submitted)
}

func TestTraceBuilder_EndAt(t *testing.T) {
	client := createTestClient(t)
	
	trace := client.Trace("test-trace")
	endTime := time.Now().UTC().Add(time.Hour)
	
	err := trace.EndAt(context.Background(), endTime)
	assert.NoError(t, err)
	assert.True(t, trace.submitted)
}

func TestTraceBuilder_SpanCreation(t *testing.T) {
	client := createTestClient(t)
	
	trace := client.Trace("test-trace")
	span := trace.Span("test-span")
	
	assert.Equal(t, "test-span", span.GetName())
	assert.Equal(t, trace.GetID(), span.GetTraceID())
	assert.NotEmpty(t, span.GetID())
}

func TestTraceBuilder_ImmutabilityAfterSubmit(t *testing.T) {
	client := createTestClient(t)
	
	trace := client.Trace("test-trace")
	originalID := trace.GetID()
	originalName := trace.GetName()
	
	// Submit the trace
	err := trace.Submit(context.Background())
	require.NoError(t, err)
	
	// Try to modify after submission - should not change anything
	trace.ID("new-id")
	trace.Name("new-name")
	trace.UserID("new-user")
	trace.AddMetadata("key", "value")
	trace.AddTag("new-tag")
	
	assert.Equal(t, originalID, trace.GetID())
	assert.Equal(t, originalName, trace.GetName())
	assert.Empty(t, trace.GetUserID())
	assert.Empty(t, trace.metadata)
	assert.Empty(t, trace.tags)
}

func TestTraceBuilder_EventConversion(t *testing.T) {
	client := createTestClient(t)
	
	trace := client.Trace("test-trace").
		UserID("user123").
		SessionID("session456").
		Input("test input").
		Output("test output").
		AddMetadata("key", "value").
		AddTag("tag1")
	
	// Test trace event conversion
	traceEvent := trace.toTraceEvent()
	assert.Equal(t, trace.id, traceEvent.ID)
	assert.Equal(t, trace.name, traceEvent.Name)
	assert.Equal(t, trace.userID, traceEvent.UserID)
	assert.Equal(t, trace.sessionID, traceEvent.SessionID)
	assert.Equal(t, trace.input, traceEvent.Input)
	assert.Equal(t, trace.output, traceEvent.Output)
	assert.Equal(t, trace.metadata, traceEvent.Metadata)
	assert.Equal(t, trace.tags, traceEvent.Tags)
	
	// Test trace create event conversion
	createEvent := trace.toTraceCreateEvent()
	assert.Equal(t, "trace-create", createEvent.Type)
	assert.Equal(t, trace.id, createEvent.TraceEvent.ID)
}

func TestTraceBuilder_ConcurrentAccess(t *testing.T) {
	client := createTestClient(t)
	
	trace := client.Trace("concurrent-trace")
	
	// Run concurrent operations
	done := make(chan bool, 10)
	
	// Start 10 goroutines modifying the trace
	for i := 0; i < 10; i++ {
		go func(i int) {
			defer func() { done <- true }()
			
			trace.AddMetadata(string(rune('a'+i)), i)
			trace.AddTag(string(rune('A'+i)))
		}(i)
	}
	
	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// The trace should have all metadata and tags
	assert.Len(t, trace.metadata, 10)
	assert.Len(t, trace.tags, 10)
}

// createTestClient creates a test Langfuse client with a mock queue
func createTestClient(t *testing.T) *Langfuse {
	config := &Config{
		Host:      "https://test.langfuse.com",
		PublicKey: "test-public-key",
		SecretKey: "test-secret-key",
		Enabled:   true,
	}
	
	// Create a test client with mock queue
	client := &Langfuse{
		config: config,
		queue:  queue.NewMockQueue(), // Mock queue for testing
	}
	
	return client
}