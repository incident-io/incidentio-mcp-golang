package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/incident-io/incidentio-mcp-golang/internal/client"
)

// CreateAlertEventTool creates alert events in incident.io
type CreateAlertEventTool struct {
	apiClient *client.Client
}

func NewCreateAlertEventTool(c *client.Client) *CreateAlertEventTool {
	return &CreateAlertEventTool{apiClient: c}
}

func (t *CreateAlertEventTool) Name() string {
	return "create_alert_event"
}

func (t *CreateAlertEventTool) Description() string {
	return "Create an alert event in incident.io"
}

func (t *CreateAlertEventTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"alert_source_id": map[string]interface{}{
				"type":        "string",
				"description": "ID of the alert source to send the event to",
				"minLength":   1,
			},
			"title": map[string]interface{}{
				"type":        "string",
				"description": "Title of the alert event",
				"minLength":   1,
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Description providing more detail about the alert",
			},
			"deduplication_key": map[string]interface{}{
				"type":        "string",
				"description": "Unique key for deduplicating alerts",
			},
			"status": map[string]interface{}{
				"type":        "string",
				"description": "Status of the alert (firing or resolved)",
				"enum":        []string{"firing", "resolved"},
				"default":     "firing",
			},
			"metadata": map[string]interface{}{
				"type":        "object",
				"description": "Additional metadata for the alert",
			},
		},
		"required":             []string{"alert_source_id", "title"},
		"additionalProperties": false,
	}
}

func (t *CreateAlertEventTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	req := &client.CreateAlertEventRequest{}

	alertSourceID, ok := args["alert_source_id"].(string)
	if !ok || alertSourceID == "" {
		return "", fmt.Errorf("alert_source_id is required")
	}
	req.AlertSourceID = alertSourceID

	title, ok := args["title"].(string)
	if !ok || title == "" {
		return "", fmt.Errorf("title is required")
	}
	req.Title = title

	if description, ok := args["description"].(string); ok {
		req.Description = description
	}

	if deduplicationKey, ok := args["deduplication_key"].(string); ok {
		req.DeduplicationKey = deduplicationKey
	}

	if status, ok := args["status"].(string); ok {
		req.Status = status
	} else {
		req.Status = "firing" // default
	}

	if metadata, ok := args["metadata"].(map[string]interface{}); ok {
		req.Metadata = metadata
	}

	alertEvent, err := t.apiClient.CreateAlertEvent(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to create alert event: %w", err)
	}

	output, err := json.MarshalIndent(alertEvent, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}
