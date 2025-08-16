package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"eino/pkg/langfuse/internal/utils"
)

// CircuitBreakerState represents the state of the circuit breaker
type CircuitBreakerState int

const (
	// StateClosed means the circuit breaker is closed and requests are passing through
	StateClosed CircuitBreakerState = iota
	// StateOpen means the circuit breaker is open and requests are being rejected
	StateOpen
	// StateHalfOpen means the circuit breaker is testing if the service has recovered
	StateHalfOpen
)

// String returns the string representation of the circuit breaker state
func (s CircuitBreakerState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreakerConfig configures the circuit breaker behavior
type CircuitBreakerConfig struct {
	// MaxFailures is the maximum number of failures before opening the circuit
	MaxFailures int
	
	// ResetTimeout is how long to wait before attempting to reset the circuit
	ResetTimeout time.Duration
	
	// HalfOpenMaxRequests is the maximum number of requests allowed in half-open state
	HalfOpenMaxRequests int
	
	// SuccessThreshold is the number of consecutive successes needed to close the circuit
	SuccessThreshold int
	
	// CounterResetInterval is how often to reset failure counters
	CounterResetInterval time.Duration
	
	// IsFailure determines if an error should be counted as a failure
	IsFailure func(err error) bool
	
	// OnStateChange is called when the circuit breaker state changes
	OnStateChange func(from, to CircuitBreakerState)
}

// DefaultCircuitBreakerConfig returns a CircuitBreakerConfig with sensible defaults
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		MaxFailures:          5,
		ResetTimeout:         30 * time.Second,
		HalfOpenMaxRequests:  3,
		SuccessThreshold:     2,
		CounterResetInterval: 5 * time.Minute,
		IsFailure: func(err error) bool {
			// Only count server errors and network errors as failures
			if langfuseErr, ok := err.(*utils.LangfuseError); ok {
				return langfuseErr.Type == utils.ErrorTypeNetwork ||
					   langfuseErr.Type == utils.ErrorTypeServerError ||
					   langfuseErr.Type == utils.ErrorTypeTimeout
			}
			return true // Default to counting all errors as failures
		},
		OnStateChange: func(from, to CircuitBreakerState) {
			// Default: do nothing
		},
	}
}

// CircuitBreaker implements the circuit breaker pattern for resilient HTTP calls
type CircuitBreaker struct {
	config *CircuitBreakerConfig
	
	mu              sync.RWMutex
	state           CircuitBreakerState
	failures        int
	requests        int
	successes       int
	lastFailureTime time.Time
	lastResetTime   time.Time
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration
func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}
	
	cb := &CircuitBreaker{
		config:        config,
		state:         StateClosed,
		lastResetTime: time.Now(),
	}
	
	// Start counter reset timer if configured
	if config.CounterResetInterval > 0 {
		go cb.startCounterResetTimer()
	}
	
	return cb
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	// Check if request is allowed
	if err := cb.allowRequest(); err != nil {
		return err
	}
	
	// Execute the function
	start := time.Now()
	err := fn()
	duration := time.Since(start)
	
	// Record the result
	cb.recordResult(err, duration)
	
	return err
}

// allowRequest checks if a request is allowed based on the current state
func (cb *CircuitBreaker) allowRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	switch cb.state {
	case StateClosed:
		return nil // Allow request
		
	case StateOpen:
		// Check if we should transition to half-open
		if time.Since(cb.lastFailureTime) > cb.config.ResetTimeout {
			cb.setState(StateHalfOpen)
			cb.requests = 0
			cb.successes = 0
			return nil // Allow the first request in half-open state
		}
		return utils.NewNetworkError(
			"circuit_breaker_open",
			"circuit breaker is open, rejecting request",
			fmt.Errorf("circuit breaker open for %v", time.Since(cb.lastFailureTime)),
		)
		
	case StateHalfOpen:
		if cb.requests >= cb.config.HalfOpenMaxRequests {
			return utils.NewNetworkError(
				"circuit_breaker_half_open_limit",
				"circuit breaker half-open request limit exceeded",
				fmt.Errorf("half-open requests: %d/%d", cb.requests, cb.config.HalfOpenMaxRequests),
			)
		}
		cb.requests++
		return nil // Allow request
		
	default:
		return utils.NewNetworkError(
			"circuit_breaker_unknown_state",
			"circuit breaker in unknown state",
			fmt.Errorf("unknown state: %v", cb.state),
		)
	}
}

