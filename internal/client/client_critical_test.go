package client

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// Test 1: Transport configuration - Clone preserves HTTP/2, TLS 1.2, timeouts
func TestTransportConfiguration(t *testing.T) {
	t.Setenv("INCIDENT_IO_API_KEY", "test-key")
	t.Setenv("MCP_DEBUG", "")

	c, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	transport, ok := c.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected *http.Transport")
	}

	// TLS minimum version should be 1.2
	if transport.TLSClientConfig == nil {
		t.Fatal("TLSClientConfig is nil")
	}
	if transport.TLSClientConfig.MinVersion != tls.VersionTLS12 {
		t.Errorf("TLS MinVersion = %d, want %d", transport.TLSClientConfig.MinVersion, tls.VersionTLS12)
	}

	// Clone from DefaultTransport should preserve these defaults
	defaultTransport := http.DefaultTransport.(*http.Transport)

	if transport.MaxIdleConns != defaultTransport.MaxIdleConns {
		t.Errorf("MaxIdleConns = %d, want %d", transport.MaxIdleConns, defaultTransport.MaxIdleConns)
	}
	if transport.IdleConnTimeout != defaultTransport.IdleConnTimeout {
		t.Errorf("IdleConnTimeout = %v, want %v", transport.IdleConnTimeout, defaultTransport.IdleConnTimeout)
	}
	if transport.TLSHandshakeTimeout != defaultTransport.TLSHandshakeTimeout {
		t.Errorf("TLSHandshakeTimeout = %v, want %v", transport.TLSHandshakeTimeout, defaultTransport.TLSHandshakeTimeout)
	}
	if transport.ForceAttemptHTTP2 != defaultTransport.ForceAttemptHTTP2 {
		t.Errorf("ForceAttemptHTTP2 = %v, want %v", transport.ForceAttemptHTTP2, defaultTransport.ForceAttemptHTTP2)
	}

	// Client-level timeout
	if c.httpClient.Timeout != 30*time.Second {
		t.Errorf("Client.Timeout = %v, want 30s", c.httpClient.Timeout)
	}
}

// Test 2: Context cancellation propagates to HTTP request
func TestContextCancellationPropagates(t *testing.T) {
	// Server that blocks until the test is done
	blocked := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-blocked
	}))
	defer srv.Close()
	defer close(blocked)

	c := &Client{
		httpClient: srv.Client(),
		baseURL:    srv.URL,
		apiKey:     "test-key",
	}

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		_, err := c.DoRequest(ctx, "GET", "/test", nil, nil)
		errCh <- err
	}()

	// Cancel the context while the request is in flight
	cancel()

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("expected error from cancelled context")
		}
		if !strings.Contains(err.Error(), "context canceled") {
			t.Errorf("expected context canceled error, got: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("request did not respect context cancellation within 5s")
	}
}

// Test 3: Concurrent client calls are safe (no race on baseURL)
func TestConcurrentClientCallsSafe(t *testing.T) {
	// Run with -race to detect data races
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := &Client{
		httpClient: srv.Client(),
		baseURL:    srv.URL,
		apiKey:     "test-key",
	}

	ctx := context.Background()
	var wg sync.WaitGroup
	const numGoroutines = 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Mix of DoRequest, DoRequestV1, DoRequestV3
			c.DoRequest(ctx, "GET", "/default", nil, nil)
			c.DoRequestV1(ctx, "GET", "/v1path", nil, nil)
			c.DoRequestV3(ctx, "GET", "/v3path", nil, nil)
		}()
	}

	wg.Wait()
}

// Test 4: Version-specific routing - V1/V3 calls hit correct base URL
func TestVersionSpecificRouting(t *testing.T) {
	tests := []struct {
		name     string
		method   func(c *Client, ctx context.Context) ([]byte, error)
		wantBase string
	}{
		{
			name: "DoRequest uses default base URL",
			method: func(c *Client, ctx context.Context) ([]byte, error) {
				return c.DoRequest(ctx, "GET", "/test", nil, nil)
			},
			wantBase: defaultBaseURL,
		},
		{
			name: "DoRequestV1 uses V1 base URL",
			method: func(c *Client, ctx context.Context) ([]byte, error) {
				return c.DoRequestV1(ctx, "GET", "/test", nil, nil)
			},
			wantBase: BaseURLV1,
		},
		{
			name: "DoRequestV3 uses V3 base URL",
			method: func(c *Client, ctx context.Context) ([]byte, error) {
				return c.DoRequestV3(ctx, "GET", "/test", nil, nil)
			},
			wantBase: BaseURLV3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedURL string
			mock := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					capturedURL = req.URL.String()
					return mockResponse(200, `{}`), nil
				},
			}

			c := &Client{
				httpClient: &http.Client{Transport: mock},
				baseURL:    defaultBaseURL,
				apiKey:     "test-key",
			}

			_, err := tt.method(c, context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.HasPrefix(capturedURL, tt.wantBase) {
				t.Errorf("URL = %q, want prefix %q", capturedURL, tt.wantBase)
			}
		})
	}
}

// Test 5: Debug flag cached at construction time
func TestDebugFlagCachedAtConstruction(t *testing.T) {
	// Set MCP_DEBUG before construction
	t.Setenv("INCIDENT_IO_API_KEY", "test-key")
	t.Setenv("MCP_DEBUG", "1")

	c, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if !c.debug {
		t.Error("debug should be true when MCP_DEBUG is set at construction")
	}

	// Unset MCP_DEBUG after construction - client should still have debug=true
	t.Setenv("MCP_DEBUG", "")
	if !c.debug {
		t.Error("debug should remain true after unsetting MCP_DEBUG (cached at construction)")
	}

	// Create another client without MCP_DEBUG
	c2, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c2.debug {
		t.Error("debug should be false when MCP_DEBUG is not set at construction")
	}
}
