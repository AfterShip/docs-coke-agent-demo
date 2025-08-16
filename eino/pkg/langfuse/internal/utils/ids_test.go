package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateNanoid(t *testing.T) {
	id := GenerateNanoid()
	
	assert.Len(t, id, DefaultNanoidLength)
	assert.True(t, IsValidNanoid(id))
	
	// Generate multiple IDs to ensure uniqueness
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		newID := GenerateNanoid()
		assert.False(t, ids[newID], "Generated duplicate ID: %s", newID)
		ids[newID] = true
	}
}

func TestGenerateNanoidWithLength(t *testing.T) {
	tests := []int{5, 10, 16, 21, 32, 50}
	
	for _, length := range tests {
		t.Run(fmt.Sprintf("length_%d", length), func(t *testing.T) {
			id := GenerateNanoidWithLength(length)
			assert.Len(t, id, length)
			assert.True(t, IsValidNanoid(id))
		})
	}
	
	// Test edge cases
	t.Run("zero_length", func(t *testing.T) {
		id := GenerateNanoidWithLength(0)
		assert.Len(t, id, DefaultNanoidLength) // Should default
	})
	
	t.Run("negative_length", func(t *testing.T) {
		id := GenerateNanoidWithLength(-5)
		assert.Len(t, id, DefaultNanoidLength) // Should default
	})
}

func TestGenerateNanoidWithOptions(t *testing.T) {
	tests := []struct {
		name     string
		alphabet string
		length   int
		wantLen  int
	}{
		{"custom alphabet", "ABC123", 10, 10},
		{"numeric only", "0123456789", 8, 8},
		{"empty alphabet", "", 10, 10}, // Should use provided length
		{"zero length", DefaultAlphabet, 0, DefaultNanoidLength}, // Should use default
		{"very short alphabet", "AB", 20, 20},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := GenerateNanoidWithOptions(tt.alphabet, tt.length)
			
			if tt.alphabet == "" {
				assert.Len(t, id, tt.wantLen)
			} else if tt.alphabet == "AB" && tt.length == 20 {
				// Special case: very short alphabet
				assert.Len(t, id, tt.wantLen)
				for _, char := range id {
					assert.Contains(t, tt.alphabet, string(char))
				}
			} else {
				assert.Len(t, id, tt.wantLen)
				
				// Verify characters are from the alphabet
				expectedAlphabet := tt.alphabet
				if expectedAlphabet == "" {
					expectedAlphabet = DefaultAlphabet
				}
				for _, char := range id {
					assert.Contains(t, expectedAlphabet, string(char))
				}
			}
		})
	}
}

func TestGenerateSafeNanoid(t *testing.T) {
	id := GenerateSafeNanoid()
	
	assert.Len(t, id, DefaultNanoidLength)
	
	// Verify all characters are from safe alphabet
	for _, char := range id {
		assert.Contains(t, SafeAlphabet, string(char))
	}
}

func TestGenerateUUID(t *testing.T) {
	id := GenerateUUID()
	
	// Should be a valid UUID format
	assert.True(t, IsValidUUID(id))
	
	// Should contain hyphens
	assert.Contains(t, id, "-")
	
	// Generate multiple UUIDs to ensure uniqueness
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		newID := GenerateUUID()
		assert.False(t, ids[newID], "Generated duplicate UUID: %s", newID)
		ids[newID] = true
	}
}

func TestGenerateUUIDWithoutHyphens(t *testing.T) {
	id := GenerateUUIDWithoutHyphens()
	
	// Should not contain hyphens
	assert.NotContains(t, id, "-")
	
	// Should be 32 characters (UUID without hyphens)
	assert.Len(t, id, 32)
	
	// Should be hexadecimal
	hexRegex := regexp.MustCompile("^[0-9a-f]{32}$")
	assert.True(t, hexRegex.MatchString(id))
}

func TestSpecificIDGenerators(t *testing.T) {
	tests := []struct {
		name      string
		generator func() string
		wantLen   int
	}{
		{"GenerateTraceID", GenerateTraceID, 16},
		{"GenerateObservationID", GenerateObservationID, 16},
		{"GenerateScoreID", GenerateScoreID, 12},
		{"GenerateSessionID", GenerateSessionID, 16},
		{"GenerateEventID", GenerateEventID, 16},
		{"GenerateRequestID", GenerateRequestID, 12},
		{"GenerateShortID", GenerateShortID, 8},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := tt.generator()
			assert.Len(t, id, tt.wantLen)
			assert.True(t, IsValidNanoid(id))
		})
	}
}

