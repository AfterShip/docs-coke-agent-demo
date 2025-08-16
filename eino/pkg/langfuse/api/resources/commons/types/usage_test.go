package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsage_JSONSerialization(t *testing.T) {
	tests := []struct {
		name  string
		usage Usage
	}{
		{
			name: "complete usage",
			usage: Usage{
				Input:      intPtr(100),
				Output:     intPtr(200),
				Total:      intPtr(300),
				Unit:       stringPtr("TOKENS"),
				InputCost:  float64Ptr(0.001),
				OutputCost: float64Ptr(0.002),
				TotalCost:  float64Ptr(0.003),
			},
		},
		{
			name: "minimal usage with tokens only",
			usage: Usage{
				Input:  intPtr(50),
				Output: intPtr(75),
				Total:  intPtr(125),
			},
		},
		{
			name: "usage with cost only",
			usage: Usage{
				InputCost:  float64Ptr(0.005),
				OutputCost: float64Ptr(0.010),
				TotalCost:  float64Ptr(0.015),
			},
		},
		{
			name:  "empty usage",
			usage: Usage{},
		},
		{
			name: "usage with different unit",
			usage: Usage{
				Input: intPtr(1000),
				Total: intPtr(1000),
				Unit:  stringPtr("CHARACTERS"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			data, err := json.Marshal(tt.usage)
			require.NoError(t, err)
			assert.NotEmpty(t, data)

			// Test unmarshaling
			var unmarshaled Usage
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)

			// Verify fields
			if tt.usage.Input != nil {
				require.NotNil(t, unmarshaled.Input)
				assert.Equal(t, *tt.usage.Input, *unmarshaled.Input)
			} else {
				assert.Nil(t, unmarshaled.Input)
			}

			if tt.usage.Output != nil {
				require.NotNil(t, unmarshaled.Output)
				assert.Equal(t, *tt.usage.Output, *unmarshaled.Output)
			} else {
				assert.Nil(t, unmarshaled.Output)
			}

			if tt.usage.Total != nil {
				require.NotNil(t, unmarshaled.Total)
				assert.Equal(t, *tt.usage.Total, *unmarshaled.Total)
			} else {
				assert.Nil(t, unmarshaled.Total)
			}

			if tt.usage.Unit != nil {
				require.NotNil(t, unmarshaled.Unit)
				assert.Equal(t, *tt.usage.Unit, *unmarshaled.Unit)
			} else {
				assert.Nil(t, unmarshaled.Unit)
			}

			if tt.usage.InputCost != nil {
				require.NotNil(t, unmarshaled.InputCost)
				assert.Equal(t, *tt.usage.InputCost, *unmarshaled.InputCost)
			} else {
				assert.Nil(t, unmarshaled.InputCost)
			}

			if tt.usage.OutputCost != nil {
				require.NotNil(t, unmarshaled.OutputCost)
				assert.Equal(t, *tt.usage.OutputCost, *unmarshaled.OutputCost)
			} else {
				assert.Nil(t, unmarshaled.OutputCost)
			}

			if tt.usage.TotalCost != nil {
				require.NotNil(t, unmarshaled.TotalCost)
				assert.Equal(t, *tt.usage.TotalCost, *unmarshaled.TotalCost)
			} else {
				assert.Nil(t, unmarshaled.TotalCost)
			}
		})
	}
}

func TestUsage_CalculateTotalTokens(t *testing.T) {
	tests := []struct {
		name          string
		usage         Usage
		expectedTotal *int
		shouldModify  bool
	}{
		{
			name: "calculate when total is nil",
			usage: Usage{
				Input:  intPtr(100),
				Output: intPtr(200),
				Total:  nil,
			},
			expectedTotal: intPtr(300),
			shouldModify:  true,
		},
		{
			name: "don't calculate when total is already set",
			usage: Usage{
				Input:  intPtr(100),
				Output: intPtr(200),
				Total:  intPtr(250), // Different from input + output
			},
			expectedTotal: intPtr(250), // Should keep existing value
			shouldModify:  false,
		},
		{
			name: "don't calculate when input is nil",
			usage: Usage{
				Input:  nil,
				Output: intPtr(200),
				Total:  nil,
			},
			expectedTotal: nil,
			shouldModify:  false,
		},
		{
			name: "don't calculate when output is nil",
			usage: Usage{
				Input:  intPtr(100),
				Output: nil,
				Total:  nil,
			},
			expectedTotal: nil,
			shouldModify:  false,
		},
		{
			name: "handle zero values",
			usage: Usage{
				Input:  intPtr(0),
				Output: intPtr(0),
				Total:  nil,
			},
			expectedTotal: intPtr(0),
			shouldModify:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalTotal := tt.usage.Total

			tt.usage.CalculateTotalTokens()

			if tt.expectedTotal != nil {
				require.NotNil(t, tt.usage.Total)
				assert.Equal(t, *tt.expectedTotal, *tt.usage.Total)
			} else {
				assert.Nil(t, tt.usage.Total)
			}

			if tt.shouldModify {
				assert.NotEqual(t, originalTotal, tt.usage.Total)
			}
		})
	}
}

