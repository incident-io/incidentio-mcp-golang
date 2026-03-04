package client

import (
	"net/http"
	"testing"
	"time"
)

func TestClient_ConnectionPooling(t *testing.T) {
	// Set environment variable for testing
	t.Setenv("INCIDENT_IO_API_KEY", "test-key")

	client, err := NewClient()
	assertNoError(t, err)

	// Verify transport is configured
	transport, ok := client.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Expected http.Transport")
	}

	// Verify connection pooling settings
	if transport.MaxIdleConns != 100 {
		t.Errorf("Expected MaxIdleConns=100, got %d", transport.MaxIdleConns)
	}

	if transport.MaxIdleConnsPerHost != 10 {
		t.Errorf("Expected MaxIdleConnsPerHost=10, got %d", transport.MaxIdleConnsPerHost)
	}

	if transport.IdleConnTimeout != 90*time.Second {
		t.Errorf("Expected IdleConnTimeout=90s, got %v", transport.IdleConnTimeout)
	}

	if transport.DisableKeepAlives {
		t.Error("Expected DisableKeepAlives=false")
	}

	// Verify timeouts
	if transport.TLSHandshakeTimeout != 10*time.Second {
		t.Errorf("Expected TLSHandshakeTimeout=10s, got %v", transport.TLSHandshakeTimeout)
	}

	if transport.ResponseHeaderTimeout != 10*time.Second {
		t.Errorf("Expected ResponseHeaderTimeout=10s, got %v", transport.ResponseHeaderTimeout)
	}

	if transport.ExpectContinueTimeout != 1*time.Second {
		t.Errorf("Expected ExpectContinueTimeout=1s, got %v", transport.ExpectContinueTimeout)
	}

	// Verify TLS config
	if transport.TLSClientConfig == nil {
		t.Fatal("Expected TLS config")
	}

	if transport.TLSClientConfig.MinVersion != 0x0303 { // TLS 1.2
		t.Errorf("Expected TLS 1.2 minimum version")
	}
}

func TestClient_Timeout(t *testing.T) {
	t.Setenv("INCIDENT_IO_API_KEY", "test-key")

	client, err := NewClient()
	assertNoError(t, err)

	// Verify client timeout
	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("Expected client timeout=30s, got %v", client.httpClient.Timeout)
	}
}

func TestClient_CacheInitialized(t *testing.T) {
	t.Setenv("INCIDENT_IO_API_KEY", "test-key")

	client, err := NewClient()
	assertNoError(t, err)

	// Verify cache is initialized
	if client.cache == nil {
		t.Fatal("Expected cache to be initialized")
	}

	// Test cache works
	client.cache.Set("test", "value")
	val, found := client.cache.Get("test")
	if !found {
		t.Error("Expected to find test key in cache")
	}
	if val != "value" {
		t.Errorf("Expected 'value', got %v", val)
	}
}

func BenchmarkClient_WithConnectionPooling(b *testing.B) {
	callCount := 0
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			return mockResponse(http.StatusOK, `{"severities": []}`), nil
		},
	}

	client := NewTestClient(mockClient)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.ListSeverities()
	}
}

func BenchmarkClient_CacheHit(b *testing.B) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return mockResponse(http.StatusOK, `{"severities": []}`), nil
		},
	}

	client := NewTestClient(mockClient)

	// Prime the cache
	_, _ = client.ListSeverities()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.ListSeverities()
	}
}

func BenchmarkClient_CacheMiss(b *testing.B) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return mockResponse(http.StatusOK, `{"severities": []}`), nil
		},
	}

	client := NewTestClient(mockClient)
	// Use very short TTL to force cache misses
	client.cache = NewCache(1 * time.Nanosecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.ListSeverities()
		time.Sleep(2 * time.Nanosecond) // Ensure cache expires
	}
}

// Made with Bob
