package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// CreateAlertEventRequest represents a request to create an alert event
type CreateAlertEventRequest struct {
	AlertSourceID    string                 `json:"alert_source_id"`
	DeduplicationKey string                 `json:"deduplication_key,omitempty"`
	Title            string                 `json:"title"`
	Description      string                 `json:"description,omitempty"`
	Status           string                 `json:"status,omitempty"` // "firing" or "resolved"
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// CreateAlertEvent creates a new alert event
func (c *Client) CreateAlertEvent(ctx context.Context, req *CreateAlertEventRequest) (*AlertEvent, error) {
	endpoint := "/alert_events/http"

	respBody, err := c.doRequest(ctx, "POST", endpoint, nil, req)
	if err != nil {
		return nil, err
	}

	var result struct {
		AlertEvent AlertEvent `json:"alert_event"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result.AlertEvent, nil
}