func TestUsage_CalculateTotalCost(t *testing.T) {
	tests := []struct {
		name              string
		usage             Usage
		expectedTotalCost *float64
		shouldModify      bool
	}{
		{
			name: "calculate when total cost is nil",
			usage: Usage{
				InputCost:  float64Ptr(0.001),
				OutputCost: float64Ptr(0.002),
				TotalCost:  nil,
			},
			expectedTotalCost: float64Ptr(0.003),
			shouldModify:      true,
		},
		{
			name: "don't calculate when total cost is already set",
			usage: Usage{
				InputCost:  float64Ptr(0.001),
				OutputCost: float64Ptr(0.002),
				TotalCost:  float64Ptr(0.005), // Different from input + output
			},
			expectedTotalCost: float64Ptr(0.005), // Should keep existing value
			shouldModify:      false,
		},
		{
			name: "don't calculate when input cost is nil",
			usage: Usage{
				InputCost:  nil,
				OutputCost: float64Ptr(0.002),
				TotalCost:  nil,
			},
			expectedTotalCost: nil,
			shouldModify:      false,
		},
		{
			name: "don't calculate when output cost is nil",
			usage: Usage{
				InputCost:  float64Ptr(0.001),
				OutputCost: nil,
				TotalCost:  nil,
			},
			expectedTotalCost: nil,
			shouldModify:      false,
		},
		{
			name: "handle zero values",
			usage: Usage{
				InputCost:  float64Ptr(0.0),
				OutputCost: float64Ptr(0.0),
				TotalCost:  nil,
			},
			expectedTotalCost: float64Ptr(0.0),
			shouldModify:      true,
		},
		{
			name: "handle floating point precision",
			usage: Usage{
				InputCost:  float64Ptr(0.0001),
				OutputCost: float64Ptr(0.0002),
				TotalCost:  nil,
			},
			expectedTotalCost: float64Ptr(0.0003),
			shouldModify:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalTotalCost := tt.usage.TotalCost

			tt.usage.CalculateTotalCost()

			if tt.expectedTotalCost != nil {
				require.NotNil(t, tt.usage.TotalCost)
				assert.InDelta(t, *tt.expectedTotalCost, *tt.usage.TotalCost, 0.000001)
			} else {
				assert.Nil(t, tt.usage.TotalCost)
			}

			if tt.shouldModify {
				assert.NotEqual(t, originalTotalCost, tt.usage.TotalCost)
			}
		})
	}
}

func TestNewUsage(t *testing.T) {
	tests := []struct {
		name         string
		inputTokens  int
		outputTokens int
	}{
		{
			name:         "positive tokens",
			inputTokens:  100,
			outputTokens: 200,
		},
		{
			name:         "zero tokens",
			inputTokens:  0,
			outputTokens: 0,
		},
		{
			name:         "large numbers",
			inputTokens:  1000000,
			outputTokens: 2000000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usage := NewUsage(tt.inputTokens, tt.outputTokens)

			require.NotNil(t, usage)
			require.NotNil(t, usage.Input)
			require.NotNil(t, usage.Output)
			require.NotNil(t, usage.Total)
			require.NotNil(t, usage.Unit)

			assert.Equal(t, tt.inputTokens, *usage.Input)
			assert.Equal(t, tt.outputTokens, *usage.Output)
			assert.Equal(t, tt.inputTokens+tt.outputTokens, *usage.Total)
			assert.Equal(t, "TOKENS", *usage.Unit)

			// Cost fields should be nil
			assert.Nil(t, usage.InputCost)
			assert.Nil(t, usage.OutputCost)
			assert.Nil(t, usage.TotalCost)
		})
	}
}

