package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// ListAlertRoutesParams contains optional parameters for listing alert routes
type ListAlertRoutesParams struct {
	PageSize int
	After    string
}

// ListAlertRoutesResponse represents the response from listing alert routes
type ListAlertRoutesResponse struct {
	AlertRoutes []AlertRoute `json:"alert_routes"`
	Pagination  struct {
		After    string `json:"after,omitempty"`
		PageSize int    `json:"page_size"`
	} `json:"pagination_info"`
}

// ListAlertRoutes returns all alert routes
func (c *Client) ListAlertRoutes(ctx context.Context, params *ListAlertRoutesParams) (*ListAlertRoutesResponse, error) {
	endpoint := "/alert_routes"

	// Set default page size
	pageSize := 25
	if params != nil && params.PageSize > 0 {
		pageSize = params.PageSize
	}

	v := url.Values{}
	v.Set("page_size", fmt.Sprintf("%d", pageSize)) // Always set (may be required)
	if params != nil && params.After != "" {
		v.Set("after", params.After)
	}

	if len(v) > 0 {
		endpoint = endpoint + "?" + v.Encode()
	}

	respBody, err := c.doRequest(ctx, "GET", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var result ListAlertRoutesResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// GetAlertRoute returns a specific alert route by ID
func (c *Client) GetAlertRoute(ctx context.Context, id string) (*AlertRoute, error) {
	endpoint := fmt.Sprintf("/alert_routes/%s", id)

	respBody, err := c.doRequest(ctx, "GET", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		AlertRoute AlertRoute `json:"alert_route"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result.AlertRoute, nil
}

// CreateAlertRouteRequest represents a request to create an alert route
type CreateAlertRouteRequest struct {
	Name         string                 `json:"name"`
	Enabled      bool                   `json:"enabled"`
	Conditions   []AlertCondition       `json:"conditions"`
	Escalations  []EscalationBinding    `json:"escalations"`
	GroupingKeys []string               `json:"grouping_keys,omitempty"`
	Template     map[string]interface{} `json:"template,omitempty"`
}

// CreateAlertRoute creates a new alert route
func (c *Client) CreateAlertRoute(ctx context.Context, req *CreateAlertRouteRequest) (*AlertRoute, error) {
	endpoint := "/alert_routes"

	respBody, err := c.doRequest(ctx, "POST", endpoint, nil, req)
	if err != nil {
		return nil, err
	}

	var result struct {
		AlertRoute AlertRoute `json:"alert_route"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result.AlertRoute, nil
}

// UpdateAlertRouteRequest represents a request to update an alert route
type UpdateAlertRouteRequest struct {
	Name         string                 `json:"name,omitempty"`
	Enabled      *bool                  `json:"enabled,omitempty"`
	Conditions   []AlertCondition       `json:"conditions,omitempty"`
	Escalations  []EscalationBinding    `json:"escalations,omitempty"`
	GroupingKeys []string               `json:"grouping_keys,omitempty"`
	Template     map[string]interface{} `json:"template,omitempty"`
}

// UpdateAlertRoute updates an alert route
func (c *Client) UpdateAlertRoute(ctx context.Context, id string, req *UpdateAlertRouteRequest) (*AlertRoute, error) {
	endpoint := fmt.Sprintf("/alert_routes/%s", id)

	respBody, err := c.doRequest(ctx, "PATCH", endpoint, nil, req)
	if err != nil {
		return nil, err
	}

	var result struct {
		AlertRoute AlertRoute `json:"alert_route"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result.AlertRoute, nil
}
