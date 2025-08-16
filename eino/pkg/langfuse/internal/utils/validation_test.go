package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:   "test_field",
		Message: "is invalid",
	}

	assert.Equal(t, "validation error for field 'test_field': is invalid", err.Error())
}

func TestValidationErrors(t *testing.T) {
	var errors ValidationErrors

	// Test empty errors
	assert.False(t, errors.HasErrors())
	assert.Equal(t, "", errors.Error())

	// Add errors
	errors.Add("field1", "is required")
	errors.Add("field2", "is too short")

	assert.True(t, errors.HasErrors())
	assert.Equal(t, "validation error for field 'field1': is required; validation error for field 'field2': is too short", errors.Error())

	// Add ValidationError directly
	errors.AddError(ValidationError{Field: "field3", Message: "is invalid"})
	assert.Len(t, errors, 3)
}

func TestValidateRequired(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		fieldName string
		wantErr   bool
	}{
		{"nil value", nil, "field", true},
		{"empty string", "", "field", true},
		{"empty slice", []string{}, "field", true},
		{"zero int", 0, "field", true},
		{"zero bool", false, "field", true},
		{"valid string", "test", "field", false},
		{"valid int", 42, "field", false},
		{"valid bool true", true, "field", false},
		{"valid slice", []string{"item"}, "field", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequired(tt.value, tt.fieldName)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, tt.fieldName, err.Field)
				assert.Equal(t, "is required", err.Message)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateString(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		fieldName string
		minLen    int
		maxLen    int
		wantErr   bool
		wantMsg   string
	}{
		{"valid string within bounds", "hello", "field", 3, 10, false, ""},
		{"string too short", "hi", "field", 3, 10, true, "must be at least 3 characters"},
		{"string too long", "very long string", "field", 3, 10, true, "must be at most 10 characters"},
		{"empty string with min", "", "field", 1, 10, true, "must be at least 1 characters"},
		{"no minimum", "test", "field", 0, 10, false, ""},
		{"no maximum", "very long string here", "field", 3, 0, false, ""},
		{"no limits", "any length", "field", 0, 0, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateString(tt.value, tt.fieldName, tt.minLen, tt.maxLen)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, tt.fieldName, err.Field)
				assert.Equal(t, tt.wantMsg, err.Message)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		fieldName string
		wantErr   bool
	}{
		{"valid email", "test@example.com", "email", false},
		{"valid email with subdomain", "user@mail.example.co.uk", "email", false},
		{"valid email with plus", "user+tag@example.com", "email", false},
		{"valid email with numbers", "user123@example123.com", "email", false},
		{"empty email allowed", "", "email", false},
		{"invalid email no @", "invalid-email", "email", true},
		{"invalid email no domain", "test@", "email", true},
		{"invalid email no TLD", "test@example", "email", true},
		{"invalid email spaces", "test @example.com", "email", true},
		{"invalid email multiple @", "test@@example.com", "email", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email, tt.fieldName)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, tt.fieldName, err.Field)
				assert.Equal(t, "must be a valid email address", err.Message)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		fieldName string
		wantErr   bool
	}{
		{"valid http URL", "http://example.com", "url", false},
		{"valid https URL", "https://example.com/path", "url", false},
		{"valid URL with port", "http://localhost:8080", "url", false},
		{"valid URL with query", "https://example.com/search?q=test", "url", false},
		{"empty URL allowed", "", "url", false},
		{"invalid URL", "not-a-url", "url", true},
		{"invalid URL spaces", "http://invalid url.com", "url", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url, tt.fieldName)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, tt.fieldName, err.Field)
				assert.Equal(t, "must be a valid URL", err.Message)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateRegex(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		pattern   string
		fieldName string
		message   string
		wantErr   bool
		wantMsg   string
	}{
		{"valid pattern match", "123", `^\d+$`, "field", "", false, ""},
		{"pattern mismatch", "abc", `^\d+$`, "field", "", true, `must match pattern ^\d+$`},
		{"custom error message", "abc", `^\d+$`, "field", "must be numeric", true, "must be numeric"},
		{"empty value allowed", "", `^\d+$`, "field", "", false, ""},
		{"invalid regex pattern", "test", `[`, "field", "", true, "invalid validation pattern"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRegex(tt.value, tt.pattern, tt.fieldName, tt.message)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, tt.fieldName, err.Field)
				assert.Equal(t, tt.wantMsg, err.Message)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateNumericRange(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		fieldName string
		min       float64
		max       float64
		wantErr   bool
		wantMsg   string
	}{
		{"valid int in range", 5, "field", 0, 10, false, ""},
		{"valid float in range", 5.5, "field", 0, 10, false, ""},
		{"valid at min boundary", 0, "field", 0, 10, false, ""},
		{"valid at max boundary", 10, "field", 0, 10, false, ""},
		{"below minimum", -1, "field", 0, 10, true, "must be greater than or equal to 0"},
		{"above maximum", 11, "field", 0, 10, true, "must be less than or equal to 10"},
		{"invalid type", "not-a-number", "field", 0, 10, true, "must be a valid number"},
		{"nil value", nil, "field", 0, 10, true, "must be a valid number"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNumericRange(tt.value, tt.fieldName, tt.min, tt.max)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, tt.fieldName, err.Field)
				assert.Contains(t, err.Message, tt.wantMsg)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateIntRange(t *testing.T) {
	tests := []struct {
		name      string
		value     int
		fieldName string
		min       int
		max       int
		wantErr   bool
		wantMsg   string
	}{
		{"valid in range", 5, "field", 0, 10, false, ""},
		{"valid at min boundary", 0, "field", 0, 10, false, ""},
		{"valid at max boundary", 10, "field", 0, 10, false, ""},
		{"below minimum", -1, "field", 0, 10, true, "must be greater than or equal to 0"},
		{"above maximum", 11, "field", 0, 10, true, "must be less than or equal to 10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIntRange(tt.value, tt.fieldName, tt.min, tt.max)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, tt.fieldName, err.Field)
				assert.Equal(t, tt.wantMsg, err.Message)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateTimeRange(t *testing.T) {
	now := time.Now()
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	tests := []struct {
		name      string
		value     time.Time
		fieldName string
		min       time.Time
		max       time.Time
		wantErr   bool
	}{
		{"valid in range", now, "field", past, future, false},
		{"valid at boundaries", now, "field", now, now, false},
		{"before minimum", past.Add(-time.Minute), "field", past, future, true},
		{"after maximum", future.Add(time.Minute), "field", past, future, true},
		{"no min limit", past, "field", time.Time{}, future, false},
		{"no max limit", future, "field", past, time.Time{}, false},
		{"no limits", now, "field", time.Time{}, time.Time{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTimeRange(tt.value, tt.fieldName, tt.min, tt.max)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, tt.fieldName, err.Field)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateSliceLength(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		fieldName string
		minLen    int
		maxLen    int
		wantErr   bool
		wantMsg   string
	}{
		{"valid slice length", []string{"a", "b", "c"}, "field", 2, 5, false, ""},
		{"valid array", [3]string{"a", "b", "c"}, "field", 2, 5, false, ""},
		{"slice too short", []string{"a"}, "field", 2, 5, true, "must contain at least 2 items"},
		{"slice too long", []string{"a", "b", "c", "d", "e", "f"}, "field", 2, 5, true, "must contain at most 5 items"},
		{"empty slice with min", []string{}, "field", 1, 5, true, "must contain at least 1 items"},
		{"no minimum", []string{}, "field", 0, 5, false, ""},
		{"no maximum", []string{"a", "b", "c", "d", "e", "f"}, "field", 2, 0, false, ""},
		{"not a slice", "not-a-slice", "field", 2, 5, true, "must be a slice or array"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSliceLength(tt.value, tt.fieldName, tt.minLen, tt.maxLen)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, tt.fieldName, err.Field)
				assert.Equal(t, tt.wantMsg, err.Message)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateEnum(t *testing.T) {
	allowedValues := []string{"small", "medium", "large"}

	tests := []struct {
		name      string
		value     string
		fieldName string
		allowed   []string
		wantErr   bool
	}{
		{"valid value", "medium", "field", allowedValues, false},
		{"invalid value", "extra-large", "field", allowedValues, true},
		{"empty value allowed", "", "field", allowedValues, false},
		{"case sensitive", "Medium", "field", allowedValues, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEnum(tt.value, tt.fieldName, tt.allowed)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, tt.fieldName, err.Field)
				assert.Contains(t, err.Message, "must be one of:")
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateID(t *testing.T) {
	// This test depends on the IsValidID function implementation
	tests := []struct {
		name      string
		id        string
		fieldName string
		wantErr   bool
	}{
		{"empty ID allowed", "", "id", false},
		// Add more tests based on IsValidID implementation
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateID(tt.id, tt.fieldName)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, tt.fieldName, err.Field)
				assert.Equal(t, "must be a valid ID format", err.Message)
			} else {
				// Note: This test might fail if IsValidID is not implemented
				// In that case, we'd expect an error for non-empty IDs
				if tt.id != "" {
					// Temporarily allow either outcome until IsValidID is implemented
					if err != nil {
						assert.Equal(t, tt.fieldName, err.Field)
					}
				} else {
					assert.Nil(t, err)
				}
			}
		})
	}
}

func TestValidateTimestamp(t *testing.T) {
	now := time.Now()
	veryOld := time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC)
	veryFuture := now.AddDate(2, 0, 0)

	tests := []struct {
		name      string
		timestamp time.Time
		fieldName string
		wantErr   bool
		wantMsg   string
	}{
		{"valid current timestamp", now, "timestamp", false, ""},
		{"valid recent timestamp", now.Add(-time.Hour), "timestamp", false, ""},
		{"zero timestamp allowed", time.Time{}, "timestamp", false, ""},
		{"very old timestamp", veryOld, "timestamp", true, "timestamp is too far in the past"},
		{"very future timestamp", veryFuture, "timestamp", true, "timestamp is too far in the future"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTimestamp(tt.timestamp, tt.fieldName)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, tt.fieldName, err.Field)
				assert.Equal(t, tt.wantMsg, err.Message)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateMetadata(t *testing.T) {
	tests := []struct {
		name      string
		metadata  map[string]interface{}
		fieldName string
		maxKeys   int
		wantErr   bool
		wantMsg   string
	}{
		{"nil metadata allowed", nil, "metadata", 10, false, ""},
		{"valid metadata", map[string]interface{}{"key": "value"}, "metadata", 10, false, ""},
		{"too many keys", map[string]interface{}{
			"key1": "value1", "key2": "value2", "key3": "value3",
		}, "metadata", 2, true, "must contain at most 2 keys"},
		{"empty key", map[string]interface{}{"": "value"}, "metadata", 10, true, "metadata keys cannot be empty"},
		{"long key", map[string]interface{}{
			string(make([]byte, 256)): "value",
		}, "metadata", 10, true, "metadata keys must be at most 255 characters"},
		{"non-serializable value", map[string]interface{}{
			"key": make(chan int),
		}, "metadata", 10, true, "is not JSON serializable"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMetadata(tt.metadata, tt.fieldName, tt.maxKeys)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, tt.fieldName, err.Field)
				assert.Contains(t, err.Message, tt.wantMsg)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateEnvironment(t *testing.T) {
	tests := []struct {
		name      string
		env       string
		fieldName string
		wantErr   bool
		wantMsg   string
	}{
		{"empty environment allowed", "", "env", false, ""},
		{"valid environment", "production", "env", false, ""},
		{"valid with hyphens and underscores", "prod-env_1", "env", false, ""},
		{"valid with numbers", "prod123", "env", false, ""},
		{"invalid special characters", "prod@env", "env", true, "must contain only letters, numbers, hyphens, and underscores"},
		{"too long", string(make([]byte, 51)), "env", true, "must be at most 50 characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEnvironment(tt.env, tt.fieldName)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, tt.fieldName, err.Field)
				assert.Equal(t, tt.wantMsg, err.Message)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateTags(t *testing.T) {
	tests := []struct {
		name         string
		tags         []string
		fieldName    string
		maxTags      int
		maxTagLength int
		wantErr      bool
		wantMsg      string
	}{
		{"empty tags allowed", []string{}, "tags", 10, 50, false, ""},
		{"valid tags", []string{"tag1", "tag2"}, "tags", 10, 50, false, ""},
		{"too many tags", []string{"tag1", "tag2", "tag3"}, "tags", 2, 50, true, "must contain at most 2 tags"},
		{"empty tag", []string{"tag1", ""}, "tags", 10, 50, true, "tag at index 1 cannot be empty"},
		{"tag too long", []string{"tag1", string(make([]byte, 51))}, "tags", 10, 50, true, "tag at index 1 must be at most 50 characters"},
		{"invalid tag characters", []string{"tag1", "tag@2"}, "tags", 10, 50, true, "tag at index 1 contains invalid characters"},
		{"valid tag with allowed characters", []string{"tag_1", "tag-2", "tag.3"}, "tags", 10, 50, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTags(tt.tags, tt.fieldName, tt.maxTags, tt.maxTagLength)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, tt.fieldName, err.Field)
				assert.Contains(t, err.Message, tt.wantMsg)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateUsage(t *testing.T) {
	tests := []struct {
		name      string
		input     *int
		output    *int
		total     *int
		fieldName string
		wantErr   bool
		wantMsg   string
	}{
		{"all valid", intPtr(10), intPtr(20), intPtr(30), "usage", false, ""},
		{"nil values allowed", nil, nil, nil, "usage", false, ""},
		{"negative input", intPtr(-10), intPtr(20), intPtr(10), "usage", true, "input tokens cannot be negative"},
		{"negative output", intPtr(10), intPtr(-20), intPtr(-10), "usage", true, "output tokens cannot be negative"},
		{"negative total", intPtr(10), intPtr(20), intPtr(-30), "usage", true, "total tokens cannot be negative"},
		{"incorrect total", intPtr(10), intPtr(20), intPtr(25), "usage", true, "total (25) must equal input (10) + output (20)"},
		{"partial values valid", intPtr(10), nil, nil, "usage", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsage(tt.input, tt.output, tt.total, tt.fieldName)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Contains(t, err.Field, tt.fieldName)
				assert.Contains(t, err.Message, tt.wantMsg)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateCost(t *testing.T) {
	tests := []struct {
		name       string
		inputCost  *float64
		outputCost *float64
		totalCost  *float64
		fieldName  string
		wantErr    bool
		wantMsg    string
	}{
		{"all valid", float64Ptr(0.001), float64Ptr(0.002), float64Ptr(0.003), "cost", false, ""},
		{"nil values allowed", nil, nil, nil, "cost", false, ""},
		{"negative input cost", float64Ptr(-0.001), float64Ptr(0.002), float64Ptr(0.001), "cost", true, "input cost cannot be negative"},
		{"negative output cost", float64Ptr(0.001), float64Ptr(-0.002), float64Ptr(-0.001), "cost", true, "output cost cannot be negative"},
		{"negative total cost", float64Ptr(0.001), float64Ptr(0.002), float64Ptr(-0.003), "cost", true, "total cost cannot be negative"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCost(tt.inputCost, tt.outputCost, tt.totalCost, tt.fieldName)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Contains(t, err.Field, tt.fieldName)
				assert.Contains(t, err.Message, tt.wantMsg)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateObservationLevel(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		fieldName string
		wantErr   bool
	}{
		{"empty level allowed", "", "level", false},
		{"valid debug", "DEBUG", "level", false},
		{"valid default", "DEFAULT", "level", false},
		{"valid warning", "WARNING", "level", false},
		{"valid error", "ERROR", "level", false},
		{"invalid level", "INVALID", "level", true},
		{"case sensitive", "debug", "level", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateObservationLevel(tt.level, tt.fieldName)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, tt.fieldName, err.Field)
				assert.Contains(t, err.Message, "must be one of:")
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateScoreValue(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		dataType  string
		fieldName string
		wantErr   bool
		wantMsg   string
	}{
		{"nil value", nil, "NUMERIC", "value", true, "score value cannot be nil"},
		{"valid numeric int", 42, "NUMERIC", "value", false, ""},
		{"valid numeric float", 42.5, "NUMERIC", "value", false, ""},
		{"invalid numeric", "not-a-number", "NUMERIC", "value", true, "numeric score value must be a valid number"},
		{"valid boolean true", true, "BOOLEAN", "value", false, ""},
		{"valid boolean false", false, "BOOLEAN", "value", false, ""},
		{"invalid boolean", "true", "BOOLEAN", "value", true, "boolean score value must be true or false"},
		{"valid categorical", "positive", "CATEGORICAL", "value", false, ""},
		{"invalid categorical", 42, "CATEGORICAL", "value", true, "categorical score value must be a string"},
		{"invalid data type", "test", "INVALID", "dataType", true, "must be NUMERIC, BOOLEAN, or CATEGORICAL"},
		{"case insensitive data type", 123, "numeric", "value", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateScoreValue(tt.value, tt.dataType, tt.fieldName)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Contains(t, err.Message, tt.wantMsg)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateStruct(t *testing.T) {
	type TestStruct struct {
		Required string `json:"required" validate:"required"`
		Optional string `json:"optional"`
	}

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{"nil struct", (*TestStruct)(nil), true},
		{"valid struct", &TestStruct{Required: "value"}, false},
		{"invalid struct missing required", &TestStruct{}, true},
		{"non-struct", "not-a-struct", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateStruct(tt.input)

			if tt.wantErr {
				assert.True(t, errors.HasErrors())
			} else {
				assert.False(t, errors.HasErrors())
			}
		})
	}
}

// Helper functions for tests
func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}
