package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"eino/pkg/langfuse/client"
)

// HTTPMiddlewareConfig contains configuration options for the HTTP middleware
type HTTPMiddlewareConfig struct {
	// Client is the Langfuse client instance to use for tracing
	Client *client.Langfuse
	
	// TraceNameFunc allows customizing trace names based on the request
	// Default: uses HTTP method and path
	TraceNameFunc func(*http.Request) string
	
	// ShouldTraceFunc determines whether a request should be traced
	// Default: traces all requests
	ShouldTraceFunc func(*http.Request) bool
	
	// UserIDExtractor extracts user ID from the request context or headers
	// Default: looks for "X-User-ID" header
	UserIDExtractor func(*http.Request) string
	
	// SessionIDExtractor extracts session ID from the request context or headers
	// Default: looks for "X-Session-ID" header
	SessionIDExtractor func(*http.Request) string
	
	// MetadataExtractor extracts additional metadata from the request
	// Default: includes method, path, user-agent, and remote addr
	MetadataExtractor func(*http.Request) map[string]interface{}
	
	// TagExtractor extracts tags from the request
	// Default: includes HTTP method
	TagExtractor func(*http.Request) []string
	
	// CaptureRequestBody determines whether to capture request body as input
	// Default: false (can be memory intensive)
	CaptureRequestBody bool
	
	// CaptureResponseBody determines whether to capture response body as output
	// Default: false (can be memory intensive)
	CaptureResponseBody bool
	
	// MaxBodySize limits the size of captured request/response bodies
	// Default: 64KB
	MaxBodySize int64
}

// DefaultHTTPMiddlewareConfig returns a default configuration
func DefaultHTTPMiddlewareConfig(langfuseClient *client.Langfuse) *HTTPMiddlewareConfig {
	return &HTTPMiddlewareConfig{
		Client:              langfuseClient,
		TraceNameFunc:       defaultTraceNameFunc,
		ShouldTraceFunc:     defaultShouldTraceFunc,
		UserIDExtractor:     defaultUserIDExtractor,
		SessionIDExtractor:  defaultSessionIDExtractor,
		MetadataExtractor:   defaultMetadataExtractor,
		TagExtractor:        defaultTagExtractor,
		CaptureRequestBody:  false,
		CaptureResponseBody: false,
		MaxBodySize:         64 * 1024, // 64KB
	}
}

// contextKey is used for storing trace information in request context
type contextKey string

const (
	traceBuilderContextKey contextKey = "langfuse_trace_builder"
	spanBuilderContextKey  contextKey = "langfuse_span_builder"
)

// responseWriter wraps http.ResponseWriter to capture response details
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int64
	body       []byte
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriter) Write(data []byte) (int, error) {
	n, err := w.ResponseWriter.Write(data)
	w.size += int64(n)
	
	// Capture response body if configured
	if cap(w.body) > 0 {
		w.body = append(w.body, data...)
	}
	
	return n, err
}

// HTTPMiddleware creates HTTP middleware that automatically traces requests
func HTTPMiddleware(config *HTTPMiddlewareConfig) func(http.Handler) http.Handler {
	if config == nil {
		panic("HTTPMiddleware config cannot be nil")
	}
	if config.Client == nil {
		panic("HTTPMiddleware config must have a Langfuse client")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if this request should be traced
			if !config.ShouldTraceFunc(r) {
				next.ServeHTTP(w, r)
				return
			}

			// Start timing
			startTime := time.Now()

			// Create trace
			traceName := config.TraceNameFunc(r)
			traceBuilder := config.Client.Trace(traceName)

			// Set trace metadata
			if metadata := config.MetadataExtractor(r); metadata != nil {
				traceBuilder.WithMetadata(metadata)
			}

			// Set trace tags
			if tags := config.TagExtractor(r); tags != nil {
				traceBuilder.WithTags(tags...)
			}

			// Set user ID if available
			if userID := config.UserIDExtractor(r); userID != "" {
				traceBuilder.WithUser(userID)
			}

			// Set session ID if available
			if sessionID := config.SessionIDExtractor(r); sessionID != "" {
				traceBuilder.WithSession(sessionID)
			}

			// Capture request input if configured
			if config.CaptureRequestBody && r.Body != nil {
				if bodyData := captureRequestBody(r, config.MaxBodySize); bodyData != nil {
					traceBuilder.WithInput(bodyData)
				}
			} else {
				// At minimum, capture basic request info
				traceBuilder.WithInput(map[string]interface{}{
					"method": r.Method,
					"url":    r.URL.String(),
					"headers": filterHeaders(r.Header),
				})
			}

			// Create span for the HTTP operation
			spanBuilder := traceBuilder.Span("http_request").
				WithStartTime(startTime)

			// Add trace and span to context
			ctx := context.WithValue(r.Context(), traceBuilderContextKey, traceBuilder)
			ctx = context.WithValue(ctx, spanBuilderContextKey, spanBuilder)
			r = r.WithContext(ctx)

			// Wrap response writer to capture response details
			wrappedWriter := &responseWriter{
				ResponseWriter: w,
				statusCode:     200, // default status code
			}

			// Prepare to capture response body if configured
			if config.CaptureResponseBody {
				wrappedWriter.body = make([]byte, 0, config.MaxBodySize)
			}

			// Execute the handler
			next.ServeHTTP(wrappedWriter, r)

			// Complete timing
			endTime := time.Now()
			duration := endTime.Sub(startTime)

			// Complete the span with response details
			spanBuilder.
				WithEndTime(endTime).
				WithOutput(map[string]interface{}{
					"status_code": wrappedWriter.statusCode,
					"size":        wrappedWriter.size,
					"duration_ms": float64(duration.Nanoseconds()) / 1e6,
				})

			// Set span level based on status code
			if wrappedWriter.statusCode >= 400 {
				if wrappedWriter.statusCode >= 500 {
					spanBuilder.WithLevel("ERROR")
				} else {
					spanBuilder.WithLevel("WARNING")
				}
				
				spanBuilder.WithStatusMessage(http.StatusText(wrappedWriter.statusCode))
			}

			// Complete span
			if err := spanBuilder.End(r.Context()); err != nil {
				// Log error but don't fail the request
				// In production, you might want to use a proper logger here
			}

			// Set trace output if response body was captured
			if config.CaptureResponseBody && len(wrappedWriter.body) > 0 {
				traceBuilder.WithOutput(map[string]interface{}{
					"status_code": wrappedWriter.statusCode,
					"body":        string(wrappedWriter.body),
					"size":        wrappedWriter.size,
					"duration_ms": float64(duration.Nanoseconds()) / 1e6,
				})
			} else {
				traceBuilder.WithOutput(map[string]interface{}{
					"status_code": wrappedWriter.statusCode,
					"size":        wrappedWriter.size,
					"duration_ms": float64(duration.Nanoseconds()) / 1e6,
				})
			}

			// Complete the trace
			if err := traceBuilder.End(r.Context()); err != nil {
				// Log error but don't fail the request
				// In production, you might want to use a proper logger here
			}
		})
	}
}

