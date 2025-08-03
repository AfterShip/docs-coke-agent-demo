package aftership_genkit

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/firebase/genkit/go/ai"
)

// Data extracts content type and data from a Part.
func Data(p *ai.Part) (contentType string, data []byte, err error) {
	if p.IsMedia() {
		// For media parts, the content is in the Text field as a data URI
		// or the ContentType and Text fields contain the type and base64 data
		contentType = p.ContentType
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		text := p.Text
		if strings.HasPrefix(text, "data:") {
			// Handle data URI format: "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQABAAD..."
			parts := strings.SplitN(text[5:], ",", 2)
			if len(parts) != 2 {
				return "", nil, fmt.Errorf("invalid data URI format")
			}

			// Parse media type and encoding
			mediaTypeParts := strings.Split(parts[0], ";")
			contentType = mediaTypeParts[0]

			// Check if base64 encoded
			isBase64 := false
			for _, part := range mediaTypeParts[1:] {
				if part == "base64" {
					isBase64 = true
					break
				}
			}

			if isBase64 {
				data, err = base64.StdEncoding.DecodeString(parts[1])
				if err != nil {
					return "", nil, fmt.Errorf("failed to decode base64 data: %w", err)
				}
			} else {
				data = []byte(parts[1])
			}

			return contentType, data, nil
		} else {
			// Assume the Text field contains base64 encoded data
			data, err = base64.StdEncoding.DecodeString(text)
			if err != nil {
				return "", nil, fmt.Errorf("failed to decode base64 data: %w", err)
			}
			return contentType, data, nil
		}
	}

	if p.IsData() {
		// For data parts, the content is in the Text field
		contentType = p.ContentType
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		text := p.Text
		if strings.HasPrefix(text, "data:") {
			// Handle data URI format
			parts := strings.SplitN(text[5:], ",", 2)
			if len(parts) != 2 {
				return "", nil, fmt.Errorf("invalid data URI format")
			}

			// Parse media type and encoding
			mediaTypeParts := strings.Split(parts[0], ";")
			contentType = mediaTypeParts[0]

			// Check if base64 encoded
			isBase64 := false
			for _, part := range mediaTypeParts[1:] {
				if part == "base64" {
					isBase64 = true
					break
				}
			}

			if isBase64 {
				data, err = base64.StdEncoding.DecodeString(parts[1])
				if err != nil {
					return "", nil, fmt.Errorf("failed to decode base64 data: %w", err)
				}
			} else {
				data = []byte(parts[1])
			}

			return contentType, data, nil
		} else {
			// Assume the Text field contains base64 encoded data
			data, err = base64.StdEncoding.DecodeString(text)
			if err != nil {
				return "", nil, fmt.Errorf("failed to decode base64 data: %w", err)
			}
			return contentType, data, nil
		}
	}

	return "", nil, fmt.Errorf("unsupported part type for data extraction")
}
