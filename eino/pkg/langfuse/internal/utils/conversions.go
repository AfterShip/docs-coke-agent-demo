package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// TimeFormats contains common time formats for parsing
var TimeFormats = []string{
	time.RFC3339,
	time.RFC3339Nano,
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05.000Z",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02",
	time.Kitchen,
	time.Stamp,
	time.StampMilli,
	time.StampMicro,
	time.StampNano,
}

// ParseTime attempts to parse a time string using various formats
func ParseTime(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, fmt.Errorf("empty time string")
	}
	
	for _, format := range TimeFormats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}
	
	// Try parsing as Unix timestamp (seconds)
	if timestamp, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
		return time.Unix(timestamp, 0), nil
	}
	
	// Try parsing as Unix timestamp (milliseconds)
	if timestamp, err := strconv.ParseInt(timeStr, 10, 64); err == nil && timestamp > 1000000000000 {
		return time.UnixMilli(timestamp), nil
	}
	
	return time.Time{}, fmt.Errorf("unable to parse time: %s", timeStr)
}

// FormatTime formats time to RFC3339Nano format (Langfuse standard)
func FormatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

// FormatTimePointer formats time pointer to RFC3339Nano format
func FormatTimePointer(t *time.Time) *string {
	if t == nil {
		return nil
	}
	formatted := FormatTime(*t)
	return &formatted
}

// ParseTimePointer parses time string into time pointer
func ParseTimePointer(timeStr *string) (*time.Time, error) {
	if timeStr == nil || *timeStr == "" {
		return nil, nil
	}
	
	t, err := ParseTime(*timeStr)
	if err != nil {
		return nil, err
	}
	
	return &t, nil
}

// NowUTC returns current time in UTC
func NowUTC() time.Time {
	return time.Now().UTC()
}

// NowUTCPointer returns current time in UTC as pointer
func NowUTCPointer() *time.Time {
	t := NowUTC()
	return &t
}

// ToFloat64 converts various numeric types to float64
func ToFloat64(value interface{}) (float64, error) {
	if value == nil {
		return 0, fmt.Errorf("value is nil")
	}
	
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	case json.Number:
		return v.Float64()
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

// ToInt64 converts various numeric types to int64
func ToInt64(value interface{}) (int64, error) {
	if value == nil {
		return 0, fmt.Errorf("value is nil")
	}
	
	switch v := value.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case uint:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		if v > 9223372036854775807 { // max int64
			return 0, fmt.Errorf("uint64 value %d exceeds int64 range", v)
		}
		return int64(v), nil
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	case json.Number:
		return v.Int64()
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", value)
	}
}