// GetTraceFromContext retrieves the current trace builder from request context
func GetTraceFromContext(ctx context.Context) *client.TraceBuilder {
	if trace, ok := ctx.Value(traceBuilderContextKey).(*client.TraceBuilder); ok {
		return trace
	}
	return nil
}

// GetSpanFromContext retrieves the current span builder from request context
func GetSpanFromContext(ctx context.Context) *client.SpanBuilder {
	if span, ok := ctx.Value(spanBuilderContextKey).(*client.SpanBuilder); ok {
		return span
	}
	return nil
}

// Default extractor functions

func defaultTraceNameFunc(r *http.Request) string {
	return r.Method + " " + r.URL.Path
}

func defaultShouldTraceFunc(r *http.Request) bool {
	// Skip common health check endpoints
	path := r.URL.Path
	return path != "/health" && path != "/healthz" && path != "/ping"
}

func defaultUserIDExtractor(r *http.Request) string {
	// Try header first
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		return userID
	}
	
	// Try Authorization header (extract from JWT or Basic auth)
	// This is a simplified example - in practice you'd parse the token properly
	if auth := r.Header.Get("Authorization"); auth != "" {
		// You could add JWT parsing logic here
		return ""
	}
	
	return ""
}

func defaultSessionIDExtractor(r *http.Request) string {
	// Try header first
	if sessionID := r.Header.Get("X-Session-ID"); sessionID != "" {
		return sessionID
	}
	
	// Try cookie
	if cookie, err := r.Cookie("session_id"); err == nil {
		return cookie.Value
	}
	
	return ""
}

func defaultMetadataExtractor(r *http.Request) map[string]interface{} {
	metadata := map[string]interface{}{
		"method":     r.Method,
		"path":       r.URL.Path,
		"user_agent": r.Header.Get("User-Agent"),
		"remote_addr": r.RemoteAddr,
		"host":       r.Host,
	}
	
	// Add query parameters if present
	if len(r.URL.RawQuery) > 0 {
		metadata["query"] = r.URL.RawQuery
	}
	
	// Add content type
	if contentType := r.Header.Get("Content-Type"); contentType != "" {
		metadata["content_type"] = contentType
	}
	
	// Add content length
	if r.ContentLength > 0 {
		metadata["content_length"] = r.ContentLength
	}
	
	return metadata
}

func defaultTagExtractor(r *http.Request) []string {
	tags := []string{
		"http",
		r.Method,
	}
	
	// Add status-based tags after response (this is a limitation of the current design)
	// In a more sophisticated implementation, you might want to add these in a callback
	
	return tags
}

// Helper functions

func captureRequestBody(r *http.Request, maxSize int64) interface{} {
	if r.Body == nil {
		return nil
	}
	
	// This is a simplified implementation
	// In production, you'd want to:
	// 1. Read the body carefully to not consume it
	// 2. Restore the body for the actual handler
	// 3. Handle different content types appropriately
	// 4. Implement proper size limits
	
	// For now, return request metadata instead of actual body
	return map[string]interface{}{
		"method":         r.Method,
		"url":           r.URL.String(),
		"content_length": r.ContentLength,
		"content_type":   r.Header.Get("Content-Type"),
	}
}

func filterHeaders(headers http.Header) map[string]interface{} {
	// Filter out sensitive headers
	sensitiveHeaders := map[string]bool{
		"authorization": true,
		"cookie":        true,
		"x-api-key":     true,
		"x-auth-token":  true,
	}
	
	filtered := make(map[string]interface{})
	for name, values := range headers {
		lowername := strings.ToLower(name)
		if !sensitiveHeaders[lowername] {
			if len(values) == 1 {
				filtered[name] = values[0]
			} else {
				filtered[name] = values
			}
		}
	}
	
	return filtered
}