package handlers

import (
	"context"
	"fmt"

	"github.com/incident-io/incidentio-mcp-golang/internal/client"
)

// ListIncidentUpdatesTool lists incident updates
type ListIncidentUpdatesTool struct {
	apiClient *client.Client
}

func NewListIncidentUpdatesTool(c *client.Client) *ListIncidentUpdatesTool {
	return &ListIncidentUpdatesTool{apiClient: c}
}

func (t *ListIncidentUpdatesTool) Name() string {
	return "list_incident_updates"
}

func (t *ListIncidentUpdatesTool) Description() string {
	return "List incident updates (status messages posted during an incident).\n\n" +
		"CRITICAL: If user mentions an incident reference like 'INC-1691', you MUST first resolve it to the full incident ID!\n\n" +
		"WORKFLOW FOR INCIDENT REFERENCES:\n" +
		"1. User says: 'updates for INC-1691'\n" +
		"2. FIRST: get_incident({\"incident_id\": \"1691\"}) to get the full incident details\n" +
		"3. Extract the full incident ID from the response (e.g., '01K3VHM0T0ZTMG9JPJ9GESB7XX')\n" +
		"4. THEN: list_incident_updates({\"incident_id\": \"01K3VHM0T0ZTMG9JPJ9GESB7XX\"})\n\n" +
		"IMPORTANT: This endpoint requires the FULL incident ID (01K3VHM0T0ZTMG9JPJ9GESB7XX), NOT the short reference (1691)!\n\n" +
		"EXAMPLES:\n" +
		"- User: 'updates for INC-1691' → get_incident({\"incident_id\": \"1691\"}) → list_incident_updates({\"incident_id\": \"01K3VHM0T0ZTMG9JPJ9GESB7XX\"})"
}

func (t *ListIncidentUpdatesTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"incident_id": map[string]interface{}{
				"type":        "string",
				"description": "Filter updates by incident ID. IMPORTANT: This must be the FULL incident ID (e.g., '01K3VHM0T0ZTMG9JPJ9GESB7XX'), NOT the short reference (e.g., '1691' from INC-1691). Use get_incident first to resolve references.",
			},
			"page_size": map[string]interface{}{
				"type":        "integer",
				"description": "Number of results per page (max 250)",
				"default":     25,
			},
		},
		"additionalProperties": false,
	}
}

func (t *ListIncidentUpdatesTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	opts := &client.ListIncidentUpdatesOptions{
		IncidentID: GetStringArg(args, "incident_id"),
		PageSize:   GetIntArg(args, "page_size", 25),
	}

	resp, err := t.apiClient.ListIncidentUpdates(ctx, opts)
	if err != nil {
		return "", err
	}

	return FormatJSONResponse(resp)
}

// GetIncidentUpdateTool gets a specific incident update
type GetIncidentUpdateTool struct {
	apiClient *client.Client
}

func NewGetIncidentUpdateTool(c *client.Client) *GetIncidentUpdateTool {
	return &GetIncidentUpdateTool{apiClient: c}
}

func (t *GetIncidentUpdateTool) Name() string {
	return "get_incident_update"
}

func (t *GetIncidentUpdateTool) Description() string {
	return "Get details of a specific incident update by ID"
}

func (t *GetIncidentUpdateTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "The incident update ID",
			},
		},
		"required":             []interface{}{"id"},
		"additionalProperties": false,
	}
}

func (t *GetIncidentUpdateTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	id := GetStringArg(args, "id")
	if id == "" {
		return "", fmt.Errorf("id parameter is required")
	}

	update, err := t.apiClient.GetIncidentUpdate(ctx, id)
	if err != nil {
		return "", err
	}

	return FormatJSONResponse(update)
}

// CreateIncidentUpdateTool creates a new incident update
type CreateIncidentUpdateTool struct {
	apiClient *client.Client
}

func NewCreateIncidentUpdateTool(c *client.Client) *CreateIncidentUpdateTool {
	return &CreateIncidentUpdateTool{apiClient: c}
}

func (t *CreateIncidentUpdateTool) Name() string {
	return "create_incident_update"
}

func (t *CreateIncidentUpdateTool) Description() string {
	return "Create a new incident update (status message) for an incident"
}

func (t *CreateIncidentUpdateTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"incident_id": map[string]interface{}{
				"type":        "string",
				"description": "The incident ID to post the update to",
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "The update message to post",
			},
		},
		"required":             []interface{}{"incident_id", "message"},
		"additionalProperties": false,
	}
}

func (t *CreateIncidentUpdateTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	incidentID, ok := args["incident_id"].(string)
	if !ok || incidentID == "" {
		return "", fmt.Errorf("incident_id parameter is required")
	}

	message, ok := args["message"].(string)
	if !ok || message == "" {
		return "", fmt.Errorf("message parameter is required")
	}

	req := &client.CreateIncidentUpdateRequest{
		IncidentID: incidentID,
		Message:    message,
	}

	update, err := t.apiClient.CreateIncidentUpdate(ctx, req)
	if err != nil {
		return "", err
	}

	return FormatJSONResponse(update)
}

// DeleteIncidentUpdateTool deletes an incident update
type DeleteIncidentUpdateTool struct {
	apiClient *client.Client
}

func NewDeleteIncidentUpdateTool(c *client.Client) *DeleteIncidentUpdateTool {
	return &DeleteIncidentUpdateTool{apiClient: c}
}

func (t *DeleteIncidentUpdateTool) Name() string {
	return "delete_incident_update"
}

func (t *DeleteIncidentUpdateTool) Description() string {
	return "Delete an incident update"
}

func (t *DeleteIncidentUpdateTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "The incident update ID to delete",
			},
		},
		"required":             []interface{}{"id"},
		"additionalProperties": false,
	}
}

func (t *DeleteIncidentUpdateTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("id parameter is required")
	}

	if err := t.apiClient.DeleteIncidentUpdate(ctx, id); err != nil {
		return "", err
	}

	return fmt.Sprintf("Successfully deleted incident update %s", id), nil
}
