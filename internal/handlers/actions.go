package handlers

import (
	"context"
	"fmt"

	"github.com/incident-io/incidentio-mcp-golang/internal/client"
)

// ListActionsTool lists actions from incident.io
type ListActionsTool struct {
	apiClient *client.Client
}

func NewListActionsTool(c *client.Client) *ListActionsTool {
	return &ListActionsTool{apiClient: c}
}

func (t *ListActionsTool) Name() string {
	return "list_actions"
}

func (t *ListActionsTool) Description() string {
	return "List actions from incident.io with optional filters. Returns paginated results.\n\n" +
		"PAGINATION: Start with page_size=25. If pagination_meta.after exists, call again with after parameter to get next page. Continue until pagination_meta.after is empty."
}

func (t *ListActionsTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"page_size": map[string]interface{}{
				"type":        "integer",
				"description": "Number of results per page (1-250). Default is 10. Start small to avoid exceeding data limits.",
				"default":     10,
				"minimum":     1,
				"maximum":     250,
			},
			"after": map[string]interface{}{
				"type":        "string",
				"description": "Pagination cursor from previous response's 'pagination_meta.after' field. Use this to fetch the next page of results.",
			},
			"incident_id": map[string]interface{}{
				"type":        "string",
				"description": "Filter actions by incident ID",
			},
			"status": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Filter by action status (outstanding, completed, deleted)",
			},
		},
	}
}

func (t *ListActionsTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	opts := &client.ListActionsOptions{
		PageSize: GetIntArg(args, "page_size", 10), // Default to small page size to avoid exceeding Claude's 1MB limit
		After:    GetStringArg(args, "after"),
	}

	if incidentID := GetStringArg(args, "incident_id"); incidentID != "" {
		opts.IncidentID = incidentID
	}

	opts.Status = GetStringArrayArg(args, "status")

	resp, err := t.apiClient.ListActions(ctx, opts)
	if err != nil {
		return "", err
	}

	// Create response with prominent pagination info
	response := map[string]interface{}{
		"actions":         resp.Actions,
		"pagination_meta": resp.PaginationMeta,
		"count":           len(resp.Actions),
	}

	// Add helpful message if there are more pages
	if resp.PaginationMeta.After != "" {
		response["next_page_hint"] = fmt.Sprintf("More results available. Use after='%s' to fetch the next page.", resp.PaginationMeta.After)
	}

	return FormatJSONResponse(response)
}

// GetActionTool retrieves a specific action
type GetActionTool struct {
	apiClient *client.Client
}

func NewGetActionTool(c *client.Client) *GetActionTool {
	return &GetActionTool{apiClient: c}
}

func (t *GetActionTool) Name() string {
	return "get_action"
}

func (t *GetActionTool) Description() string {
	return "Get details of a specific action by ID"
}

func (t *GetActionTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "The action ID",
			},
		},
		"required": []string{"id"},
	}
}

func (t *GetActionTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	id := GetStringArg(args, "id")
	if id == "" {
		return "", fmt.Errorf("id parameter is required")
	}

	action, err := t.apiClient.GetAction(ctx, id)
	if err != nil {
		return "", err
	}

	return FormatJSONResponse(action)
}
