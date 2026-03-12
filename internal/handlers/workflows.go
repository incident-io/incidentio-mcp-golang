package handlers

import (
	"context"
	"fmt"

	"github.com/incident-io/incidentio-mcp-golang/internal/client"
)

// ListWorkflowsTool lists workflows from incident.io
type ListWorkflowsTool struct {
	apiClient *client.Client
}

func NewListWorkflowsTool(c *client.Client) *ListWorkflowsTool {
	return &ListWorkflowsTool{apiClient: c}
}

func (t *ListWorkflowsTool) Name() string {
	return "list_workflows"
}

func (t *ListWorkflowsTool) Description() string {
	return "List workflows from incident.io with optional pagination"
}

func (t *ListWorkflowsTool) InputSchema() map[string]interface{} {
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

func (t *ListWorkflowsTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	params := &client.ListWorkflowsParams{
		PageSize: GetIntArg(args, "page_size", 25),
		After:    GetStringArg(args, "after"),
	}

	result, err := t.apiClient.ListWorkflows(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to list workflows: %w", err)
	}

	return FormatJSONResponse(result)
}

// GetWorkflowTool gets details of a specific workflow
type GetWorkflowTool struct {
	apiClient *client.Client
}

func NewGetWorkflowTool(c *client.Client) *GetWorkflowTool {
	return &GetWorkflowTool{apiClient: c}
}

func (t *GetWorkflowTool) Name() string {
	return "get_workflow"
}

func (t *GetWorkflowTool) Description() string {
	return "Get details of a specific workflow by ID"
}

func (t *GetWorkflowTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "The workflow ID",
				"minLength":   1,
			},
		},
		"required":             []string{"id"},
		"additionalProperties": false,
	}
}

func (t *GetWorkflowTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	id := GetStringArg(args, "id")
	if id == "" {
		return "", fmt.Errorf("workflow ID is required")
	}

	workflow, err := t.apiClient.GetWorkflow(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to get workflow: %w", err)
	}

	return FormatJSONResponse(workflow)
}

// UpdateWorkflowTool updates a workflow
type UpdateWorkflowTool struct {
	apiClient *client.Client
}

func NewUpdateWorkflowTool(c *client.Client) *UpdateWorkflowTool {
	return &UpdateWorkflowTool{apiClient: c}
}

func (t *UpdateWorkflowTool) Name() string {
	return "update_workflow"
}

func (t *UpdateWorkflowTool) Description() string {
	return "Update a workflow's configuration"
}

func (t *UpdateWorkflowTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "The workflow ID to update",
				"minLength":   1,
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "New name for the workflow",
			},
			"enabled": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether the workflow should be enabled",
			},
			"state": map[string]interface{}{
				"type":        "object",
				"description": "State configuration for the workflow",
			},
		},
		"required":             []string{"id"},
		"additionalProperties": false,
	}
}

func (t *UpdateWorkflowTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("workflow ID is required")
	}

	req := &client.UpdateWorkflowRequest{}

	if name, ok := args["name"].(string); ok {
		req.Name = name
	}

	if enabled, ok := args["enabled"].(bool); ok {
		req.Enabled = &enabled
	}

	if state, ok := args["state"].(map[string]interface{}); ok {
		req.State = state
	}

	workflow, err := t.apiClient.UpdateWorkflow(ctx, id, req)
	if err != nil {
		return "", fmt.Errorf("failed to update workflow: %w", err)
	}

	return FormatJSONResponse(workflow)
}
