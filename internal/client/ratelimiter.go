package client

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	mu             sync.Mutex
	tokens         float64
	maxTokens      float64
	refillRate     float64 // tokens per second
	lastRefill     time.Time
	requestsCount  int64
	throttledCount int64
}

// NewRateLimiter creates a new rate limiter
// maxTokens: maximum number of tokens (burst capacity)
// refillRate: tokens added per second
func NewRateLimiter(maxTokens, refillRate float64) *RateLimiter {
	return &RateLimiter{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Wait blocks until a token is available or context is cancelled
func (rl *RateLimiter) Wait(ctx context.Context) error {
	for {
		if rl.tryAcquire() {
			return nil
		}

		// Calculate wait time until next token
		waitTime := rl.timeUntilNextToken()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			// Try again after waiting
		}
	}
}

// tryAcquire attempts to acquire a token without blocking
func (rl *RateLimiter) tryAcquire() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refill()
	rl.requestsCount++

	if rl.tokens >= 1.0 {
		rl.tokens--
		return true
	}

	rl.throttledCount++
	return false
}

// refill adds tokens based on elapsed time (must be called with lock held)
func (rl *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()

	// Add tokens based on elapsed time
	rl.tokens += elapsed * rl.refillRate

	// Cap at max tokens
	if rl.tokens > rl.maxTokens {
		rl.tokens = rl.maxTokens
	}

	rl.lastRefill = now
}

// timeUntilNextToken calculates how long to wait for the next token
func (rl *RateLimiter) timeUntilNextToken() time.Duration {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.tokens >= 1.0 {
		return 0
	}

	tokensNeeded := 1.0 - rl.tokens
	secondsNeeded := tokensNeeded / rl.refillRate
	return time.Duration(secondsNeeded * float64(time.Second))
}

// Stats returns rate limiter statistics
func (rl *RateLimiter) Stats() RateLimiterStats {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	return RateLimiterStats{
		RequestsCount:  rl.requestsCount,
		ThrottledCount: rl.throttledCount,
		CurrentTokens:  rl.tokens,
		MaxTokens:      rl.maxTokens,
		RefillRate:     rl.refillRate,
	}
}

// RateLimiterStats contains rate limiter statistics
type RateLimiterStats struct {
	RequestsCount  int64
	ThrottledCount int64
	CurrentTokens  float64
	MaxTokens      float64
	RefillRate     float64
}

// RetryConfig configures retry behavior with exponential backoff
type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	Multiplier     float64
	// RetryableStatusCodes are HTTP status codes that should trigger a retry
	RetryableStatusCodes map[int]bool
}

// DefaultRetryConfig returns sensible defaults for retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     10 * time.Second,
		Multiplier:     2.0,
		RetryableStatusCodes: map[int]bool{
			408: true, // Request Timeout
			429: true, // Too Many Requests
			500: true, // Internal Server Error
			502: true, // Bad Gateway
			503: true, // Service Unavailable
			504: true, // Gateway Timeout
		},
	}
}

// ShouldRetry determines if an error or status code should trigger a retry
func (rc *RetryConfig) ShouldRetry(statusCode int, err error) bool {
	// Always retry on network errors
	if err != nil {
		return true
	}

	// Check if status code is retryable
	return rc.RetryableStatusCodes[statusCode]
}

// CalculateBackoff calculates the backoff duration for a given attempt
func (rc *RetryConfig) CalculateBackoff(attempt int) time.Duration {
	if attempt <= 0 {
		return rc.InitialBackoff
	}

	// Exponential backoff: initialBackoff * (multiplier ^ attempt)
	backoff := float64(rc.InitialBackoff) * pow(rc.Multiplier, float64(attempt))

	// Cap at max backoff
	if backoff > float64(rc.MaxBackoff) {
		backoff = float64(rc.MaxBackoff)
	}

	return time.Duration(backoff)
}

// pow is a simple integer power function
func pow(base, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}

// RetryableError wraps an error with retry information
type RetryableError struct {
	Err        error
	Attempt    int
	MaxRetries int
	StatusCode int
}

func (e *RetryableError) Error() string {
	return fmt.Sprintf("request failed after %d/%d attempts (status: %d): %v",
		e.Attempt, e.MaxRetries, e.StatusCode, e.Err)
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// Made with Bob
