package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// ListActionsOptions represents options for listing actions
type ListActionsOptions struct {
	PageSize   int
	After      string
	IncidentID string
	Status     []string
}

// ListActionsResponse represents the response from listing actions
type ListActionsResponse struct {
	Actions []Action `json:"actions"`
	ListResponse
}

// ListActions retrieves a single page of actions
// Pagination is controlled by the caller using PageSize and After parameters
func (c *Client) ListActions(ctx context.Context, opts *ListActionsOptions) (*ListActionsResponse, error) {
	pageSize := 10 // Conservative default to avoid exceeding MCP client limits
	after := ""

	if opts != nil {
		if opts.PageSize > 0 {
			pageSize = opts.PageSize
		}
		if opts.After != "" {
			after = opts.After
		}
	}

	params := url.Values{}
	params.Set("page_size", strconv.Itoa(pageSize))

	if after != "" {
		params.Set("after", after)
	}

	if opts != nil {
		if opts.IncidentID != "" {
			params.Set("incident_id", opts.IncidentID)
		}
		for _, status := range opts.Status {
			params.Add("status", status)
		}
	}

	respBody, err := c.doRequest(ctx, "GET", "/actions", params, nil)
	if err != nil {
		return nil, err
	}

	var response ListActionsResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// GetAction retrieves a specific action by ID
func (c *Client) GetAction(ctx context.Context, id string) (*Action, error) {
	respBody, err := c.doRequest(ctx, "GET", fmt.Sprintf("/actions/%s", id), nil, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Action Action `json:"action"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response.Action, nil
}
