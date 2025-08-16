package utils

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTime(t *testing.T) {
	tests := []struct {
		name        string
		timeStr     string
		wantErr     bool
		expectValid bool
	}{
		{"empty string", "", true, false},
		{"RFC3339", "2024-01-15T12:00:00Z", false, true},
		{"RFC3339Nano", "2024-01-15T12:00:00.123456789Z", false, true},
		{"ISO format", "2024-01-15T12:00:00.000Z", false, true},
		{"date only", "2024-01-15", false, true},
		{"datetime without timezone", "2024-01-15T12:00:00", false, true},
		{"space separated", "2024-01-15 12:00:00", false, true},
		{"unix timestamp seconds", "1705320000", false, true},
		{"unix timestamp milliseconds", "1705320000000", false, true},
		{"invalid format", "invalid-date", true, false},
		{"kitchen time", "3:04PM", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTime(tt.timeStr)

			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, result.IsZero())
			} else {
				assert.NoError(t, err)
				if tt.expectValid {
					assert.False(t, result.IsZero())
				}
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 12, 30, 45, 123456789, time.UTC)

	formatted := FormatTime(testTime)

	assert.Equal(t, "2024-01-15T12:30:45.123456789Z", formatted)

	// Test that it's always UTC
	localTime := testTime.In(time.FixedZone("EST", -5*3600))
	formattedLocal := FormatTime(localTime)
	assert.Equal(t, "2024-01-15T12:30:45.123456789Z", formattedLocal) // Should be same time in UTC
}

func TestFormatTimePointer(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 12, 30, 45, 123456789, time.UTC)

	// Test with non-nil time
	result := FormatTimePointer(&testTime)
	require.NotNil(t, result)
	assert.Equal(t, "2024-01-15T12:30:45.123456789Z", *result)

	// Test with nil time
	result = FormatTimePointer(nil)
	assert.Nil(t, result)
}