func TestNewUsageWithCost(t *testing.T) {
	tests := []struct {
		name         string
		inputTokens  int
		outputTokens int
		inputCost    float64
		outputCost   float64
	}{
		{
			name:         "typical usage",
			inputTokens:  100,
			outputTokens: 200,
			inputCost:    0.001,
			outputCost:   0.002,
		},
		{
			name:         "zero cost",
			inputTokens:  100,
			outputTokens: 200,
			inputCost:    0.0,
			outputCost:   0.0,
		},
		{
			name:         "high precision cost",
			inputTokens:  1000,
			outputTokens: 2000,
			inputCost:    0.000123,
			outputCost:   0.000456,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usage := NewUsageWithCost(tt.inputTokens, tt.outputTokens, tt.inputCost, tt.outputCost)

			require.NotNil(t, usage)
			require.NotNil(t, usage.Input)
			require.NotNil(t, usage.Output)
			require.NotNil(t, usage.Total)
			require.NotNil(t, usage.Unit)
			require.NotNil(t, usage.InputCost)
			require.NotNil(t, usage.OutputCost)
			require.NotNil(t, usage.TotalCost)

			assert.Equal(t, tt.inputTokens, *usage.Input)
			assert.Equal(t, tt.outputTokens, *usage.Output)
			assert.Equal(t, tt.inputTokens+tt.outputTokens, *usage.Total)
			assert.Equal(t, "TOKENS", *usage.Unit)

			assert.InDelta(t, tt.inputCost, *usage.InputCost, 0.000001)
			assert.InDelta(t, tt.outputCost, *usage.OutputCost, 0.000001)
			assert.InDelta(t, tt.inputCost+tt.outputCost, *usage.TotalCost, 0.000001)
		})
	}
}

func TestUsageCreateRequest_JSONSerialization(t *testing.T) {
	request := UsageCreateRequest{
		Input:      intPtr(100),
		Output:     intPtr(200),
		Total:      intPtr(300),
		Unit:       stringPtr("TOKENS"),
		InputCost:  float64Ptr(0.001),
		OutputCost: float64Ptr(0.002),
		TotalCost:  float64Ptr(0.003),
	}

	// Test marshaling
	data, err := json.Marshal(request)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test unmarshaling
	var unmarshaled UsageCreateRequest
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	// Verify fields
	assert.Equal(t, *request.Input, *unmarshaled.Input)
	assert.Equal(t, *request.Output, *unmarshaled.Output)
	assert.Equal(t, *request.Total, *unmarshaled.Total)
	assert.Equal(t, *request.Unit, *unmarshaled.Unit)
	assert.InDelta(t, *request.InputCost, *unmarshaled.InputCost, 0.000001)
	assert.InDelta(t, *request.OutputCost, *unmarshaled.OutputCost, 0.000001)
	assert.InDelta(t, *request.TotalCost, *unmarshaled.TotalCost, 0.000001)
}

func TestUsage_EdgeCases(t *testing.T) {
	t.Run("large token numbers", func(t *testing.T) {
		usage := Usage{
			Input:  intPtr(1000000),
			Output: intPtr(2000000),
		}

		usage.CalculateTotalTokens()

		require.NotNil(t, usage.Total)
		assert.Equal(t, 3000000, *usage.Total)
	})

	t.Run("negative token values in struct", func(t *testing.T) {
		// Note: The struct itself doesn't prevent negative values
		// Validation should be done at the application layer
		usage := Usage{
			Input:  intPtr(-100),
			Output: intPtr(-200),
		}

		usage.CalculateTotalTokens()

		require.NotNil(t, usage.Total)
		assert.Equal(t, -300, *usage.Total)
	})

	t.Run("very small cost values", func(t *testing.T) {
		usage := Usage{
			InputCost:  float64Ptr(0.000000001),
			OutputCost: float64Ptr(0.000000002),
		}

		usage.CalculateTotalCost()

		require.NotNil(t, usage.TotalCost)
		assert.InDelta(t, 0.000000003, *usage.TotalCost, 0.000000000001)
	})

	t.Run("chain calculations", func(t *testing.T) {
		usage := Usage{
			Input:      intPtr(100),
			Output:     intPtr(200),
			InputCost:  float64Ptr(0.001),
			OutputCost: float64Ptr(0.002),
		}

		// Both calculations should work
		usage.CalculateTotalTokens()
		usage.CalculateTotalCost()

		require.NotNil(t, usage.Total)
		require.NotNil(t, usage.TotalCost)
		assert.Equal(t, 300, *usage.Total)
		assert.InDelta(t, 0.003, *usage.TotalCost, 0.000001)
	})
}

// Helper functions for tests
func float64Ptr(f float64) *float64 {
	return &f
}
