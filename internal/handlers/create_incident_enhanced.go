package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/incident-io/incidentio-mcp-golang/internal/client"
)

// CreateIncidentEnhancedTool creates a new incident with smart defaults
type CreateIncidentEnhancedTool struct {
	apiClient *client.Client
}

func NewCreateIncidentEnhancedTool(c *client.Client) *CreateIncidentEnhancedTool {
	return &CreateIncidentEnhancedTool{apiClient: c}
}

func (t *CreateIncidentEnhancedTool) Name() string {
	return "create_incident_smart"
}

func (t *CreateIncidentEnhancedTool) Description() string {
	return "Create a new incident with smart defaults - automatically fetches first available severity, type, and status if not provided"
}

func (t *CreateIncidentEnhancedTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "The incident name/title",
			},
			"summary": map[string]interface{}{
				"type":        "string",
				"description": "A summary of the incident",
			},
			"severity_id": map[string]interface{}{
				"type":        "string",
				"description": "The severity ID (auto-selected if not provided)",
			},
			"incident_type_id": map[string]interface{}{
				"type":        "string",
				"description": "The incident type ID (auto-selected if not provided)",
			},
			"incident_status_id": map[string]interface{}{
				"type":        "string",
				"description": "The incident status ID (auto-selected if not provided)",
			},
			"mode": map[string]interface{}{
				"type":        "string",
				"description": "The incident mode (standard, retrospective, tutorial)",
				"enum":        []string{"standard", "retrospective", "tutorial"},
				"default":     "standard",
			},
			"visibility": map[string]interface{}{
				"type":        "string",
				"description": "The incident visibility (public, private)",
				"enum":        []string{"public", "private"},
				"default":     "public",
			},
			"slack_channel_name_override": map[string]interface{}{
				"type":        "string",
				"description": "Override the auto-generated Slack channel name",
			},
		},
		"required":             []interface{}{"name"},
		"additionalProperties": false,
	}
}

func (t *CreateIncidentEnhancedTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	name, ok := args["name"].(string)
	if !ok {
		return "", fmt.Errorf("name parameter is required")
	}

	// Validate name input
	if err := ValidateStringInput(name, "name", MaxNameLength, true); err != nil {
		return "", err
	}

	// Generate idempotency key using timestamp and name
	idempotencyKey := fmt.Sprintf("mcp-%d-%s", time.Now().UnixNano(), name)

	req := &client.CreateIncidentRequest{
		IdempotencyKey: idempotencyKey,
		Name:           name,
		Mode:           "standard", // Default to standard mode
		Visibility:     "public",   // Default to public visibility
	}

	// Collect auto-fetched defaults info
	var autoDefaults []string

	// Set provided values first
	if summary, ok := args["summary"].(string); ok {
		if err := ValidateStringInput(summary, "summary", MaxSummaryLength, false); err != nil {
			return "", err
		}
		req.Summary = summary
	}
	if statusID, ok := args["incident_status_id"].(string); ok {
		req.IncidentStatusID = statusID
	}
	if severityID, ok := args["severity_id"].(string); ok {
		req.SeverityID = severityID
	}
	if typeID, ok := args["incident_type_id"].(string); ok {
		req.IncidentTypeID = typeID
	}
	if mode, ok := args["mode"].(string); ok {
		req.Mode = mode
	}
	if visibility, ok := args["visibility"].(string); ok {
		req.Visibility = visibility
	}
	if slackOverride, ok := args["slack_channel_name_override"].(string); ok {
		req.SlackChannelNameOverride = slackOverride
	}

	// Auto-fetch severity if not provided
	if req.SeverityID == "" {
		severities, err := t.apiClient.ListSeverities(ctx)
		if err == nil && len(severities.Severities) > 0 {
			// Select the first severity (usually the least severe)
			req.SeverityID = severities.Severities[len(severities.Severities)-1].ID
			autoDefaults = append(autoDefaults, fmt.Sprintf("severity_id auto-selected: %s (%s)",
				severities.Severities[len(severities.Severities)-1].ID,
				severities.Severities[len(severities.Severities)-1].Name))
		}
	}

	// Auto-fetch incident type if not provided
	if req.IncidentTypeID == "" {
		types, err := t.apiClient.ListIncidentTypes(ctx)
		if err == nil && len(types.IncidentTypes) > 0 {
			// Select the first incident type
			req.IncidentTypeID = types.IncidentTypes[0].ID
			autoDefaults = append(autoDefaults, fmt.Sprintf("incident_type_id auto-selected: %s (%s)",
				types.IncidentTypes[0].ID,
				types.IncidentTypes[0].Name))
		}
	}

	// Auto-fetch incident status if not provided using V1 API
	if req.IncidentStatusID == "" {
		respBody, err := t.apiClient.DoRequestV1(ctx, "GET", "/incident_statuses", nil, nil)
		if err == nil {
			var statusResponse struct {
				IncidentStatuses []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"incident_statuses"`
			}

			if json.Unmarshal(respBody, &statusResponse) == nil && len(statusResponse.IncidentStatuses) > 0 {
				// Look for "triage" or "investigating" status first
				for _, status := range statusResponse.IncidentStatuses {
					if strings.ToLower(status.Name) == "triage" || strings.ToLower(status.Name) == "investigating" {
						req.IncidentStatusID = status.ID
						autoDefaults = append(autoDefaults, fmt.Sprintf("incident_status_id auto-selected: %s (%s)",
							status.ID, status.Name))
						break
					}
				}
				// If no triage status found, use the first one
				if req.IncidentStatusID == "" {
					req.IncidentStatusID = statusResponse.IncidentStatuses[0].ID
					autoDefaults = append(autoDefaults, fmt.Sprintf("incident_status_id auto-selected: %s (%s)",
						statusResponse.IncidentStatuses[0].ID,
						statusResponse.IncidentStatuses[0].Name))
				}
			}
		}
	}

	// Create the incident
	incident, err := t.apiClient.CreateIncident(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to create incident: %w", err)
	}

	// Format the response
	result, err := json.MarshalIndent(incident, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	// Add auto-defaults information if any
	if len(autoDefaults) > 0 {
		return fmt.Sprintf("%s\n\nAuto-selected defaults:\n%s", result, strings.Join(autoDefaults, "\n")), nil
	}

	return string(result), nil
}
