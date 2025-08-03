package llm

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"
)

// DemoFunctionHandling demonstrates how function types are handled in slog
func TestDemoFunctionHandling(t *testing.T) {
	// Create a handler
	handler := NewSlogHandler()

	// Named function
	namedFunc := func() string { return "named function" }

	// Create a log record with function attributes
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "Demo: Function handling in slog", 0)
	record.AddAttrs(
		slog.String("normal", "regular string value"),
		slog.Any("named_func", namedFunc),
		slog.Any("anonymous_func", func() int { return 42 }),
		slog.Any("method_func", strings.Contains), // method function
	)

	// This should handle functions gracefully without JSON serialization errors
	err := handler.Handle(context.Background(), record)
	if err != nil {
		t.Errorf("Handler.Handle() returned error: %v", err)
	}

	// Test individual function handling
	t.Run("individual function tests", func(t *testing.T) {
		testCases := []struct {
			name     string
			value    slog.Value
			contains string
		}{
			{
				name:     "named function",
				value:    slog.AnyValue(namedFunc),
				contains: "func:",
			},
			{
				name:     "anonymous function",
				value:    slog.AnyValue(func() { /* anonymous */ }),
				contains: "func:",
			},
			{
				name:     "method function",
				value:    slog.AnyValue(strings.Contains),
				contains: "func:",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				field := handleUnknownType("test", tc.value)

				if !strings.Contains(field.String, tc.contains) {
					t.Errorf("Expected field to contain '%s', got '%s'", tc.contains, field.String)
				}

				// Log the actual output for manual verification
				t.Logf("Function type '%s' resulted in: %s", tc.name, field.String)
			})
		}
	})
}