func TestParseTimePointer(t *testing.T) {
	tests := []struct {
		name    string
		timeStr *string
		wantErr bool
		wantNil bool
	}{
		{"nil pointer", nil, false, true},
		{"empty string pointer", stringPtr(""), false, true},
		{"valid time string", stringPtr("2024-01-15T12:00:00Z"), false, false},
		{"invalid time string", stringPtr("invalid"), true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTimePointer(tt.timeStr)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.wantNil {
				assert.Nil(t, result)
			} else if !tt.wantErr {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestNowUTC(t *testing.T) {
	before := time.Now().UTC()
	result := NowUTC()
	after := time.Now().UTC()

	// Verify it's UTC
	assert.Equal(t, time.UTC, result.Location())

	// Verify it's within expected time range (allowing for small delays)
	assert.True(t, result.After(before.Add(-time.Second)))
	assert.True(t, result.Before(after.Add(time.Second)))
}

func TestNowUTCPointer(t *testing.T) {
	result := NowUTCPointer()

	require.NotNil(t, result)
	assert.Equal(t, time.UTC, result.Location())
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		want    float64
		wantErr bool
	}{
		{"nil value", nil, 0, true},
		{"float64", float64(42.5), 42.5, false},
		{"float32", float32(42.5), 42.5, false},
		{"int", int(42), 42.0, false},
		{"int8", int8(42), 42.0, false},
		{"int16", int16(42), 42.0, false},
		{"int32", int32(42), 42.0, false},
		{"int64", int64(42), 42.0, false},
		{"uint", uint(42), 42.0, false},
		{"uint8", uint8(42), 42.0, false},
		{"uint16", uint16(42), 42.0, false},
		{"uint32", uint32(42), 42.0, false},
		{"uint64", uint64(42), 42.0, false},
		{"string number", "42.5", 42.5, false},
		{"string invalid", "not-a-number", 0, true},
		{"json.Number", json.Number("42.5"), 42.5, false},
		{"unsupported type", []int{1, 2, 3}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ToFloat64(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestToInt64(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		want    int64
		wantErr bool
	}{
		{"nil value", nil, 0, true},
		{"int64", int64(42), 42, false},
		{"int", int(42), 42, false},
		{"int8", int8(42), 42, false},
		{"int16", int16(42), 42, false},
		{"int32", int32(42), 42, false},
		{"uint", uint(42), 42, false},
		{"uint8", uint8(42), 42, false},
		{"uint16", uint16(42), 42, false},
		{"uint32", uint32(42), 42, false},
		{"uint64 valid", uint64(42), 42, false},
		{"uint64 overflow", uint64(9223372036854775808), 0, true}, // max int64 + 1
		{"float32", float32(42.7), 42, false},
		{"float64", float64(42.7), 42, false},
		{"string number", "42", 42, false},
		{"string invalid", "not-a-number", 0, true},
		{"json.Number", json.Number("42"), 42, false},
		{"unsupported type", []int{1, 2, 3}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ToInt64(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestToString(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name  string
		value interface{}
		want  string
	}{
		{"nil", nil, ""},
		{"string", "hello", "hello"},
		{"byte slice", []byte("hello"), "hello"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"int", int(42), "42"},
		{"int8", int8(42), "42"},
		{"int16", int16(42), "42"},
		{"int32", int32(42), "42"},
		{"int64", int64(42), "42"},
		{"uint", uint(42), "42"},
		{"uint8", uint8(42), "42"},
		{"uint16", uint16(42), "42"},
		{"uint32", uint32(42), "42"},
		{"uint64", uint64(42), "42"},
		{"float32", float32(42.5), "42.5"},
		{"float64", float64(42.5), "42.5"},
		{"time", testTime, "2024-01-15T12:00:00Z"},
		{"other type", []int{1, 2}, "[1 2]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToString(tt.value)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestToStringPointer(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		wantNil  bool
		expected string
	}{
		{"nil value", nil, true, ""},
		{"empty string", "", true, ""},
		{"valid string", "hello", false, "hello"},
		{"zero int", 0, true, ""},
		{"non-zero int", 42, false, "42"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToStringPointer(tt.value)

			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expected, *result)
			}
		})
	}
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  bool
	}{
		{"nil", nil, true},
		{"empty string", "", true},
		{"non-empty string", "hello", false},
		{"false bool", false, true},
		{"true bool", true, false},
		{"zero int", 0, true},
		{"non-zero int", 42, false},
		{"zero uint", uint(0), true},
		{"non-zero uint", uint(42), false},
		{"zero float", 0.0, true},
		{"non-zero float", 42.5, false},
		{"nil pointer", (*string)(nil), true},
		{"non-nil pointer", stringPtr("hello"), false},
		{"empty slice", []string{}, true},
		{"non-empty slice", []string{"item"}, false},
		{"empty map", map[string]string{}, true},
		{"non-empty map", map[string]string{"key": "value"}, false},
		{"zero time", time.Time{}, true},
		{"non-zero time", time.Now(), false},
		{"struct", struct{}{}, false}, // structs are never considered empty unless they're time.Time
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEmpty(tt.value)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestIsValidJSON(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want bool
	}{
		{"valid object", `{"key": "value"}`, true},
		{"valid array", `[1, 2, 3]`, true},
		{"valid string", `"hello"`, true},
		{"valid number", `42`, true},
		{"valid boolean", `true`, true},
		{"valid null", `null`, true},
		{"invalid json", `{key: "value"}`, false}, // missing quotes around key
		{"invalid json", `{"key": value}`, false}, // missing quotes around value
		{"empty string", ``, false},
		{"malformed", `{`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidJSON(tt.str)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestIsJSONSerializable(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  bool
	}{
		{"string", "hello", true},
		{"int", 42, true},
		{"float", 42.5, true},
		{"bool", true, true},
		{"slice", []string{"a", "b"}, true},
		{"map", map[string]interface{}{"key": "value"}, true},
		{"struct", struct{ Name string }{Name: "test"}, true},
		{"channel", make(chan int), false},
		{"function", func() {}, false},
		{"complex", complex(1, 2), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsJSONSerializable(tt.value)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestToJSON(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		want    string
		wantErr bool
	}{
		{"string", "hello", `"hello"`, false},
		{"int", 42, "42", false},
		{"map", map[string]string{"key": "value"}, `{"key":"value"}`, false},
		{"slice", []int{1, 2, 3}, "[1,2,3]", false},
		{"channel", make(chan int), "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ToJSON(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestFromJSON(t *testing.T) {
	tests := []struct {
		name    string
		jsonStr string
		target  interface{}
		wantErr bool
	}{
		{"valid string", `"hello"`, new(string), false},
		{"valid int", "42", new(int), false},
		{"valid map", `{"key":"value"}`, new(map[string]string), false},
		{"invalid json", `{invalid}`, new(map[string]string), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := FromJSON(tt.jsonStr, tt.target)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestToJSONIndent(t *testing.T) {
	value := map[string]interface{}{
		"key1": "value1",
		"key2": map[string]interface{}{
			"nested": "value",
		},
	}

	result, err := ToJSONIndent(value, "  ")
	require.NoError(t, err)

	// Should contain proper indentation
	assert.Contains(t, result, "\n")
	assert.Contains(t, result, "  ")

	// Should be valid JSON
	var parsed map[string]interface{}
	err = json.Unmarshal([]byte(result), &parsed)
	assert.NoError(t, err)
}

func TestMergeMetadata(t *testing.T) {
	tests := []struct {
		name     string
		base     map[string]interface{}
		override map[string]interface{}
		want     map[string]interface{}
	}{
		{
			name:     "both nil",
			base:     nil,
			override: nil,
			want:     nil,
		},
		{
			name:     "base nil",
			base:     nil,
			override: map[string]interface{}{"key": "value"},
			want:     map[string]interface{}{"key": "value"},
		},
		{
			name:     "override nil",
			base:     map[string]interface{}{"key": "value"},
			override: nil,
			want:     map[string]interface{}{"key": "value"},
		},
		{
			name:     "merge without conflict",
			base:     map[string]interface{}{"key1": "value1"},
			override: map[string]interface{}{"key2": "value2"},
			want:     map[string]interface{}{"key1": "value1", "key2": "value2"},
		},
		{
			name:     "override existing key",
			base:     map[string]interface{}{"key": "original"},
			override: map[string]interface{}{"key": "overridden"},
			want:     map[string]interface{}{"key": "overridden"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeMetadata(tt.base, tt.override)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestCloneMetadata(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]interface{}
		wantNil  bool
	}{
		{"nil metadata", nil, true},
		{"empty metadata", map[string]interface{}{}, false},
		{"simple metadata", map[string]interface{}{"key": "value"}, false},
		{
			"complex metadata",
			map[string]interface{}{
				"string": "value",
				"number": 42,
				"nested": map[string]interface{}{"inner": "value"},
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CloneMetadata(tt.metadata)

			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.metadata, result)

				// Verify it's a deep copy by modifying original
				if tt.metadata != nil && len(tt.metadata) > 0 {
					// Modify original
					for k := range tt.metadata {
						tt.metadata[k] = "modified"
						break
					}

					// Result should not be affected
					assert.NotEqual(t, tt.metadata, result)
				}
			}
		})
	}
}

func TestFilterMetadata(t *testing.T) {
	metadata := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	tests := []struct {
		name     string
		metadata map[string]interface{}
		keys     []string
		want     map[string]interface{}
		wantNil  bool
	}{
		{"nil metadata", nil, []string{"key1"}, nil, true},
		{"empty keys", metadata, []string{}, nil, true},
		{"filter existing keys", metadata, []string{"key1", "key3"}, map[string]interface{}{"key1": "value1", "key3": "value3"}, false},
		{"filter non-existing keys", metadata, []string{"nonexistent"}, nil, true},
		{"mix existing and non-existing", metadata, []string{"key1", "nonexistent"}, map[string]interface{}{"key1": "value1"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterMetadata(tt.metadata, tt.keys)

			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestSanitizeMetadata(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]interface{}
		want     map[string]interface{}
		wantNil  bool
	}{
		{"nil metadata", nil, nil, true},
		{"empty metadata", map[string]interface{}{}, nil, true},
		{
			"remove nil and empty strings",
			map[string]interface{}{
				"valid":        "value",
				"nil_value":    nil,
				"empty_string": "",
				"zero_int":     0, // should be kept
			},
			map[string]interface{}{
				"valid":    "value",
				"zero_int": 0,
			},
			false,
		},
		{
			"all invalid values",
			map[string]interface{}{
				"nil_value":    nil,
				"empty_string": "",
			},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeMetadata(tt.metadata)

			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestConvertToMap(t *testing.T) {
	type TestStruct struct {
		Name     string  `json:"name"`
		Age      int     `json:"age"`
		Optional *string `json:"optional,omitempty"`
	}

	tests := []struct {
		name    string
		value   interface{}
		want    map[string]interface{}
		wantErr bool
	}{
		{
			"simple struct",
			TestStruct{Name: "John", Age: 30},
			map[string]interface{}{"name": "John", "age": float64(30)}, // JSON numbers become float64
			false,
		},
		{
			"struct with nil pointer",
			TestStruct{Name: "John", Age: 30, Optional: nil},
			map[string]interface{}{"name": "John", "age": float64(30)},
			false,
		},
		{
			"struct with pointer value",
			TestStruct{Name: "John", Age: 30, Optional: stringPtr("value")},
			map[string]interface{}{"name": "John", "age": float64(30), "optional": "value"},
			false,
		},
		{
			"map",
			map[string]string{"key": "value"},
			map[string]interface{}{"key": "value"},
			false,
		},
		{
			"unmarshalable value",
			make(chan int),
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertToMap(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestDurationFunctions(t *testing.T) {
	start := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Hour)

	t.Run("GetDurationBetween", func(t *testing.T) {
		duration := GetDurationBetween(start, end)
		assert.Equal(t, 2*time.Hour, duration)
	})

	t.Run("GetDurationFromNow", func(t *testing.T) {
		future := time.Now().Add(time.Hour)
		duration := GetDurationFromNow(future)

		// Should be approximately 1 hour (allowing for test execution time)
		assert.True(t, duration > 59*time.Minute)
		assert.True(t, duration < 61*time.Minute)
	})

	t.Run("GetDurationToNow", func(t *testing.T) {
		past := time.Now().Add(-time.Hour)
		duration := GetDurationToNow(past)

		// Should be approximately 1 hour (allowing for test execution time)
		assert.True(t, duration > 59*time.Minute)
		assert.True(t, duration < 61*time.Minute)
	})
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"seconds", 30 * time.Second, "30.00s"},
		{"minutes", 5 * time.Minute, "5.00m"},
		{"hours", 3 * time.Hour, "3.00h"},
		{"days", 2 * 24 * time.Hour, "2.00d"},
		{"mixed duration", 90 * time.Minute, "1.50h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.duration)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name   string
		str    string
		maxLen int
		want   string
	}{
		{"no truncation needed", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"truncate with ellipsis", "hello world", 8, "hello..."},
		{"very short max", "hello", 3, "hel"},
		{"very short max with ellipsis", "hello", 2, "he"},
		{"empty string", "", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateString(tt.str, tt.maxLen)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestCoalesceString(t *testing.T) {
	tests := []struct {
		name    string
		strings []string
		want    string
	}{
		{"first non-empty", []string{"", "second", "third"}, "second"},
		{"all empty", []string{"", "", ""}, ""},
		{"first non-empty is first", []string{"first", "second", "third"}, "first"},
		{"no strings", []string{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CoalesceString(tt.strings...)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestCoalescePointer(t *testing.T) {
	str1 := "first"
	str2 := "second"

	tests := []struct {
		name     string
		pointers []*string
		want     *string
	}{
		{"first non-nil", []*string{nil, &str2}, &str2},
		{"all nil", []*string{nil, nil}, nil},
		{"first non-nil is first", []*string{&str1, &str2}, &str1},
		{"no pointers", []*string{}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CoalescePointer(tt.pointers...)
			if tt.want == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, *tt.want, *result)
			}
		})
	}
}

func TestDerefPointer(t *testing.T) {
	value := "test"
	defaultValue := "default"

	// Test with non-nil pointer
	result := DerefPointer(&value, defaultValue)
	assert.Equal(t, value, result)

	// Test with nil pointer
	result = DerefPointer((*string)(nil), defaultValue)
	assert.Equal(t, defaultValue, result)
}

func TestToPointer(t *testing.T) {
	value := "test"
	result := ToPointer(value)

	require.NotNil(t, result)
	assert.Equal(t, value, *result)
}

// Helper function for tests
func stringPtr(s string) *string {
	return &s
}
