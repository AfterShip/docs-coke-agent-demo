package utils

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	// DefaultNanoidLength is the default length for nanoid generation
	DefaultNanoidLength = 21
	
	// DefaultAlphabet is the default alphabet for nanoid generation
	DefaultAlphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	
	// SafeAlphabet is a URL-safe alphabet for nanoid generation
	SafeAlphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz"
)

// GenerateNanoid generates a nanoid with the default alphabet and length
func GenerateNanoid() string {
	return GenerateNanoidWithOptions(DefaultAlphabet, DefaultNanoidLength)
}

// GenerateNanoidWithLength generates a nanoid with the default alphabet and specified length
func GenerateNanoidWithLength(length int) string {
	return GenerateNanoidWithOptions(DefaultAlphabet, length)
}

// GenerateNanoidWithOptions generates a nanoid with custom alphabet and length
func GenerateNanoidWithOptions(alphabet string, length int) string {
	if length <= 0 {
		length = DefaultNanoidLength
	}
	
	if alphabet == "" {
		alphabet = DefaultAlphabet
	}
	
	alphabetLen := len(alphabet)
	if alphabetLen == 0 {
		return ""
	}
	
	// Calculate the mask for efficient random generation
	mask := (2 << int(math.Log2(float64(alphabetLen-1)))) - 1
	step := int(math.Ceil(1.6 * float64(mask*length) / float64(alphabetLen)))
	
	result := make([]byte, length)
	for i, j := 0, 0; i < length; {
		bytes := make([]byte, step)
		_, err := rand.Read(bytes)
		if err != nil {
			// Fallback to less secure but working method
			return generateFallbackNanoid(alphabet, length)
		}
		
		for ; j < step && i < length; j++ {
			if int(bytes[j])&mask < alphabetLen {
				result[i] = alphabet[int(bytes[j])&mask]
				i++
			}
		}
		j = 0
	}
	
	return string(result)
}

// generateFallbackNanoid generates a nanoid using math/big for fallback
func generateFallbackNanoid(alphabet string, length int) string {
	alphabetLen := big.NewInt(int64(len(alphabet)))
	result := make([]byte, length)
	
	for i := 0; i < length; i++ {
		randomIndex, err := rand.Int(rand.Reader, alphabetLen)
		if err != nil {
			// Ultimate fallback using current time
			result[i] = alphabet[int(time.Now().UnixNano())%len(alphabet)]
		} else {
			result[i] = alphabet[randomIndex.Int64()]
		}
	}
	
	return string(result)
}

// GenerateSafeNanoid generates a URL-safe nanoid
func GenerateSafeNanoid() string {
	return GenerateNanoidWithOptions(SafeAlphabet, DefaultNanoidLength)
}

// GenerateUUID generates a UUID v4
func GenerateUUID() string {
	return uuid.New().String()
}

// GenerateUUIDWithoutHyphens generates a UUID v4 without hyphens
func GenerateUUIDWithoutHyphens() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

// GenerateTraceID generates a trace ID using nanoid
func GenerateTraceID() string {
	return GenerateNanoidWithLength(16)
}

// GenerateObservationID generates an observation ID using nanoid
func GenerateObservationID() string {
	return GenerateNanoidWithLength(16)
}

// GenerateScoreID generates a score ID using nanoid
func GenerateScoreID() string {
	return GenerateNanoidWithLength(12)
}

// GenerateSessionID generates a session ID using nanoid
func GenerateSessionID() string {
	return GenerateNanoidWithLength(16)
}

// GenerateEventID generates a generic event ID
func GenerateEventID() string {
	return GenerateNanoidWithLength(16)
}

// GenerateRequestID generates a request ID for logging and tracing
func GenerateRequestID() string {
	return GenerateNanoidWithLength(12)
}

// GenerateCorrelationID generates a correlation ID for distributed tracing
func GenerateCorrelationID() string {
	return GenerateUUID()
}

// GenerateAPIKey generates an API key-like string
func GenerateAPIKey(prefix string) string {
	if prefix == "" {
		prefix = "lf"
	}
	return fmt.Sprintf("%s_%s", prefix, GenerateNanoidWithLength(32))
}

// GenerateSecretKey generates a secret key-like string
func GenerateSecretKey(prefix string) string {
	if prefix == "" {
		prefix = "sk"
	}
	return fmt.Sprintf("%s_%s", prefix, GenerateNanoidWithLength(48))
}

// IsValidNanoid validates if a string is a valid nanoid format
func IsValidNanoid(id string) bool {
	if len(id) == 0 {
		return false
	}
	
	// Check if all characters are from the default alphabet
	for _, char := range id {
		if !strings.ContainsRune(DefaultAlphabet, char) {
			return false
		}
	}
	
	return true
}

// IsValidUUID validates if a string is a valid UUID format
func IsValidUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

// IsValidID validates if a string is a valid ID (either nanoid or UUID)
func IsValidID(id string) bool {
	return IsValidNanoid(id) || IsValidUUID(id)
}

// GenerateIDWithTimestamp generates an ID with a timestamp prefix
func GenerateIDWithTimestamp() string {
	timestamp := time.Now().UnixMilli()
	nanoid := GenerateNanoidWithLength(10)
	return fmt.Sprintf("%d_%s", timestamp, nanoid)
}

// GenerateSortableID generates a sortable ID based on timestamp
func GenerateSortableID() string {
	// Use base36 encoding for timestamp to make it shorter
	timestamp := time.Now().UnixMilli()
	timestampStr := fmt.Sprintf("%s", big.NewInt(timestamp).Text(36))
	nanoid := GenerateNanoidWithLength(8)
	return fmt.Sprintf("%s_%s", timestampStr, nanoid)
}

// ExtractTimestampFromSortableID extracts the timestamp from a sortable ID
func ExtractTimestampFromSortableID(id string) (*time.Time, error) {
	parts := strings.Split(id, "_")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid sortable ID format: %s", id)
	}
	
	timestampBig, ok := big.NewInt(0).SetString(parts[0], 36)
	if !ok {
		return nil, fmt.Errorf("invalid timestamp in sortable ID: %s", parts[0])
	}
	
	timestamp := time.UnixMilli(timestampBig.Int64())
	return &timestamp, nil
}

// GenerateShortID generates a short ID (8 characters)
func GenerateShortID() string {
	return GenerateNanoidWithLength(8)
}

// GenerateRandomString generates a random string of specified length using default alphabet
func GenerateRandomString(length int) string {
	return GenerateNanoidWithOptions(DefaultAlphabet, length)
}

// GenerateNumericID generates a numeric-only ID
func GenerateNumericID(length int) string {
	numericAlphabet := "0123456789"
	return GenerateNanoidWithOptions(numericAlphabet, length)
}

// GenerateAlphaID generates an alphabetic-only ID
func GenerateAlphaID(length int) string {
	alphaAlphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	return GenerateNanoidWithOptions(alphaAlphabet, length)
}