func TestGenerateCorrelationID(t *testing.T) {
	id := GenerateCorrelationID()
	assert.True(t, IsValidUUID(id))
}

func TestGenerateAPIKey(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		want   string
	}{
		{"default prefix", "", "lf_"},
		{"custom prefix", "pk", "pk_"},
		{"empty custom prefix", "", "lf_"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := GenerateAPIKey(tt.prefix)
			
			assert.True(t, strings.HasPrefix(key, tt.want))
			
			// Extract the ID part and verify it's a valid nanoid
			parts := strings.Split(key, "_")
			require.Len(t, parts, 2)
			assert.Len(t, parts[1], 32) // API keys use 32 character nanoids
			assert.True(t, IsValidNanoid(parts[1]))
		})
	}
}

func TestGenerateSecretKey(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		want   string
	}{
		{"default prefix", "", "sk_"},
		{"custom prefix", "secret", "secret_"},
		{"empty custom prefix", "", "sk_"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := GenerateSecretKey(tt.prefix)
			
			assert.True(t, strings.HasPrefix(key, tt.want))
			
			// Extract the ID part and verify it's a valid nanoid
			parts := strings.Split(key, "_")
			require.Len(t, parts, 2)
			assert.Len(t, parts[1], 48) // Secret keys use 48 character nanoids
			assert.True(t, IsValidNanoid(parts[1]))
		})
	}
}

func TestIsValidNanoid(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want bool
	}{
		{"empty string", "", false},
		{"valid nanoid", GenerateNanoid(), true},
		{"valid short nanoid", GenerateNanoidWithLength(5), true},
		{"invalid characters", "hello@world", false},
		{"valid with mixed case", "AbC123xyz", true},
		{"only numbers", "123456789", true},
		{"only letters", "abcdefghi", true},
		{"special characters not in alphabet", "test-id_with*special", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidNanoid(tt.id)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestIsValidUUID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want bool
	}{
		{"valid UUID v4", GenerateUUID(), true},
		{"valid UUID manual", "550e8400-e29b-41d4-a716-446655440000", true},
		{"invalid format", "not-a-uuid", false},
		{"invalid length", "550e8400-e29b-41d4-a716", false},
		{"invalid characters", "550e8400-e29b-41d4-a716-44665544000g", false},
		{"empty string", "", false},
		{"UUID without hyphens", GenerateUUIDWithoutHyphens(), false}, // IsValidUUID expects hyphens
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidUUID(tt.id)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestIsValidID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want bool
	}{
		{"valid nanoid", GenerateNanoid(), true},
		{"valid UUID", GenerateUUID(), true},
		{"invalid format", "not-valid", false},
		{"empty string", "", false},
		{"valid UUID without hyphens", GenerateUUIDWithoutHyphens(), false}, // Should be false since IsValidUUID expects hyphens
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidID(tt.id)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestGenerateIDWithTimestamp(t *testing.T) {
	before := time.Now()
	id := GenerateIDWithTimestamp()
	after := time.Now()
	
	parts := strings.Split(id, "_")
	require.Len(t, parts, 2)
	
	// Verify timestamp is reasonable
	timestamp, err := strconv.ParseInt(parts[0], 10, 64)
	require.NoError(t, err)
	
	timestampTime := time.UnixMilli(timestamp)
	assert.True(t, timestampTime.After(before.Add(-time.Second)))
	assert.True(t, timestampTime.Before(after.Add(time.Second)))
	
	// Verify nanoid part
	assert.Len(t, parts[1], 10)
	assert.True(t, IsValidNanoid(parts[1]))
}

func TestGenerateSortableID(t *testing.T) {
	// Generate multiple IDs to test sorting
	var ids []string
	for i := 0; i < 10; i++ {
		ids = append(ids, GenerateSortableID())
		time.Sleep(time.Millisecond) // Ensure different timestamps
	}
	
	// Verify IDs are in ascending order (since they're timestamp-based)
	for i := 1; i < len(ids); i++ {
		assert.True(t, ids[i-1] < ids[i], "IDs should be sortable: %s should be < %s", ids[i-1], ids[i])
	}
	
	// Verify format
	for _, id := range ids {
		parts := strings.Split(id, "_")
		require.Len(t, parts, 2, "ID should have format timestamp_nanoid: %s", id)
		assert.True(t, len(parts[0]) > 0, "Timestamp part should not be empty")
		assert.Len(t, parts[1], 8, "Nanoid part should be 8 characters")
		assert.True(t, IsValidNanoid(parts[1]))
	}
}

func TestExtractTimestampFromSortableID(t *testing.T) {
	// Generate a sortable ID and extract timestamp
	before := time.Now()
	id := GenerateSortableID()
	after := time.Now()
	
	timestamp, err := ExtractTimestampFromSortableID(id)
	require.NoError(t, err)
	require.NotNil(t, timestamp)
	
	// Verify timestamp is within expected range
	assert.True(t, timestamp.After(before.Add(-time.Second)))
	assert.True(t, timestamp.Before(after.Add(time.Second)))
	
	// Test invalid formats
	tests := []struct {
		name string
		id   string
	}{
		{"no underscore", "invalidid"},
		{"multiple underscores", "part1_part2_part3"},
		{"invalid timestamp", "invalid_nanoid123"},
		{"empty string", ""},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timestamp, err := ExtractTimestampFromSortableID(tt.id)
			assert.Error(t, err)
			assert.Nil(t, timestamp)
		})
	}
}

