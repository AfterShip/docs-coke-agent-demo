package utils

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"time"
)

// Note: ValidationError is defined in errors.go to avoid conflicts

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// HasErrors returns true if there are any validation errors
func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}

// Add adds a validation error to the collection
func (e *ValidationErrors) Add(field, message string) {
	*e = append(*e, ValidationError{Field: field, Message: message})
}

// AddError adds a ValidationError to the collection
func (e *ValidationErrors) AddError(err ValidationError) {
	*e = append(*e, err)
}

// ValidateRequired validates that a field is not empty
func ValidateRequired(value interface{}, fieldName string) *ValidationError {
	if IsEmpty(value) {
		return &ValidationError{Field: fieldName, Message: "is required"}
	}
	return nil
}

// ValidateString validates string fields
func ValidateString(value string, fieldName string, minLen, maxLen int) *ValidationError {
	if minLen > 0 && len(value) < minLen {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf("must be at least %d characters", minLen)}
	}

	if maxLen > 0 && len(value) > maxLen {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf("must be at most %d characters", maxLen)}
	}

	return nil
}

// ValidateEmail validates email format
func ValidateEmail(email, fieldName string) *ValidationError {
	if email == "" {
		return nil // Allow empty emails, use ValidateRequired for required validation
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return &ValidationError{Field: fieldName, Message: "must be a valid email address"}
	}

	return nil
}

// ValidateURL validates URL format
func ValidateURL(urlStr, fieldName string) *ValidationError {
	if urlStr == "" {
		return nil // Allow empty URLs, use ValidateRequired for required validation
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return &ValidationError{Field: fieldName, Message: "must be a valid URL"}
	}

	// Additional validation to ensure it's a proper URL
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return &ValidationError{Field: fieldName, Message: "must be a valid URL"}
	}

	return nil
}

// ValidateRegex validates string against a regex pattern
func ValidateRegex(value, pattern, fieldName, message string) *ValidationError {
	if value == "" {
		return nil // Allow empty values, use ValidateRequired for required validation
	}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return &ValidationError{Field: fieldName, Message: "invalid validation pattern"}
	}

	if !regex.MatchString(value) {
		if message == "" {
			message = fmt.Sprintf("must match pattern %s", pattern)
		}
		return &ValidationError{Field: fieldName, Message: message}
	}

	return nil
}

// ValidateNumericRange validates numeric values within a range
func ValidateNumericRange(value interface{}, fieldName string, min, max float64) *ValidationError {
	floatValue, err := ToFloat64(value)
	if err != nil {
		return &ValidationError{Field: fieldName, Message: "must be a valid number"}
	}

	if floatValue < min {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf("must be greater than or equal to %g", min)}
	}

	if floatValue > max {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf("must be less than or equal to %g", max)}
	}

	return nil
}

// ValidateIntRange validates integer values within a range
func ValidateIntRange(value int, fieldName string, min, max int) *ValidationError {
	if value < min {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf("must be greater than or equal to %d", min)}
	}

	if value > max {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf("must be less than or equal to %d", max)}
	}

	return nil
}

// ValidateTimeRange validates that a time is within a range
func ValidateTimeRange(value time.Time, fieldName string, min, max time.Time) *ValidationError {
	if !min.IsZero() && value.Before(min) {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf("must be after %s", min.Format(time.RFC3339))}
	}

	if !max.IsZero() && value.After(max) {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf("must be before %s", max.Format(time.RFC3339))}
	}

	return nil
}

// ValidateSliceLength validates slice length
func ValidateSliceLength(value interface{}, fieldName string, minLen, maxLen int) *ValidationError {
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return &ValidationError{Field: fieldName, Message: "must be a slice or array"}
	}

	length := v.Len()

	if minLen > 0 && length < minLen {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf("must contain at least %d items", minLen)}
	}

	if maxLen > 0 && length > maxLen {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf("must contain at most %d items", maxLen)}
	}

	return nil
}

// ValidateEnum validates that a value is one of the allowed values
func ValidateEnum(value string, fieldName string, allowedValues []string) *ValidationError {
	if value == "" {
		return nil // Allow empty values, use ValidateRequired for required validation
	}

	for _, allowed := range allowedValues {
		if value == allowed {
			return nil
		}
	}

	return &ValidationError{Field: fieldName, Message: fmt.Sprintf("must be one of: %s", strings.Join(allowedValues, ", "))}
}

