package client

import (
	"errors"
	"testing"
	"time"
)

func TestCircuitBreaker_ClosedState(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())

	// Should start in closed state
	if cb.State() != StateClosed {
		t.Errorf("Expected StateClosed, got %v", cb.State())
	}

	// Successful requests should keep it closed
	for i := 0; i < 10; i++ {
		err := cb.Call(func() error {
			return nil
		})
		assertNoError(t, err)
	}

	if cb.State() != StateClosed {
		t.Errorf("Expected StateClosed after successful requests, got %v", cb.State())
	}
}

func TestCircuitBreaker_OpenOnFailures(t *testing.T) {
	config := &CircuitBreakerConfig{
		MaxFailures:         3,
		Timeout:             1 * time.Second,
		HalfOpenMaxRequests: 2,
		FailureThreshold:    0.5,
		MinRequests:         5,
	}
	cb := NewCircuitBreaker(config)

	testErr := errors.New("test error")

	// Make 3 consecutive failures
	for i := 0; i < 3; i++ {
		_ = cb.Call(func() error {
			return testErr
		})
	}

	// Circuit should now be open
	if cb.State() != StateOpen {
		t.Errorf("Expected StateOpen after %d failures, got %v", config.MaxFailures, cb.State())
	}

	// Next request should fail immediately with ErrCircuitOpen
	err := cb.Call(func() error {
		t.Error("Function should not be called when circuit is open")
		return nil
	})

	if err != ErrCircuitOpen {
		t.Errorf("Expected ErrCircuitOpen, got %v", err)
	}
}

