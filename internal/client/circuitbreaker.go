package client

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	// ErrCircuitOpen is returned when the circuit breaker is open
	ErrCircuitOpen = errors.New("circuit breaker is open")
	// ErrTooManyRequests is returned when too many requests are in flight
	ErrTooManyRequests = errors.New("too many requests")
)

// CircuitState represents the state of the circuit breaker
type CircuitState int

const (
	// StateClosed means requests are allowed through
	StateClosed CircuitState = iota
	// StateOpen means requests are blocked
	StateOpen
	// StateHalfOpen means limited requests are allowed to test recovery
	StateHalfOpen
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreakerConfig configures circuit breaker behavior
type CircuitBreakerConfig struct {
	// MaxFailures is the number of failures before opening the circuit
	MaxFailures int
	// Timeout is how long to wait before attempting to close the circuit
	Timeout time.Duration
	// HalfOpenMaxRequests is the max requests allowed in half-open state
	HalfOpenMaxRequests int
	// FailureThreshold is the percentage of failures to trigger opening (0-1)
	FailureThreshold float64
	// MinRequests is the minimum number of requests before checking threshold
	MinRequests int
}

// DefaultCircuitBreakerConfig returns sensible defaults
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		MaxFailures:         5,
		Timeout:             30 * time.Second,
		HalfOpenMaxRequests: 3,
		FailureThreshold:    0.5, // 50% failure rate
		MinRequests:         10,
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu                  sync.RWMutex
	config              *CircuitBreakerConfig
	state               CircuitState
	failures            int
	successes           int
	requests            int
	lastFailureTime     time.Time
	lastStateChange     time.Time
	halfOpenRequests    int
	consecutiveFailures int
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}

	return &CircuitBreaker{
		config:          config,
		state:           StateClosed,
		lastStateChange: time.Now(),
	}
}

// Call executes a function with circuit breaker protection
func (cb *CircuitBreaker) Call(fn func() error) error {
	// Check if we can proceed
	if err := cb.beforeRequest(); err != nil {
		return err
	}

	// Execute the function
	err := fn()

	// Record the result
	cb.afterRequest(err)

	return err
}

// beforeRequest checks if the request should be allowed
func (cb *CircuitBreaker) beforeRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		// Allow request
		cb.requests++
		return nil

	case StateOpen:
		// Check if timeout has elapsed
		if time.Since(cb.lastStateChange) > cb.config.Timeout {
			// Transition to half-open
			cb.state = StateHalfOpen
			cb.halfOpenRequests = 0
			cb.lastStateChange = time.Now()
			cb.requests++
			cb.halfOpenRequests++
			return nil
		}
		// Circuit is still open
		return ErrCircuitOpen

	case StateHalfOpen:
		// Allow limited requests
		if cb.halfOpenRequests >= cb.config.HalfOpenMaxRequests {
			return ErrTooManyRequests
		}
		cb.requests++
		cb.halfOpenRequests++
		return nil

	default:
		return fmt.Errorf("unknown circuit breaker state: %v", cb.state)
	}
}

// afterRequest records the result of a request
func (cb *CircuitBreaker) afterRequest(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.onFailure()
	} else {
		cb.onSuccess()
	}
}

// onFailure handles a failed request
func (cb *CircuitBreaker) onFailure() {
	cb.failures++
	cb.consecutiveFailures++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		// Check if we should open the circuit
		if cb.shouldOpen() {
			cb.state = StateOpen
			cb.lastStateChange = time.Now()
		}

	case StateHalfOpen:
		// Any failure in half-open state reopens the circuit
		cb.state = StateOpen
		cb.lastStateChange = time.Now()
		cb.halfOpenRequests = 0
	}
}

// onSuccess handles a successful request
func (cb *CircuitBreaker) onSuccess() {
	cb.successes++
	cb.consecutiveFailures = 0

	switch cb.state {
	case StateHalfOpen:
		// If all half-open requests succeed, close the circuit
		if cb.halfOpenRequests >= cb.config.HalfOpenMaxRequests {
			cb.state = StateClosed
			cb.lastStateChange = time.Now()
			cb.reset()
		}
	}
}

// shouldOpen determines if the circuit should open based on failure rate
func (cb *CircuitBreaker) shouldOpen() bool {
	// Check consecutive failures
	if cb.consecutiveFailures >= cb.config.MaxFailures {
		return true
	}

	// Check failure rate if we have enough requests
	if cb.requests >= cb.config.MinRequests {
		failureRate := float64(cb.failures) / float64(cb.requests)
		if failureRate >= cb.config.FailureThreshold {
			return true
		}
	}

	return false
}

// reset clears the counters
func (cb *CircuitBreaker) reset() {
	cb.failures = 0
	cb.successes = 0
	cb.requests = 0
	cb.consecutiveFailures = 0
	cb.halfOpenRequests = 0
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Stats returns circuit breaker statistics
func (cb *CircuitBreaker) Stats() CircuitBreakerStats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return CircuitBreakerStats{
		State:               cb.state,
		Failures:            cb.failures,
		Successes:           cb.successes,
		Requests:            cb.requests,
		ConsecutiveFailures: cb.consecutiveFailures,
		LastFailureTime:     cb.lastFailureTime,
		LastStateChange:     cb.lastStateChange,
	}
}

// CircuitBreakerStats contains circuit breaker statistics
type CircuitBreakerStats struct {
	State               CircuitState
	Failures            int
	Successes           int
	Requests            int
	ConsecutiveFailures int
	LastFailureTime     time.Time
	LastStateChange     time.Time
}

// Reset manually resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.lastStateChange = time.Now()
	cb.reset()
}

// Made with Bob
