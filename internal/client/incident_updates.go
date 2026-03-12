package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// ListIncidentUpdates retrieves incident updates with optional filtering
func (c *Client) ListIncidentUpdates(ctx context.Context, opts *ListIncidentUpdatesOptions) (*ListIncidentUpdatesResponse, error) {
	// Set default page size
	pageSize := 25
	if opts != nil && opts.PageSize > 0 {
		pageSize = opts.PageSize
	}

	params := url.Values{}
	params.Set("page_size", strconv.Itoa(pageSize)) // Always set (may be required)

	if opts != nil {
		if opts.IncidentID != "" {
			params.Set("incident_id", opts.IncidentID)
		}
		if opts.After != "" {
			params.Set("after", opts.After)
		}
	}

	respBody, err := c.doRequest(ctx, "GET", "/incident_updates", params, nil)
	if err != nil {
		return nil, err
	}

	var response ListIncidentUpdatesResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// GetIncidentUpdate retrieves a specific incident update by ID
func (c *Client) GetIncidentUpdate(ctx context.Context, id string) (*IncidentUpdate, error) {
	respBody, err := c.doRequest(ctx, "GET", fmt.Sprintf("/incident_updates/%s", id), nil, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		IncidentUpdate IncidentUpdate `json:"incident_update"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response.IncidentUpdate, nil
}

// CreateIncidentUpdate creates a new incident update
func (c *Client) CreateIncidentUpdate(ctx context.Context, req *CreateIncidentUpdateRequest) (*IncidentUpdate, error) {
	// Validate required fields
	if req.IncidentID == "" {
		return nil, fmt.Errorf("incident_id is required")
	}
	if req.Message == "" {
		return nil, fmt.Errorf("message is required")
	}

	respBody, err := c.doRequest(ctx, "POST", "/incident_updates", nil, req)
	if err != nil {
		return nil, err
	}

	var response struct {
		IncidentUpdate IncidentUpdate `json:"incident_update"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response.IncidentUpdate, nil
}

// DeleteIncidentUpdate deletes an incident update
func (c *Client) DeleteIncidentUpdate(ctx context.Context, id string) error {
	_, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/incident_updates/%s", id), nil, nil)
	return err
}