// ValidateID validates ID format (nanoid or UUID)
func ValidateID(id, fieldName string) *ValidationError {
	if id == "" {
		return nil // Allow empty IDs, use ValidateRequired for required validation
	}

	if !IsValidID(id) {
		return &ValidationError{Field: fieldName, Message: "must be a valid ID format"}
	}

	return nil
}

// ValidateJSON validates JSON format
func ValidateJSON(jsonStr, fieldName string) *ValidationError {
	if jsonStr == "" {
		return nil // Allow empty JSON, use ValidateRequired for required validation
	}

	if !IsValidJSON(jsonStr) {
		return &ValidationError{Field: fieldName, Message: "must be valid JSON"}
	}

	return nil
}

// ValidateTimestamp validates timestamp format and reasonableness
func ValidateTimestamp(timestamp time.Time, fieldName string) *ValidationError {
	if timestamp.IsZero() {
		return nil // Allow zero timestamps, use ValidateRequired for required validation
	}

	// Check if timestamp is too far in the past (before 2000)
	if timestamp.Before(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)) {
		return &ValidationError{Field: fieldName, Message: "timestamp is too far in the past"}
	}

	// Check if timestamp is too far in the future (more than 1 year from now)
	futureLimit := time.Now().AddDate(1, 0, 0)
	if timestamp.After(futureLimit) {
		return &ValidationError{Field: fieldName, Message: "timestamp is too far in the future"}
	}

	return nil
}

// ValidateMetadata validates metadata object structure
func ValidateMetadata(metadata map[string]interface{}, fieldName string, maxKeys int) *ValidationError {
	if metadata == nil {
		return nil
	}

	if maxKeys > 0 && len(metadata) > maxKeys {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf("must contain at most %d keys", maxKeys)}
	}

	// Validate that all keys are strings and values are JSON-serializable
	for key, value := range metadata {
		if key == "" {
			return &ValidationError{Field: fieldName, Message: "metadata keys cannot be empty"}
		}

		if len(key) > 255 {
			return &ValidationError{Field: fieldName, Message: "metadata keys must be at most 255 characters"}
		}

		if !IsJSONSerializable(value) {
			return &ValidationError{Field: fieldName, Message: fmt.Sprintf("metadata value for key '%s' is not JSON serializable", key)}
		}
	}

	return nil
}

// ValidateEnvironment validates environment name format
func ValidateEnvironment(env, fieldName string) *ValidationError {
	if env == "" {
		return nil // Allow empty environment, use ValidateRequired for required validation
	}

	// Environment should be alphanumeric with hyphens and underscores
	envRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !envRegex.MatchString(env) {
		return &ValidationError{Field: fieldName, Message: "must contain only letters, numbers, hyphens, and underscores"}
	}

	if len(env) > 50 {
		return &ValidationError{Field: fieldName, Message: "must be at most 50 characters"}
	}

	return nil
}

// ValidateTags validates tags array
func ValidateTags(tags []string, fieldName string, maxTags, maxTagLength int) *ValidationError {
	if len(tags) == 0 {
		return nil
	}

	if maxTags > 0 && len(tags) > maxTags {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf("must contain at most %d tags", maxTags)}
	}

	for i, tag := range tags {
		if tag == "" {
			return &ValidationError{Field: fieldName, Message: fmt.Sprintf("tag at index %d cannot be empty", i)}
		}

		if maxTagLength > 0 && len(tag) > maxTagLength {
			return &ValidationError{Field: fieldName, Message: fmt.Sprintf("tag at index %d must be at most %d characters", i, maxTagLength)}
		}

		// Tags should not contain special characters that might cause issues
		tagRegex := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
		if !tagRegex.MatchString(tag) {
			return &ValidationError{Field: fieldName, Message: fmt.Sprintf("tag at index %d contains invalid characters", i)}
		}
	}

	return nil
}

