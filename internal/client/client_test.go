package client

import (
	"bytes"
	"io"
	"net/http"
	"testing"
	"time"
)

// MockHTTPClient is a mock implementation of http.Client for testing
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

// NewTestClient creates a client with a mock HTTP client for testing
func NewTestClient(mockClient *MockHTTPClient) *Client {
	return &Client{
		httpClient: &http.Client{
			Transport: mockClient,
		},
		baseURL:        "https://api.test.incident.io",
		apiKey:         "test-api-key",
		cache:          NewCache(5 * time.Minute), // Initialize cache for tests
		rateLimiter:    NewRateLimiter(1000, 100), // High limits for tests
		circuitBreaker: NewCircuitBreaker(DefaultCircuitBreakerConfig()),
		retryConfig:    DefaultRetryConfig(),
	}
}

// RoundTrip implements the http.RoundTripper interface
func (m *MockHTTPClient) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

// Helper function to create a mock response
func mockResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Header:     make(http.Header),
	}
}

// Helper function to assert no error
func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Helper function to assert error
func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error but got none")
	}
}

// Helper function to assert string equality
func assertEqual(t *testing.T, expected, actual string) {
	t.Helper()
	if expected != actual {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}
