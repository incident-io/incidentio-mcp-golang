package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/incident-io/incidentio-mcp-golang/internal/client"
)

// ListAlertSourcesTool lists alert sources from incident.io
type ListAlertSourcesTool struct {
	apiClient *client.Client
}

func NewListAlertSourcesTool(c *client.Client) *ListAlertSourcesTool {
	return &ListAlertSourcesTool{apiClient: c}
}

func (t *ListAlertSourcesTool) Name() string {
	return "list_alert_sources"
}

func (t *ListAlertSourcesTool) Description() string {
	return "List available alert sources in incident.io"
}

func (t *ListAlertSourcesTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"page_size": map[string]interface{}{
				"type":        "integer",
				"description": "Number of results per page",
				"minimum":     1,
				"maximum":     250,
			},
			"after": map[string]interface{}{
				"type":        "string",
				"description": "Pagination cursor for next page",
			},
		},
		"additionalProperties": false,
	}
}

func (t *ListAlertSourcesTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	params := &client.ListAlertSourcesParams{}

	if pageSize, ok := args["page_size"].(float64); ok {
		params.PageSize = int(pageSize)
	}
	if after, ok := args["after"].(string); ok {
		params.After = after
	}

	result, err := t.apiClient.ListAlertSources(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to list alert sources: %w", err)
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}
