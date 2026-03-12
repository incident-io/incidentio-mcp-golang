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
	BaseURLV1      = "https://api.incident.io/v1"
	BaseURLV3      = "https://api.incident.io/v3"
	userAgent      = "incidentio-mcp-server/0.1.0"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	debug      bool
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

	// Clone the default transport to preserve HTTP/2 support, connection pooling,
	// TLS handshake timeout, and other production-ready defaults.
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	return &Client{
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
		baseURL: baseURL,
		apiKey:  apiKey,
		debug:   os.Getenv("MCP_DEBUG") != "",
	}, nil
}

// DoRequest exposes the internal doRequest method using the default base URL
func (c *Client) DoRequest(ctx context.Context, method, path string, params url.Values, body interface{}) ([]byte, error) {
	return c.doRequestWithBase(ctx, c.baseURL, method, path, params, body)
}

// DoRequestV1 makes a request against the V1 API
func (c *Client) DoRequestV1(ctx context.Context, method, path string, params url.Values, body interface{}) ([]byte, error) {
	return c.doRequestWithBase(ctx, BaseURLV1, method, path, params, body)
}

// DoRequestV3 makes a request against the V3 API
func (c *Client) DoRequestV3(ctx context.Context, method, path string, params url.Values, body interface{}) ([]byte, error) {
	return c.doRequestWithBase(ctx, BaseURLV3, method, path, params, body)
}

func (c *Client) doRequest(ctx context.Context, method, path string, params url.Values, body interface{}) ([]byte, error) {
	return c.doRequestWithBase(ctx, c.baseURL, method, path, params, body)
}

func (c *Client) doRequestWithBase(ctx context.Context, baseURL, method, path string, params url.Values, body interface{}) ([]byte, error) {
	endpoint := baseURL + path

	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	// Debug logging to stderr (won't interfere with MCP protocol)
	if c.debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] %s %s\n", method, endpoint)
	}

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errorResp ErrorResponse
		if err := json.Unmarshal(respBody, &errorResp); err != nil {
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
		}
		// If error message is empty, show the full response
		if errorResp.Error.Message == "" {
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
		}
		// Include more details from the error response
		errorMsg := fmt.Sprintf("API error: %s (HTTP %d)", errorResp.Error.Message, resp.StatusCode)
		if errorResp.Error.Code != "" {
			errorMsg += fmt.Sprintf(" [code: %s]", errorResp.Error.Code)
		}
		return nil, fmt.Errorf("%s. Full response: %s", errorMsg, string(respBody))
	}

	return respBody, nil
}

type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"error"`
}
