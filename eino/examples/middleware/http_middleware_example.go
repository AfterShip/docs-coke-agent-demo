package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"eino/pkg/langfuse/client"
	"eino/pkg/langfuse/middleware"
)

// Example HTTP handler
type UserHandler struct {
	langfuse *client.Langfuse
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	// Extract trace from middleware context
	trace := middleware.TraceFromContext(r.Context())
	if trace != nil {
		// Add additional metadata to the trace from middleware
		trace.WithMetadata(map[string]interface{}{
			"handler":    "GetUser",
			"user_agent": r.UserAgent(),
			"ip_address": r.RemoteAddr,
		})
	}

	// Simulate user lookup
	userID := r.URL.Query().Get("id")
	if userID == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	// Create a span for database operation
	if trace != nil {
		dbSpan := trace.Span("database-lookup").
			WithInput(map[string]interface{}{
				"query": "SELECT * FROM users WHERE id = ?",
				"params": []string{userID},
			}).
			WithStartTime(time.Now()).
			WithMetadata(map[string]interface{}{
				"operation": "database_query",
				"table":     "users",
			})

		// Simulate database query
		time.Sleep(50 * time.Millisecond)

		dbSpan.WithOutput(map[string]interface{}{
			"rows_returned": 1,
			"execution_time": "50ms",
		}).WithEndTime(time.Now())

		if err := dbSpan.End(); err != nil {
			log.Printf("Failed to submit database span: %v", err)
		}
	}

	// Mock user data
	user := map[string]interface{}{
		"id":    userID,
		"name":  "John Doe",
		"email": "john@example.com",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	trace := middleware.TraceFromContext(r.Context())

	var userData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&userData); err != nil {
		if trace != nil {
			trace.WithLevel("ERROR").
				WithStatusMessage("Invalid JSON in request body")
		}
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if trace != nil {
		// Add input data to trace
		trace.WithInput(userData)

		// Create validation span
		validationSpan := trace.Span("input-validation").
			WithInput(userData).
			WithStartTime(time.Now())

		// Simulate validation
		time.Sleep(10 * time.Millisecond)

		validationResult := map[string]interface{}{
			"valid": true,
			"required_fields": []string{"name", "email"},
			"validation_rules": []string{"email_format", "name_length"},
		}

		validationSpan.WithOutput(validationResult).
			WithEndTime(time.Now())

		if err := validationSpan.End(); err != nil {
			log.Printf("Failed to submit validation span: %v", err)
		}

		// Create database insertion span
		dbSpan := trace.Span("database-insert").
			WithInput(map[string]interface{}{
				"query": "INSERT INTO users (name, email) VALUES (?, ?)",
				"params": []string{userData["name"].(string), userData["email"].(string)},
			}).
			WithStartTime(time.Now())

		// Simulate database insert
		time.Sleep(30 * time.Millisecond)

		newUserID := "user_12345"
		dbSpan.WithOutput(map[string]interface{}{
			"user_id": newUserID,
			"rows_affected": 1,
		}).WithEndTime(time.Now())

		if err := dbSpan.End(); err != nil {
			log.Printf("Failed to submit database span: %v", err)
		}
	}

	// Return created user
	response := map[string]interface{}{
		"id":      "user_12345",
		"name":    userData["name"],
		"email":   userData["email"],
		"created": true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Custom middleware for adding business logic tracing
func BusinessLogicMiddleware(langfuse *client.Langfuse) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get existing trace from Langfuse middleware
			trace := middleware.TraceFromContext(r.Context())
			
			if trace != nil {
				// Add business logic span
				businessSpan := trace.Span("business-logic").
					WithInput(map[string]interface{}{
						"method":    r.Method,
						"path":      r.URL.Path,
						"timestamp": time.Now(),
					}).
					WithStartTime(time.Now()).
					WithMetadata(map[string]interface{}{
						"middleware": "business_logic",
						"layer":      "application",
					})

				// Create custom response writer to capture response
				rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

				// Process request
				next.ServeHTTP(rw, r)

				// Complete business span
				businessSpan.WithOutput(map[string]interface{}{
					"status_code":    rw.statusCode,
					"response_size":  rw.bytesWritten,
					"processing_time": time.Since(businessSpan.GetStartTime()).String(),
				}).WithEndTime(time.Now())

				if err := businessSpan.End(); err != nil {
					log.Printf("Failed to submit business logic span: %v", err)
				}
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}

// Custom response writer to capture response details
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

func main() {
	// Initialize Langfuse client
	langfuseClient, err := client.NewWithOptions(
		client.WithCredentials("your-public-key", "your-secret-key"),
		client.WithHost("https://cloud.langfuse.com"),
		client.WithDebug(true),
		client.WithEnvironment("development"),
		client.WithFlushSettings(5, 10*time.Second), // More frequent flushing for demo
	)
	if err != nil {
		log.Fatal("Failed to create Langfuse client:", err)
	}
	defer langfuseClient.Shutdown(context.Background())

	// Create HTTP server with Langfuse middleware
	mux := http.NewServeMux()

	// Initialize handlers
	userHandler := &UserHandler{langfuse: langfuseClient}

	// Register routes
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			userHandler.GetUser(w, r)
		case http.MethodPost:
			userHandler.CreateUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		trace := middleware.TraceFromContext(r.Context())
		if trace != nil {
			trace.WithInput("health_check").WithOutput("healthy")
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	// Apply middleware in order (outer to inner)
	handler := middleware.HTTP(langfuseClient)(mux)                          // Langfuse tracing
	handler = BusinessLogicMiddleware(langfuseClient)(handler)                // Custom business logic
	handler = loggingMiddleware(handler)                                      // Request logging

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	fmt.Println("=== HTTP Middleware Example ===")
	fmt.Println("Server starting on :8080")
	fmt.Println("\nTest endpoints:")
	fmt.Println("GET  /users?id=123")
	fmt.Println("POST /users (with JSON body)")
	fmt.Println("GET  /health")
	fmt.Println("\nExample curl commands:")
	fmt.Println(`curl "http://localhost:8080/users?id=123"`)
	fmt.Println(`curl -X POST "http://localhost:8080/users" -H "Content-Type: application/json" -d '{"name":"Alice","email":"alice@example.com"}'`)
	fmt.Println(`curl "http://localhost:8080/health"`)
	fmt.Println("\nPress Ctrl+C to stop the server")

	log.Fatal(server.ListenAndServe())
}

// Simple logging middleware for demonstration
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Wrap response writer to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(rw, r)
		
		log.Printf("%s %s %d %v", 
			r.Method, 
			r.URL.Path, 
			rw.statusCode, 
			time.Since(start))
	})
}