// ValidateUsage validates usage metrics
func ValidateUsage(input, output, total *int, fieldName string) *ValidationError {
	if input != nil && *input < 0 {
		return &ValidationError{Field: fieldName + ".input", Message: "input tokens cannot be negative"}
	}

	if output != nil && *output < 0 {
		return &ValidationError{Field: fieldName + ".output", Message: "output tokens cannot be negative"}
	}

	if total != nil && *total < 0 {
		return &ValidationError{Field: fieldName + ".total", Message: "total tokens cannot be negative"}
	}

	// Validate that total equals input + output if all are provided
	if input != nil && output != nil && total != nil {
		expectedTotal := *input + *output
		if *total != expectedTotal {
			return &ValidationError{Field: fieldName + ".total", Message: fmt.Sprintf("total (%d) must equal input (%d) + output (%d)", *total, *input, *output)}
		}
	}

	return nil
}

// ValidateCost validates cost values
func ValidateCost(inputCost, outputCost, totalCost *float64, fieldName string) *ValidationError {
	if inputCost != nil && *inputCost < 0 {
		return &ValidationError{Field: fieldName + ".inputCost", Message: "input cost cannot be negative"}
	}

	if outputCost != nil && *outputCost < 0 {
		return &ValidationError{Field: fieldName + ".outputCost", Message: "output cost cannot be negative"}
	}

	if totalCost != nil && *totalCost < 0 {
		return &ValidationError{Field: fieldName + ".totalCost", Message: "total cost cannot be negative"}
	}

	return nil
}

// ValidateObservationLevel validates observation level
func ValidateObservationLevel(level string, fieldName string) *ValidationError {
	if level == "" {
		return nil
	}

	allowedLevels := []string{"DEBUG", "DEFAULT", "WARNING", "ERROR"}
	return ValidateEnum(level, fieldName, allowedLevels)
}

// ValidateScoreValue validates score values based on data type
func ValidateScoreValue(value interface{}, dataType string, fieldName string) *ValidationError {
	if value == nil {
		return &ValidationError{Field: fieldName, Message: "score value cannot be nil"}
	}

	switch strings.ToUpper(dataType) {
	case "NUMERIC":
		_, err := ToFloat64(value)
		if err != nil {
			return &ValidationError{Field: fieldName, Message: "numeric score value must be a valid number"}
		}

	case "BOOLEAN":
		if _, ok := value.(bool); !ok {
			return &ValidationError{Field: fieldName, Message: "boolean score value must be true or false"}
		}

	case "CATEGORICAL":
		if _, ok := value.(string); !ok {
			return &ValidationError{Field: fieldName, Message: "categorical score value must be a string"}
		}

	default:
		return &ValidationError{Field: "dataType", Message: "must be NUMERIC, BOOLEAN, or CATEGORICAL"}
	}

	return nil
}

// ValidateStruct validates a struct using reflection and validation tags
func ValidateStruct(s interface{}) ValidationErrors {
	var errors ValidationErrors

	v := reflect.ValueOf(s)
	t := reflect.TypeOf(s)

	// Handle pointer to struct
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			errors.Add("struct", "cannot be nil")
			return errors
		}
		v = v.Elem()
		t = t.Elem()
	}

	if v.Kind() != reflect.Struct {
		errors.Add("input", "must be a struct")
		return errors
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		fieldName := fieldType.Name
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag != "" && jsonTag != "-" {
			// Use JSON tag name if available
			if comma := strings.Index(jsonTag, ","); comma != -1 {
				fieldName = jsonTag[:comma]
			} else {
				fieldName = jsonTag
			}
		}

		// Check validation tags
		validateTag := fieldType.Tag.Get("validate")
		if validateTag != "" {
			if err := validateFieldByTag(field.Interface(), fieldName, validateTag); err != nil {
				errors.AddError(*err)
			}
		}
	}

	return errors
}

// validateFieldByTag validates a field based on validation tag
func validateFieldByTag(value interface{}, fieldName, tag string) *ValidationError {
	rules := strings.Split(tag, ",")

	for _, rule := range rules {
		rule = strings.TrimSpace(rule)

		if rule == "required" {
			if err := ValidateRequired(value, fieldName); err != nil {
				return err
			}
		}

		// Add more validation rules as needed
		// e.g., "min:5", "max:100", "email", etc.
	}

	return nil
}