func TestGenerateRandomString(t *testing.T) {
	tests := []int{1, 5, 10, 50, 100}
	
	for _, length := range tests {
		t.Run(fmt.Sprintf("length_%d", length), func(t *testing.T) {
			str := GenerateRandomString(length)
			assert.Len(t, str, length)
			
			// Verify all characters are from default alphabet
			for _, char := range str {
				assert.Contains(t, DefaultAlphabet, string(char))
			}
		})
	}
}

func TestGenerateNumericID(t *testing.T) {
	tests := []int{5, 10, 20}
	
	for _, length := range tests {
		t.Run(fmt.Sprintf("length_%d", length), func(t *testing.T) {
			id := GenerateNumericID(length)
			assert.Len(t, id, length)
			
			// Verify all characters are numeric
			numericRegex := regexp.MustCompile("^[0-9]+$")
			assert.True(t, numericRegex.MatchString(id))
		})
	}
}

func TestGenerateAlphaID(t *testing.T) {
	tests := []int{5, 10, 20}
	
	for _, length := range tests {
		t.Run(fmt.Sprintf("length_%d", length), func(t *testing.T) {
			id := GenerateAlphaID(length)
			assert.Len(t, id, length)
			
			// Verify all characters are alphabetic
			alphaRegex := regexp.MustCompile("^[a-zA-Z]+$")
			assert.True(t, alphaRegex.MatchString(id))
		})
	}
}

func TestAlphabetConstants(t *testing.T) {
	// Test DefaultAlphabet
	assert.Equal(t, "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz", DefaultAlphabet)
	assert.Len(t, DefaultAlphabet, 62)
	
	// Test SafeAlphabet
	assert.Equal(t, "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz", SafeAlphabet)
	assert.Len(t, SafeAlphabet, 63)
	
	// Test that SafeAlphabet includes underscore
	assert.Contains(t, SafeAlphabet, "_")
	
	// Test that DefaultAlphabet doesn't include underscore
	assert.NotContains(t, DefaultAlphabet, "_")
}

func TestDefaultConstants(t *testing.T) {
	assert.Equal(t, 21, DefaultNanoidLength)
}

func TestConcurrentGeneration(t *testing.T) {
	const numGoroutines = 100
	const idsPerGoroutine = 10
	
	idCh := make(chan string, numGoroutines*idsPerGoroutine)
	
	// Generate IDs concurrently
	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < idsPerGoroutine; j++ {
				idCh <- GenerateNanoid()
			}
		}()
	}
	
	// Collect all IDs
	ids := make(map[string]bool)
	for i := 0; i < numGoroutines*idsPerGoroutine; i++ {
		id := <-idCh
		assert.False(t, ids[id], "Generated duplicate ID in concurrent test: %s", id)
		ids[id] = true
		assert.True(t, IsValidNanoid(id))
	}
}

func BenchmarkGenerateNanoid(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateNanoid()
	}
}

func BenchmarkGenerateUUID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateUUID()
	}
}

func BenchmarkIsValidNanoid(b *testing.B) {
	id := GenerateNanoid()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		IsValidNanoid(id)
	}
}

func BenchmarkIsValidUUID(b *testing.B) {
	id := GenerateUUID()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		IsValidUUID(id)
	}
}