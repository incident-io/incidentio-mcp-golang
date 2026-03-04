package client

import (
	"context"
	"testing"
	"time"
)

func TestRateLimiter_BasicOperation(t *testing.T) {
	rl := NewRateLimiter(2, 1) // 2 tokens, 1 per second

	ctx := context.Background()

	// First two requests should succeed immediately
	start := time.Now()

	err := rl.Wait(ctx)
	assertNoError(t, err)

	err = rl.Wait(ctx)
	assertNoError(t, err)

	elapsed := time.Since(start)
	if elapsed > 100*time.Millisecond {
		t.Errorf("First two requests should be immediate, took %v", elapsed)
	}

	// Third request should wait for refill
	start = time.Now()
	err = rl.Wait(ctx)
	assertNoError(t, err)
	elapsed = time.Since(start)

	// Should wait approximately 1 second for refill
	if elapsed < 900*time.Millisecond || elapsed > 1100*time.Millisecond {
		t.Errorf("Expected ~1s wait, got %v", elapsed)
	}
}

func TestRateLimiter_ContextCancellation(t *testing.T) {
	rl := NewRateLimiter(1, 0.1) // 1 token, very slow refill

	// Use up the token
	ctx := context.Background()
	err := rl.Wait(ctx)
	assertNoError(t, err)

	// Create a context that will be cancelled
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// This should fail due to context cancellation
	err = rl.Wait(ctx)
	if err == nil {
		t.Error("Expected context cancellation error")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("Expected DeadlineExceeded, got %v", err)
	}
}

func TestRateLimiter_Stats(t *testing.T) {
	rl := NewRateLimiter(10, 5)

	ctx := context.Background()

	// Make some requests
	for i := 0; i < 5; i++ {
		_ = rl.Wait(ctx)
	}

	stats := rl.Stats()

	if stats.RequestsCount != 5 {
		t.Errorf("Expected 5 requests, got %d", stats.RequestsCount)
	}

	if stats.MaxTokens != 10 {
		t.Errorf("Expected max tokens 10, got %f", stats.MaxTokens)
	}

	if stats.RefillRate != 5 {
		t.Errorf("Expected refill rate 5, got %f", stats.RefillRate)
	}
}

func TestRateLimiter_Refill(t *testing.T) {
	rl := NewRateLimiter(5, 10) // 5 tokens, 10 per second

	ctx := context.Background()

	// Use all tokens
	for i := 0; i < 5; i++ {
		err := rl.Wait(ctx)
		assertNoError(t, err)
	}

	// Wait for refill (should get 1 token in 100ms at 10/sec rate)
	time.Sleep(150 * time.Millisecond)

	// Should be able to make another request without long wait
	start := time.Now()
	err := rl.Wait(ctx)
	assertNoError(t, err)
	elapsed := time.Since(start)

	if elapsed > 50*time.Millisecond {
		t.Errorf("Expected quick request after refill, took %v", elapsed)
	}
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	rl := NewRateLimiter(100, 50)

	ctx := context.Background()
	done := make(chan bool)
	errors := make(chan error, 10)

	// Launch multiple goroutines
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				if err := rl.Wait(ctx); err != nil {
					errors <- err
					return
				}
			}
			done <- true
		}()
	}

	// Wait for all to complete
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// Success
		case err := <-errors:
			t.Errorf("Unexpected error: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for goroutines")
		}
	}

	stats := rl.Stats()
	if stats.RequestsCount != 100 {
		t.Errorf("Expected 100 requests, got %d", stats.RequestsCount)
	}
}

func TestRetryConfig_ShouldRetry(t *testing.T) {
	config := DefaultRetryConfig()

	tests := []struct {
		name       string
		statusCode int
		err        error
		want       bool
	}{
		{"429 Too Many Requests", 429, nil, true},
		{"500 Internal Server Error", 500, nil, true},
		{"502 Bad Gateway", 502, nil, true},
		{"503 Service Unavailable", 503, nil, true},
		{"504 Gateway Timeout", 504, nil, true},
		{"408 Request Timeout", 408, nil, true},
		{"200 OK", 200, nil, false},
		{"404 Not Found", 404, nil, false},
		{"Network Error", 0, context.DeadlineExceeded, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.ShouldRetry(tt.statusCode, tt.err)
			if got != tt.want {
				t.Errorf("ShouldRetry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRetryConfig_CalculateBackoff(t *testing.T) {
	config := DefaultRetryConfig()

	tests := []struct {
		attempt int
		min     time.Duration
		max     time.Duration
	}{
		{0, 100 * time.Millisecond, 100 * time.Millisecond},
		{1, 200 * time.Millisecond, 200 * time.Millisecond},
		{2, 400 * time.Millisecond, 400 * time.Millisecond},
		{3, 800 * time.Millisecond, 800 * time.Millisecond},
		{10, 10 * time.Second, 10 * time.Second}, // Should cap at MaxBackoff
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.attempt)), func(t *testing.T) {
			backoff := config.CalculateBackoff(tt.attempt)
			if backoff < tt.min || backoff > tt.max {
				t.Errorf("CalculateBackoff(%d) = %v, want between %v and %v",
					tt.attempt, backoff, tt.min, tt.max)
			}
		})
	}
}

func TestRetryableError(t *testing.T) {
	err := &RetryableError{
		Err:        context.DeadlineExceeded,
		Attempt:    3,
		MaxRetries: 3,
		StatusCode: 503,
	}

	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Error message should not be empty")
	}

	// Test Unwrap
	unwrapped := err.Unwrap()
	if unwrapped != context.DeadlineExceeded {
		t.Errorf("Expected DeadlineExceeded, got %v", unwrapped)
	}
}

func BenchmarkRateLimiter_Wait(b *testing.B) {
	rl := NewRateLimiter(float64(b.N), float64(b.N)) // Enough tokens for all iterations
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rl.Wait(ctx)
	}
}

func BenchmarkRateLimiter_Stats(b *testing.B) {
	rl := NewRateLimiter(100, 50)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rl.Stats()
	}
}

// Made with Bob
