package middleware

import (
	"context"
	"net/http"

	"eino/pkg/langfuse/client"
)

// ContextPropagationConfig contains configuration for context propagation
type ContextPropagationConfig struct {
	// TraceIDHeader is the HTTP header used to propagate trace IDs
	// Default: "X-Trace-ID"
	TraceIDHeader string
	
	// SpanIDHeader is the HTTP header used to propagate span IDs
	// Default: "X-Span-ID"
	SpanIDHeader string
	
	// UserIDHeader is the HTTP header used to propagate user IDs
	// Default: "X-User-ID"
	UserIDHeader string
	
	// SessionIDHeader is the HTTP header used to propagate session IDs
	// Default: "X-Session-ID"
	SessionIDHeader string
}

// DefaultContextPropagationConfig returns default configuration for context propagation
func DefaultContextPropagationConfig() *ContextPropagationConfig {
	return &ContextPropagationConfig{
		TraceIDHeader:   "X-Trace-ID",
		SpanIDHeader:    "X-Span-ID",
		UserIDHeader:    "X-User-ID",
		SessionIDHeader: "X-Session-ID",
	}
}

// TraceContext holds trace context information for propagation
type TraceContext struct {
	TraceID   string
	SpanID    string
	UserID    string
	SessionID string
}

// ExtractTraceContext extracts trace context from HTTP headers
func ExtractTraceContext(headers http.Header, config *ContextPropagationConfig) *TraceContext {
	if config == nil {
		config = DefaultContextPropagationConfig()
	}
	
	return &TraceContext{
		TraceID:   headers.Get(config.TraceIDHeader),
		SpanID:    headers.Get(config.SpanIDHeader),
		UserID:    headers.Get(config.UserIDHeader),
		SessionID: headers.Get(config.SessionIDHeader),
	}
}

// InjectTraceContext injects trace context into HTTP headers
func InjectTraceContext(headers http.Header, traceCtx *TraceContext, config *ContextPropagationConfig) {
	if config == nil {
		config = DefaultContextPropagationConfig()
	}
	
	if traceCtx.TraceID != "" {
		headers.Set(config.TraceIDHeader, traceCtx.TraceID)
	}
	if traceCtx.SpanID != "" {
		headers.Set(config.SpanIDHeader, traceCtx.SpanID)
	}
	if traceCtx.UserID != "" {
		headers.Set(config.UserIDHeader, traceCtx.UserID)
	}
	if traceCtx.SessionID != "" {
		headers.Set(config.SessionIDHeader, traceCtx.SessionID)
	}
}

// PropagateTraceContext creates a new HTTP request with trace context propagated
func PropagateTraceContext(req *http.Request, config *ContextPropagationConfig) *http.Request {
	if config == nil {
		config = DefaultContextPropagationConfig()
	}
	
	ctx := req.Context()
	
	// Get trace and span from context
	if traceBuilder := GetTraceFromContext(ctx); traceBuilder != nil {
		// Extract trace ID if available (this would require extending TraceBuilder)
		// For now, we'll use what we can get from context
		req.Header.Set(config.TraceIDHeader, traceBuilder.GetID())
		
		if spanBuilder := GetSpanFromContext(ctx); spanBuilder != nil {
			req.Header.Set(config.SpanIDHeader, spanBuilder.GetID())
		}
		
		// Extract user and session IDs if available
		if userID := traceBuilder.GetUserID(); userID != "" {
			req.Header.Set(config.UserIDHeader, userID)
		}
		
		if sessionID := traceBuilder.GetSessionID(); sessionID != "" {
			req.Header.Set(config.SessionIDHeader, sessionID)
		}
	}
	
	return req
}

// HTTPClientWithTracing wraps an HTTP client to automatically propagate trace context
type HTTPClientWithTracing struct {
	*http.Client
	config *ContextPropagationConfig
}

// NewHTTPClientWithTracing creates a new HTTP client that automatically propagates trace context
func NewHTTPClientWithTracing(client *http.Client, config *ContextPropagationConfig) *HTTPClientWithTracing {
	if client == nil {
		client = http.DefaultClient
	}
	if config == nil {
		config = DefaultContextPropagationConfig()
	}
	
	return &HTTPClientWithTracing{
		Client: client,
		config: config,
	}
}

// Do executes an HTTP request with trace context propagation
func (c *HTTPClientWithTracing) Do(req *http.Request) (*http.Response, error) {
	// Propagate trace context
	req = PropagateTraceContext(req, c.config)
	
	// Execute the request
	return c.Client.Do(req)
}

// RoundTripper that automatically propagates trace context
type tracingRoundTripper struct {
	base   http.RoundTripper
	config *ContextPropagationConfig
}

// NewTracingRoundTripper creates a new RoundTripper that propagates trace context
func NewTracingRoundTripper(base http.RoundTripper, config *ContextPropagationConfig) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	if config == nil {
		config = DefaultContextPropagationConfig()
	}
	
	return &tracingRoundTripper{
		base:   base,
		config: config,
	}
}

func (t *tracingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Propagate trace context
	req = PropagateTraceContext(req, t.config)
	
	// Execute the request
	return t.base.RoundTrip(req)
}

// ContextWithTrace adds a trace builder to the context
func ContextWithTrace(ctx context.Context, trace *client.TraceBuilder) context.Context {
	return context.WithValue(ctx, traceBuilderContextKey, trace)
}

