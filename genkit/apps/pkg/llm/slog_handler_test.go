package llm

import (
	"context"
	"log/slog"
	"testing"
	"time"
)

func TestSlogAttrToZapField_SafeHandling(t *testing.T) {
	tests := []struct {
		name string
		attr slog.Attr
	}{
		{
			name: "string attribute",
			attr: slog.String("key", "value"),
		},
		{
			name: "int64 attribute",
			attr: slog.Int64("count", 42),
		},
		{
			name: "bool attribute",
			attr: slog.Bool("flag", true),
		},
		{
			name: "time attribute",
			attr: slog.Time("timestamp", time.Now()),
		},
		{
			name: "group attribute - should not cause JSON serialization error",
			attr: slog.Group("group", slog.String("nested", "value")),
		},
		{
			name: "function attribute - should get function name",
			attr: slog.Any("callback", func() string { return "test" }),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This should not panic or cause JSON serialization errors
			field := slogAttrToZapField(tt.attr, []string{})

			// Verify field has a key
			if field.Key == "" {
				t.Error("Field key should not be empty")
			}

			// Verify the function can handle the conversion without errors
			// The main test is that this doesn't panic or cause JSON serialization issues
		})
	}
}

func TestHandleUnknownType_FunctionHandling(t *testing.T) {
	tests := []struct {
		name     string
		value    slog.Value
		expected string // expected prefix or content
	}{
		{
			name:     "named function",
			value:    slog.AnyValue(testFunction),
			expected: "func:",
		},
		{
			name:     "anonymous function",
			value:    slog.AnyValue(func() string { return "anonymous" }),
			expected: "func:",
		},
		{
			name:     "non-function value",
			value:    slog.AnyValue(map[string]int{"key": 1}),
			expected: "map[", // should fallback to string representation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := handleUnknownType("test_key", tt.value)

			// Verify field key
			if field.Key != "test_key" {
				t.Errorf("Expected key 'test_key', got '%s'", field.Key)
			}

			// For functions, verify it starts with "func:"
			// For other types, verify it contains expected content
			fieldStr := field.String
			if tt.expected == "func:" {
				if !startsWith(fieldStr, "func:") {
					t.Errorf("Expected function field to start with 'func:', got '%s'", fieldStr)
				}
			} else {
				if !contains(fieldStr, tt.expected) {
					t.Errorf("Expected field to contain '%s', got '%s'", tt.expected, fieldStr)
				}
			}
		})
	}
}

// testFunction is a named function for testing
func testFunction() string {
	return "test"
}

// Helper functions for string checking
func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestSlogHandler_HandleWithPotentiallyProblematicAttrs(t *testing.T) {
	handler := NewSlogHandler()

	// Create a record with various attribute types
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
	record.AddAttrs(
		slog.String("string", "value"),
		slog.Int("int", 42),
		slog.Group("group", slog.String("nested", "value")),
	)

	// This should not panic or cause JSON serialization errors
	err := handler.Handle(context.Background(), record)
	if err != nil {
		t.Errorf("Handler.Handle() returned error: %v", err)
	}
}

func TestSlogLevelToZapLevel(t *testing.T) {
	tests := []struct {
		name      string
		slogLevel slog.Level
		expected  string // We'll compare the string representation
	}{
		{
			name:      "debug level",
			slogLevel: slog.LevelDebug,
			expected:  "debug",
		},
		{
			name:      "info level",
			slogLevel: slog.LevelInfo,
			expected:  "info",
		},
		{
			name:      "warn level",
			slogLevel: slog.LevelWarn,
			expected:  "warn",
		},
		{
			name:      "error level",
			slogLevel: slog.LevelError,
			expected:  "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zapLevel := slogLevelToZapLevel(tt.slogLevel)
			if zapLevel.String() != tt.expected {
				t.Errorf("slogLevelToZapLevel() = %v, expected %v", zapLevel.String(), tt.expected)
			}
		})
	}
}