// ToString converts various types to string
func ToString(value interface{}) string {
	if value == nil {
		return ""
	}
	
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case fmt.Stringer:
		return v.String()
	case bool:
		return strconv.FormatBool(v)
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case time.Time:
		return FormatTime(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ToStringPointer converts value to string pointer
func ToStringPointer(value interface{}) *string {
	if value == nil {
		return nil
	}
	
	str := ToString(value)
	if str == "" {
		return nil
	}
	
	return &str
}

// IsEmpty checks if a value is considered empty
func IsEmpty(value interface{}) bool {
	if value == nil {
		return true
	}
	
	v := reflect.ValueOf(value)
	
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Array, reflect.Slice, reflect.Map, reflect.Chan:
		return v.Len() == 0
	case reflect.Struct:
		// Check if it's time.Time and is zero
		if t, ok := value.(time.Time); ok {
			return t.IsZero()
		}
		return false
	default:
		return false
	}
}

// IsValidJSON checks if a string is valid JSON
func IsValidJSON(str string) bool {
	var js interface{}
	return json.Unmarshal([]byte(str), &js) == nil
}

// IsJSONSerializable checks if a value can be serialized to JSON
func IsJSONSerializable(value interface{}) bool {
	_, err := json.Marshal(value)
	return err == nil
}

// ToJSON converts value to JSON string
func ToJSON(value interface{}) (string, error) {
	bytes, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// FromJSON parses JSON string into value
func FromJSON(jsonStr string, target interface{}) error {
	return json.Unmarshal([]byte(jsonStr), target)
}

// ToJSONIndent converts value to indented JSON string
func ToJSONIndent(value interface{}, indent string) (string, error) {
	bytes, err := json.MarshalIndent(value, "", indent)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// MergeMetadata merges two metadata maps, with the second overriding the first
func MergeMetadata(base, override map[string]interface{}) map[string]interface{} {
	if base == nil && override == nil {
		return nil
	}
	
	result := make(map[string]interface{})
	
	// Copy base metadata
	for k, v := range base {
		result[k] = v
	}
	
	// Override with second metadata
	for k, v := range override {
		result[k] = v
	}
	
	return result
}

// CloneMetadata creates a deep copy of metadata map
func CloneMetadata(metadata map[string]interface{}) map[string]interface{} {
	if metadata == nil {
		return nil
	}
	
	// Use JSON marshaling/unmarshaling for deep copy
	bytes, err := json.Marshal(metadata)
	if err != nil {
		// Fallback to shallow copy if JSON marshaling fails
		result := make(map[string]interface{}, len(metadata))
		for k, v := range metadata {
			result[k] = v
		}
		return result
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(bytes, &result); err != nil {
		// Fallback to shallow copy if JSON unmarshaling fails
		result = make(map[string]interface{}, len(metadata))
		for k, v := range metadata {
			result[k] = v
		}
	}
	
	return result
}

// FilterMetadata filters metadata by keys
func FilterMetadata(metadata map[string]interface{}, keys []string) map[string]interface{} {
	if metadata == nil || len(keys) == 0 {
		return nil
	}
	
	keySet := make(map[string]bool, len(keys))
	for _, key := range keys {
		keySet[key] = true
	}
	
	result := make(map[string]interface{})
	for k, v := range metadata {
		if keySet[k] {
			result[k] = v
		}
	}
	
	if len(result) == 0 {
		return nil
	}
	
	return result
}

// SanitizeMetadata removes nil values and empty strings from metadata
func SanitizeMetadata(metadata map[string]interface{}) map[string]interface{} {
	if metadata == nil {
		return nil
	}
	
	result := make(map[string]interface{})
	for k, v := range metadata {
		if v == nil {
			continue
		}
		
		if str, ok := v.(string); ok && str == "" {
			continue
		}
		
		result[k] = v
	}
	
	if len(result) == 0 {
		return nil
	}
	
	return result
}

// ConvertToMap converts struct to map[string]interface{} using JSON tags
func ConvertToMap(value interface{}) (map[string]interface{}, error) {
	// Use JSON marshaling/unmarshaling to respect JSON tags
	bytes, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal value: %w", err)
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(bytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}
	
	return result, nil
}

// GetDurationBetween calculates duration between two times
func GetDurationBetween(start, end time.Time) time.Duration {
	return end.Sub(start)
}

// GetDurationFromNow calculates duration from now to given time
func GetDurationFromNow(t time.Time) time.Duration {
	return time.Until(t)
}

// GetDurationToNow calculates duration from given time to now
func GetDurationToNow(t time.Time) time.Duration {
	return time.Since(t)
}

// FormatDuration formats duration in human-readable format
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.2fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.2fm", d.Minutes())
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%.2fh", d.Hours())
	} else {
		days := d.Hours() / 24
		return fmt.Sprintf("%.2fd", days)
	}
}

// TruncateString truncates string to specified length with ellipsis
func TruncateString(str string, maxLen int) string {
	if len(str) <= maxLen {
		return str
	}
	
	if maxLen <= 3 {
		return str[:maxLen]
	}
	
	return str[:maxLen-3] + "..."
}

// CoalesceString returns the first non-empty string
func CoalesceString(strings ...string) string {
	for _, str := range strings {
		if str != "" {
			return str
		}
	}
	return ""
}

// CoalescePointer returns the first non-nil pointer
func CoalescePointer[T any](pointers ...*T) *T {
	for _, ptr := range pointers {
		if ptr != nil {
			return ptr
		}
	}
	return nil
}

// DerefPointer dereferences pointer with default value
func DerefPointer[T any](ptr *T, defaultValue T) T {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

// ToPointer converts value to pointer
func ToPointer[T any](value T) *T {
	return &value
}