// ContextWithSpan adds a span builder to the context
func ContextWithSpan(ctx context.Context, span *client.SpanBuilder) context.Context {
	return context.WithValue(ctx, spanBuilderContextKey, span)
}

// ContextWithTraceAndSpan adds both trace and span builders to the context
func ContextWithTraceAndSpan(ctx context.Context, trace *client.TraceBuilder, span *client.SpanBuilder) context.Context {
	ctx = context.WithValue(ctx, traceBuilderContextKey, trace)
	ctx = context.WithValue(ctx, spanBuilderContextKey, span)
	return ctx
}

// StartSpanFromContext starts a new span from the trace in the context
func StartSpanFromContext(ctx context.Context, name string) (*client.SpanBuilder, context.Context) {
	if traceBuilder := GetTraceFromContext(ctx); traceBuilder != nil {
		spanBuilder := traceBuilder.Span(name)
		newCtx := context.WithValue(ctx, spanBuilderContextKey, spanBuilder)
		return spanBuilder, newCtx
	}
	return nil, ctx
}

// StartChildSpanFromContext starts a new child span from the current span in the context
func StartChildSpanFromContext(ctx context.Context, name string) (*client.SpanBuilder, context.Context) {
	if parentSpan := GetSpanFromContext(ctx); parentSpan != nil {
		// Create a child span (this would require extending SpanBuilder to support children)
		childSpan := parentSpan.ChildSpan(name)
		newCtx := context.WithValue(ctx, spanBuilderContextKey, childSpan)
		return childSpan, newCtx
	}
	
	// Fall back to creating a span from trace
	return StartSpanFromContext(ctx, name)
}

// FinishSpanInContext finishes the current span in the context
func FinishSpanInContext(ctx context.Context) error {
	if spanBuilder := GetSpanFromContext(ctx); spanBuilder != nil {
		return spanBuilder.End(ctx)
	}
	return nil
}

// AddEventToSpanInContext adds an event to the current span in the context
func AddEventToSpanInContext(ctx context.Context, name string, attributes map[string]interface{}) {
	if spanBuilder := GetSpanFromContext(ctx); spanBuilder != nil {
		// This would require extending SpanBuilder to support events
		// For now, we could add this to metadata
		if attributes == nil {
			attributes = make(map[string]interface{})
		}
		attributes["event_name"] = name
		spanBuilder.WithMetadata(attributes)
	}
}

// SetSpanErrorInContext marks the current span as having an error
func SetSpanErrorInContext(ctx context.Context, err error) {
	if spanBuilder := GetSpanFromContext(ctx); spanBuilder != nil {
		spanBuilder.WithLevel("ERROR").WithStatusMessage(err.Error())
	}
}

// SetSpanAttributesInContext adds attributes to the current span in the context
func SetSpanAttributesInContext(ctx context.Context, attributes map[string]interface{}) {
	if spanBuilder := GetSpanFromContext(ctx); spanBuilder != nil {
		// Merge with existing metadata
		spanBuilder.WithMetadata(attributes)
	}
}

// GetTraceIDFromContext extracts the trace ID from the context
func GetTraceIDFromContext(ctx context.Context) string {
	if traceBuilder := GetTraceFromContext(ctx); traceBuilder != nil {
		return traceBuilder.GetID()
	}
	return ""
}

// GetSpanIDFromContext extracts the span ID from the context
func GetSpanIDFromContext(ctx context.Context) string {
	if spanBuilder := GetSpanFromContext(ctx); spanBuilder != nil {
		return spanBuilder.GetID()
	}
	return ""
}

// GetUserIDFromContext extracts the user ID from the trace in the context
func GetUserIDFromContext(ctx context.Context) string {
	if traceBuilder := GetTraceFromContext(ctx); traceBuilder != nil {
		return traceBuilder.GetUserID()
	}
	return ""
}

// GetSessionIDFromContext extracts the session ID from the trace in the context
func GetSessionIDFromContext(ctx context.Context) string {
	if traceBuilder := GetTraceFromContext(ctx); traceBuilder != nil {
		return traceBuilder.GetSessionID()
	}
	return ""
}

// TraceMiddleware is a simplified HTTP middleware that only adds trace context
func TraceMiddleware(langfuseClient *client.Langfuse) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create or extract trace context
			var traceBuilder *client.TraceBuilder
			
			// Check if trace context is already in headers (propagated from upstream)
			config := DefaultContextPropagationConfig()
			traceCtx := ExtractTraceContext(r.Header, config)
			
			if traceCtx.TraceID != "" {
				// Continue existing trace (this would require extending the API)
				traceBuilder = langfuseClient.Trace(r.Method + " " + r.URL.Path)
				// Would need to set the trace ID to continue the existing trace
			} else {
				// Start new trace
				traceBuilder = langfuseClient.Trace(r.Method + " " + r.URL.Path)
			}
			
			// Add trace to context
			ctx := ContextWithTrace(r.Context(), traceBuilder)
			r = r.WithContext(ctx)
			
			// Add trace ID to response headers for downstream services
			w.Header().Set(config.TraceIDHeader, traceBuilder.GetID())
			
			// Execute next handler
			next.ServeHTTP(w, r)
		})
	}
}