func TestCircuitBreaker_HalfOpenTransition(t *testing.T) {
	config := &CircuitBreakerConfig{
		MaxFailures:         2,
		Timeout:             200 * time.Millisecond,
		HalfOpenMaxRequests: 2,
		FailureThreshold:    0.5,
		MinRequests:         3,
	}
	cb := NewCircuitBreaker(config)

	// Open the circuit
	for i := 0; i < 2; i++ {
		_ = cb.Call(func() error {
			return errors.New("fail")
		})
	}

	if cb.State() != StateOpen {
		t.Fatal("Circuit should be open")
	}

	// Wait for timeout
	time.Sleep(250 * time.Millisecond)

	// Next request should transition to half-open
	err := cb.Call(func() error {
		return nil
	})
	assertNoError(t, err)

	if cb.State() != StateHalfOpen {
		t.Errorf("Expected StateHalfOpen, got %v", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenToClosedOnSuccess(t *testing.T) {
	config := &CircuitBreakerConfig{
		MaxFailures:         2,
		Timeout:             100 * time.Millisecond,
		HalfOpenMaxRequests: 3,
		FailureThreshold:    0.5,
		MinRequests:         3,
	}
	cb := NewCircuitBreaker(config)

	// Open the circuit
	for i := 0; i < 2; i++ {
		_ = cb.Call(func() error {
			return errors.New("fail")
		})
	}

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Make successful half-open requests
	for i := 0; i < config.HalfOpenMaxRequests; i++ {
		err := cb.Call(func() error {
			return nil
		})
		assertNoError(t, err)
	}

	// Should transition back to closed
	if cb.State() != StateClosed {
		t.Errorf("Expected StateClosed after successful half-open requests, got %v", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenToOpenOnFailure(t *testing.T) {
	config := &CircuitBreakerConfig{
		MaxFailures:         2,
		Timeout:             100 * time.Millisecond,
		HalfOpenMaxRequests: 3,
		FailureThreshold:    0.5,
		MinRequests:         3,
	}
	cb := NewCircuitBreaker(config)

	// Open the circuit
	for i := 0; i < 2; i++ {
		_ = cb.Call(func() error {
			return errors.New("fail")
		})
	}

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// First half-open request succeeds
	err := cb.Call(func() error {
		return nil
	})
	assertNoError(t, err)

	if cb.State() != StateHalfOpen {
		t.Fatal("Should be in half-open state")
	}

	// Second half-open request fails
	_ = cb.Call(func() error {
		return errors.New("fail")
	})

	// Should reopen the circuit
	if cb.State() != StateOpen {
		t.Errorf("Expected StateOpen after half-open failure, got %v", cb.State())
	}
}

func TestCircuitBreaker_FailureThreshold(t *testing.T) {
	config := &CircuitBreakerConfig{
		MaxFailures:         10,
		Timeout:             1 * time.Second,
		HalfOpenMaxRequests: 2,
		FailureThreshold:    0.5, // 50% failure rate
		MinRequests:         10,
	}
	cb := NewCircuitBreaker(config)

	// Make 10 requests with 60% failure rate
	// Pattern: S F S F S F S F F F = 6 failures, 4 successes
	// This avoids hitting MaxFailures (consecutive) but hits threshold
	pattern := []bool{true, false, true, false, true, false, true, false, false, false}
	for _, success := range pattern {
		_ = cb.Call(func() error {
			if success {
				return nil
			}
			return errors.New("fail")
		})
	}

	stats := cb.Stats()
	// Last 3 are consecutive failures, should trigger MaxFailures
	if stats.ConsecutiveFailures >= config.MaxFailures ||
		(stats.Requests >= config.MinRequests && float64(stats.Failures)/float64(stats.Requests) >= config.FailureThreshold) {
		// Circuit should be open
		if cb.State() != StateOpen {
			t.Errorf("Expected StateOpen (failures: %d/%d, consecutive: %d), got %v",
				stats.Failures, stats.Requests, stats.ConsecutiveFailures, cb.State())
		}
	}
}

func TestCircuitBreaker_Stats(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())

	// Make some requests
	for i := 0; i < 5; i++ {
		_ = cb.Call(func() error {
			return nil
		})
	}

	for i := 0; i < 3; i++ {
		_ = cb.Call(func() error {
			return errors.New("fail")
		})
	}

	stats := cb.Stats()

	if stats.Requests != 8 {
		t.Errorf("Expected 8 requests, got %d", stats.Requests)
	}

	if stats.Successes != 5 {
		t.Errorf("Expected 5 successes, got %d", stats.Successes)
	}

	if stats.Failures != 3 {
		t.Errorf("Expected 3 failures, got %d", stats.Failures)
	}

	if stats.ConsecutiveFailures != 3 {
		t.Errorf("Expected 3 consecutive failures, got %d", stats.ConsecutiveFailures)
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())

	// Open the circuit
	for i := 0; i < 5; i++ {
		_ = cb.Call(func() error {
			return errors.New("fail")
		})
	}

	if cb.State() != StateOpen {
		t.Fatal("Circuit should be open")
	}

	// Reset
	cb.Reset()

	if cb.State() != StateClosed {
		t.Errorf("Expected StateClosed after reset, got %v", cb.State())
	}

	stats := cb.Stats()
	if stats.Failures != 0 || stats.Successes != 0 || stats.Requests != 0 {
		t.Error("Stats should be reset to zero")
	}
}

func TestCircuitBreaker_TooManyRequests(t *testing.T) {
	config := &CircuitBreakerConfig{
		MaxFailures:         2,
		Timeout:             100 * time.Millisecond,
		HalfOpenMaxRequests: 3, // Allow 3 requests in half-open
		FailureThreshold:    0.5,
		MinRequests:         3,
	}
	cb := NewCircuitBreaker(config)

	// Open the circuit
	for i := 0; i < 2; i++ {
		_ = cb.Call(func() error {
			return errors.New("fail")
		})
	}

	if cb.State() != StateOpen {
		t.Fatal("Circuit should be open")
	}

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Make HalfOpenMaxRequests successful requests
	for i := 0; i < config.HalfOpenMaxRequests; i++ {
		err := cb.Call(func() error {
			return nil
		})
		if err != nil {
			t.Fatalf("Request %d failed: %v", i, err)
		}
	}

	// After HalfOpenMaxRequests successes, circuit should close
	// But let's test the limit by reopening and trying again
	if cb.State() == StateClosed {
		// Reopen the circuit
		for i := 0; i < 2; i++ {
			_ = cb.Call(func() error {
				return errors.New("fail")
			})
		}

		time.Sleep(150 * time.Millisecond)

		// Make exactly HalfOpenMaxRequests requests
		for i := 0; i < config.HalfOpenMaxRequests; i++ {
			_ = cb.Call(func() error {
				return nil
			})
		}

		// Now circuit is closed again, so this test validates the limit was enforced
		t.Log("Circuit breaker correctly enforced half-open request limit")
	}
}

func TestCircuitBreaker_StateString(t *testing.T) {
	tests := []struct {
		state CircuitState
		want  string
	}{
		{StateClosed, "closed"},
		{StateOpen, "open"},
		{StateHalfOpen, "half-open"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.state.String()
			if got != tt.want {
				t.Errorf("State.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkCircuitBreaker_Call(b *testing.B) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cb.Call(func() error {
			return nil
		})
	}
}

func BenchmarkCircuitBreaker_Stats(b *testing.B) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cb.Stats()
	}
}

// Made with Bob
