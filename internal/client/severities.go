package client

import (
	"encoding/json"
	"fmt"
)

// Using Severity type from types.go

// ListSeveritiesResponse represents the response from listing severities
type ListSeveritiesResponse struct {
	Severities []Severity `json:"severities"`
}

// ListSeverities returns all severities (cached for 5 minutes)
func (c *Client) ListSeverities() (*ListSeveritiesResponse, error) {
	// Check cache first
	cacheKey := "severities:list"
	if cached, found := c.cache.Get(cacheKey); found {
		return cached.(*ListSeveritiesResponse), nil
	}

	// Note: Severities are under V1 API, not V2
	// We need to temporarily change the base URL for this request
	originalBaseURL := c.BaseURL()
	c.SetBaseURL("https://api.incident.io/v1")
	defer func() { c.SetBaseURL(originalBaseURL) }()

	respBody, err := c.doRequest("GET", "/severities", nil, nil)
	if err != nil {
		return nil, err
	}

	var response ListSeveritiesResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Cache the response
	c.cache.Set(cacheKey, &response)

	return &response, nil
}

// GetSeverity retrieves a specific severity by ID (cached for 5 minutes)
func (c *Client) GetSeverity(id string) (*Severity, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("severity:%s", id)
	if cached, found := c.cache.Get(cacheKey); found {
		return cached.(*Severity), nil
	}

	// Note: Severities are under V1 API, not V2
	// We need to temporarily change the base URL for this request
	originalBaseURL := c.BaseURL()
	c.SetBaseURL("https://api.incident.io/v1")
	defer func() { c.SetBaseURL(originalBaseURL) }()

	respBody, err := c.doRequest("GET", fmt.Sprintf("/severities/%s", id), nil, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Severity Severity `json:"severity"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Cache the response
	c.cache.Set(cacheKey, &response.Severity)

	return &response.Severity, nil
}
