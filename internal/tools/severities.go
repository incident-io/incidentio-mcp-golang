package tools

import (
	"encoding/json"
	"fmt"

	"github.com/incident-io/incidentio-mcp-golang/internal/incidentio"
)

// ListSeveritiesTool lists available severities
type ListSeveritiesTool struct {
	client *incidentio.Client
}

func NewListSeveritiesTool(client *incidentio.Client) *ListSeveritiesTool {
	return &ListSeveritiesTool{client: client}
}

func (t *ListSeveritiesTool) Name() string {
	return "list_severities"
}

func (t *ListSeveritiesTool) Description() string {
	return "List available severity levels"
}

func (t *ListSeveritiesTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":                 "object",
		"properties":           map[string]interface{}{},
		"additionalProperties": false,
	}
}

func (t *ListSeveritiesTool) Execute(args map[string]interface{}) (string, error) {
	result, err := t.client.ListSeverities()
	if err != nil {
		return "", fmt.Errorf("failed to list severities: %w", err)
	}

	// Format the output to be more readable
	output := fmt.Sprintf("Found %d severity levels:\n\n", len(result.Severities))

	for _, severity := range result.Severities {
		output += fmt.Sprintf("ID: %s\n", severity.ID)
		output += fmt.Sprintf("Name: %s\n", severity.Name)
		if severity.Description != "" {
			output += fmt.Sprintf("Description: %s\n", severity.Description)
		}
		output += fmt.Sprintf("Rank: %d\n", severity.Rank)
		output += "\n"
	}

	// Also return the raw JSON
	jsonOutput, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return output, nil
	}

	return output + "\nRaw JSON:\n" + string(jsonOutput), nil
}

// GetSeverityTool gets a specific severity by ID
type GetSeverityTool struct {
	client *incidentio.Client
}

func NewGetSeverityTool(client *incidentio.Client) *GetSeverityTool {
	return &GetSeverityTool{client: client}
}

func (t *GetSeverityTool) Name() string {
	return "get_severity"
}

func (t *GetSeverityTool) Description() string {
	return "Get details of a specific severity level by ID"
}

func (t *GetSeverityTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "The severity ID",
			},
		},
		"required":             []interface{}{"id"},
		"additionalProperties": false,
	}
}

func (t *GetSeverityTool) Execute(args map[string]interface{}) (string, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("id parameter is required")
	}

	severity, err := t.client.GetSeverity(id)
	if err != nil {
		return "", fmt.Errorf("failed to get severity: %w", err)
	}

	output := "Severity Details:\n\n"
	output += fmt.Sprintf("ID: %s\n", severity.ID)
	output += fmt.Sprintf("Name: %s\n", severity.Name)
	if severity.Description != "" {
		output += fmt.Sprintf("Description: %s\n", severity.Description)
	}
	output += fmt.Sprintf("Rank: %d\n", severity.Rank)
	output += fmt.Sprintf("Created: %s\n", severity.CreatedAt.Format("2006-01-02 15:04:05"))
	output += fmt.Sprintf("Updated: %s\n", severity.UpdatedAt.Format("2006-01-02 15:04:05"))

	// Also return the raw JSON
	jsonOutput, err := json.MarshalIndent(severity, "", "  ")
	if err != nil {
		return output, nil
	}

	return output + "\nRaw JSON:\n" + string(jsonOutput), nil
}
