package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// Using IncidentType from types.go

// ListIncidentTypesResponse represents the response from listing incident types
type ListIncidentTypesResponse struct {
	IncidentTypes []IncidentType `json:"incident_types"`
}

// ListIncidentTypes returns all incident types
func (c *Client) ListIncidentTypes(ctx context.Context) (*ListIncidentTypesResponse, error) {
	respBody, err := c.doRequest(ctx, "GET", "/incident_types", nil, nil)
	if err != nil {
		return nil, err
	}

	var response ListIncidentTypesResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}
