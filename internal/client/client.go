package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	defaultBaseURL = "https://api.incident.io/v2"
	userAgent      = "incidentio-mcp-server/0.1.0"
)

type Client struct {
	httpClient     *http.Client
	baseURL        string
	apiKey         string
	cache          *Cache
	rateLimiter    *RateLimiter
	circuitBreaker *CircuitBreaker
	retryConfig    *RetryConfig
}

func NewClient() (*Client, error) {
	apiKey := os.Getenv("INCIDENT_IO_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("INCIDENT_IO_API_KEY environment variable is required")
	}

	baseURL := os.Getenv("INCIDENT_IO_BASE_URL")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	// Configure HTTP transport with connection pooling for better performance
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		// Connection pooling settings
		MaxIdleConns:        100,              // Maximum idle connections across all hosts
		MaxIdleConnsPerHost: 10,               // Maximum idle connections per host
		IdleConnTimeout:     90 * time.Second, // How long idle connections stay open
		DisableKeepAlives:   false,            // Enable keep-alive for connection reuse
		// Timeouts
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &Client{
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
		baseURL:        baseURL,
		apiKey:         apiKey,
		cache:          NewCache(5 * time.Minute), // Cache static data for 5 minutes
		rateLimiter:    NewRateLimiter(100, 10),   // 100 tokens, 10 per second
		circuitBreaker: NewCircuitBreaker(DefaultCircuitBreakerConfig()),
		retryConfig:    DefaultRetryConfig(),
	}, nil
}

// BaseURL returns the current base URL
func (c *Client) BaseURL() string {
	return c.baseURL
}

// SetBaseURL sets the base URL
func (c *Client) SetBaseURL(baseURL string) {
	c.baseURL = baseURL
}

// DoRequest exposes the internal doRequest method
func (c *Client) DoRequest(method, path string, params url.Values, body interface{}) ([]byte, error) {
	return c.doRequest(method, path, params, body)
}

func (c *Client) doRequest(method, path string, params url.Values, body interface{}) ([]byte, error) {
	// Use context with timeout for better control
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return c.doRequestWithContext(ctx, method, path, params, body)
}

// doRequestWithContext performs an HTTP request with context support, rate limiting, circuit breaker, and retry logic
func (c *Client) doRequestWithContext(ctx context.Context, method, path string, params url.Values, body interface{}) ([]byte, error) {
	var lastErr error
	var statusCode int

	// Retry loop with exponential backoff
	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		// Apply rate limiting
		if err := c.rateLimiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter error: %w", err)
		}

		// Execute request with circuit breaker protection
		var respBody []byte
		err := c.circuitBreaker.Call(func() error {
			var err error
			respBody, statusCode, err = c.executeRequest(ctx, method, path, params, body)
			return err
		})

		// If circuit breaker is open, return immediately
		if err == ErrCircuitOpen {
			return nil, fmt.Errorf("circuit breaker is open, service may be down: %w", err)
		}

		// If request succeeded, return
		if err == nil {
			return respBody, nil
		}

		lastErr = err

		// Check if we should retry
		if attempt < c.retryConfig.MaxRetries && c.retryConfig.ShouldRetry(statusCode, err) {
			backoff := c.retryConfig.CalculateBackoff(attempt)

			if os.Getenv("MCP_DEBUG") != "" {
				fmt.Fprintf(os.Stderr, "[DEBUG] Retry attempt %d/%d after %v (status: %d, error: %v)\n",
					attempt+1, c.retryConfig.MaxRetries, backoff, statusCode, err)
			}

			// Wait for backoff period
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
				// Continue to next attempt
			}
		} else {
			// No more retries or not retryable
			break
		}
	}

	// All retries exhausted
	return nil, &RetryableError{
		Err:        lastErr,
		Attempt:    c.retryConfig.MaxRetries + 1,
		MaxRetries: c.retryConfig.MaxRetries,
		StatusCode: statusCode,
	}
}

// executeRequest performs the actual HTTP request
func (c *Client) executeRequest(ctx context.Context, method, path string, params url.Values, body interface{}) ([]byte, int, error) {
	endpoint := c.baseURL + path

	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	// Debug logging to stderr (won't interfere with MCP protocol)
	if os.Getenv("MCP_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "[DEBUG] %s %s\n", method, endpoint)
	}

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errorResp ErrorResponse
		if err := json.Unmarshal(respBody, &errorResp); err != nil {
			return nil, resp.StatusCode, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
		}
		// If error message is empty, show the full response
		if errorResp.Error.Message == "" {
			return nil, resp.StatusCode, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
		}
		// Include more details from the error response
		errorMsg := fmt.Sprintf("API error: %s (HTTP %d)", errorResp.Error.Message, resp.StatusCode)
		if errorResp.Error.Code != "" {
			errorMsg += fmt.Sprintf(" [code: %s]", errorResp.Error.Code)
		}
		return nil, resp.StatusCode, fmt.Errorf("%s. Full response: %s", errorMsg, string(respBody))
	}

	return respBody, resp.StatusCode, nil
}

type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"error"`
}
