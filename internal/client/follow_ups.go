package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// ListFollowUpsOptions represents options for listing follow-ups
type ListFollowUpsOptions struct {
	IncidentID   string
	IncidentMode string
}

// ListFollowUpsResponse represents the response from listing follow-ups
type ListFollowUpsResponse struct {
	FollowUps []FollowUp `json:"follow_ups"`
}

// ListFollowUps retrieves all follow-ups for an organization
func (c *Client) ListFollowUps(ctx context.Context, opts *ListFollowUpsOptions) (*ListFollowUpsResponse, error) {
	params := url.Values{}

	if opts != nil {
		if opts.IncidentID != "" {
			params.Set("incident_id", opts.IncidentID)
		}
		if opts.IncidentMode != "" {
			params.Set("incident_mode", opts.IncidentMode)
		}
	}

	respBody, err := c.doRequest(ctx, "GET", "/follow_ups", params, nil)
	if err != nil {
		return nil, err
	}

	var response ListFollowUpsResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// GetFollowUp retrieves a specific follow-up by ID
func (c *Client) GetFollowUp(ctx context.Context, id string) (*FollowUp, error) {
	respBody, err := c.doRequest(ctx, "GET", fmt.Sprintf("/follow_ups/%s", id), nil, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		FollowUp FollowUp `json:"follow_up"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response.FollowUp, nil
}