// recordResult records the result of a request and updates the circuit breaker state
func (cb *CircuitBreaker) recordResult(err error, duration time.Duration) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	if err != nil && cb.config.IsFailure(err) {
		cb.recordFailure()
	} else {
		cb.recordSuccess()
	}
}

// recordFailure records a failure and potentially opens the circuit
func (cb *CircuitBreaker) recordFailure() {
	cb.failures++
	cb.lastFailureTime = time.Now()
	
	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.config.MaxFailures {
			cb.setState(StateOpen)
		}
		
	case StateHalfOpen:
		// Any failure in half-open state immediately opens the circuit
		cb.setState(StateOpen)
		cb.failures = cb.config.MaxFailures // Ensure we stay open
	}
}

// recordSuccess records a success and potentially closes the circuit
func (cb *CircuitBreaker) recordSuccess() {
	switch cb.state {
	case StateClosed:
		// Reset failure count on success
		cb.failures = 0
		
	case StateHalfOpen:
		cb.successes++
		if cb.successes >= cb.config.SuccessThreshold {
			cb.setState(StateClosed)
			cb.failures = 0
			cb.successes = 0
			cb.requests = 0
		}
	}
}

// setState changes the circuit breaker state and notifies listeners
func (cb *CircuitBreaker) setState(newState CircuitBreakerState) {
	if cb.state != newState {
		oldState := cb.state
		cb.state = newState
		
		// Notify state change
		if cb.config.OnStateChange != nil {
			go cb.config.OnStateChange(oldState, newState)
		}
	}
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Failures returns the current failure count
func (cb *CircuitBreaker) Failures() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failures
}

// Reset manually resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	oldState := cb.state
	cb.state = StateClosed
	cb.failures = 0
	cb.successes = 0
	cb.requests = 0
	cb.lastResetTime = time.Now()
	
	if cb.config.OnStateChange != nil && oldState != StateClosed {
		go cb.config.OnStateChange(oldState, StateClosed)
	}
}

// Stats returns statistics about the circuit breaker
func (cb *CircuitBreaker) Stats() CircuitBreakerStats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	
	return CircuitBreakerStats{
		State:           cb.state,
		Failures:        cb.failures,
		Requests:        cb.requests,
		Successes:       cb.successes,
		LastFailureTime: cb.lastFailureTime,
		LastResetTime:   cb.lastResetTime,
	}
}

// CircuitBreakerStats represents statistics about the circuit breaker
type CircuitBreakerStats struct {
	State           CircuitBreakerState
	Failures        int
	Requests        int
	Successes       int
	LastFailureTime time.Time
	LastResetTime   time.Time
}

// String returns a string representation of the circuit breaker stats
func (s CircuitBreakerStats) String() string {
	return fmt.Sprintf(
		"CircuitBreaker{state=%s, failures=%d, requests=%d, successes=%d}",
		s.State.String(), s.Failures, s.Requests, s.Successes,
	)
}

// startCounterResetTimer starts a timer to periodically reset failure counters
func (cb *CircuitBreaker) startCounterResetTimer() {
	ticker := time.NewTicker(cb.config.CounterResetInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		cb.mu.Lock()
		if cb.state == StateClosed && time.Since(cb.lastResetTime) > cb.config.CounterResetInterval {
			cb.failures = 0
			cb.lastResetTime = time.Now()
		}
		cb.mu.Unlock()
	}
}

// IsOpen returns true if the circuit breaker is open
func (cb *CircuitBreaker) IsOpen() bool {
	return cb.State() == StateOpen
}

// IsClosed returns true if the circuit breaker is closed
func (cb *CircuitBreaker) IsClosed() bool {
	return cb.State() == StateClosed
}

// IsHalfOpen returns true if the circuit breaker is half-open
func (cb *CircuitBreaker) IsHalfOpen() bool {
	return cb.State() == StateHalfOpen
}