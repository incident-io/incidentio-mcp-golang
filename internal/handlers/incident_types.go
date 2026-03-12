package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/incident-io/incidentio-mcp-golang/internal/client"
)

// ListIncidentTypesTool lists available incident types
type ListIncidentTypesTool struct {
	apiClient *client.Client
}

func NewListIncidentTypesTool(c *client.Client) *ListIncidentTypesTool {
	return &ListIncidentTypesTool{apiClient: c}
}

func (t *ListIncidentTypesTool) Name() string {
	return "list_incident_types"
}

func (t *ListIncidentTypesTool) Description() string {
	return "List available incident types"
}

func (t *ListIncidentTypesTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":                 "object",
		"properties":           map[string]interface{}{},
		"additionalProperties": false,
	}
}

func (t *ListIncidentTypesTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	result, err := t.apiClient.ListIncidentTypes(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list incident types: %w", err)
	}

	// Format the output to be more readable
	output := fmt.Sprintf("Found %d incident types:\n\n", len(result.IncidentTypes))

	for _, incidentType := range result.IncidentTypes {
		output += fmt.Sprintf("ID: %s\n", incidentType.ID)
		output += fmt.Sprintf("Name: %s\n", incidentType.Name)
		if incidentType.Description != "" {
			output += fmt.Sprintf("Description: %s\n", incidentType.Description)
		}
		if incidentType.IsDefault {
			output += "Default: Yes\n"
		}
		output += "\n"
	}

	// Also return the raw JSON
	jsonOutput, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return output, nil
	}

	return output + "\nRaw JSON:\n" + string(jsonOutput), nil
}
