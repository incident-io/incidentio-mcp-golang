package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// ListAlertsOptions represents options for listing alerts
type ListAlertsOptions struct {
	PageSize           int
	After              string
	Status             []string
	DeduplicationKey   string // Filter by deduplication key
	CreatedAtGte       string // Filter alerts created on or after this date
	CreatedAtLte       string // Filter alerts created on or before this date
	CreatedAtDateRange string // Filter alerts created within date range (format: "2024-12-02~2024-12-08")
}

// ListIncidentAlertsOptions represents options for listing incident alerts
type ListIncidentAlertsOptions struct {
	PageSize   int
	After      string
	AlertID    string
	IncidentID string
}

// ListAlertsResponse represents the response from listing alerts
type ListAlertsResponse struct {
	Alerts []Alert `json:"alerts"`
	ListResponse
}

// ListIncidentAlertsResponse represents the response from listing incident alerts
type ListIncidentAlertsResponse struct {
	IncidentAlerts []IncidentAlert `json:"incident_alerts"`
	ListResponse
}

// ListAlerts retrieves a single page of alerts
// Pagination is controlled by the caller using PageSize and After parameters
func (c *Client) ListAlerts(ctx context.Context, opts *ListAlertsOptions) (*ListAlertsResponse, error) {
	pageSize := 25 // API default page size
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
		for _, status := range opts.Status {
			params.Add("status[one_of]", status)
		}

		if opts.DeduplicationKey != "" {
			params.Set("deduplication_key[is]", opts.DeduplicationKey)
		}

		if opts.CreatedAtGte != "" {
			params.Set("created_at[gte]", opts.CreatedAtGte)
		}
		if opts.CreatedAtLte != "" {
			params.Set("created_at[lte]", opts.CreatedAtLte)
		}
		if opts.CreatedAtDateRange != "" {
			params.Set("created_at[date_range]", opts.CreatedAtDateRange)
		}
	}

	respBody, err := c.doRequest(ctx, "GET", "/alerts", params, nil)
	if err != nil {
		return nil, err
	}

	var response ListAlertsResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// GetAlert retrieves a specific alert by ID
func (c *Client) GetAlert(ctx context.Context, id string) (*Alert, error) {
	respBody, err := c.doRequest(ctx, "GET", fmt.Sprintf("/alerts/%s", id), nil, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Alert Alert `json:"alert"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response.Alert, nil
}

// ListIncidentAlerts retrieves a single page of incident alerts
// Pagination is controlled by the caller using PageSize and After parameters
func (c *Client) ListIncidentAlerts(ctx context.Context, opts *ListIncidentAlertsOptions) (*ListIncidentAlertsResponse, error) {
	pageSize := 25 // Default page size as per API documentation
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
		if opts.AlertID != "" {
			params.Set("alert_id", opts.AlertID)
		}
		if opts.IncidentID != "" {
			params.Set("incident_id", opts.IncidentID)
		}
	}

	respBody, err := c.doRequest(ctx, "GET", "/incident_alerts", params, nil)
	if err != nil {
		return nil, err
	}

	var response ListIncidentAlertsResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}
