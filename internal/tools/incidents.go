package tools

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/incident-io/incidentio-mcp-golang/internal/incidentio"
)

// ListIncidentsTool lists incidents from incident.io
type ListIncidentsTool struct {
	client *incidentio.Client
}

func NewListIncidentsTool(client *incidentio.Client) *ListIncidentsTool {
	return &ListIncidentsTool{client: client}
}

func (t *ListIncidentsTool) Name() string {
	return "list_incidents"
}

func (t *ListIncidentsTool) Description() string {
	return `List incidents from incident.io with optional filtering by status and severity.

USAGE WORKFLOW:
1. To filter by severity, first call 'list_severities' to get available severity IDs
2. Use the severity ID (not the name) in the 'severity' parameter
3. Multiple severity IDs can be provided to match any of them (OR logic)
4. Status filters can be combined with severity filters

PARAMETERS:
- page_size: Number of results (default 25, max 250). Set to 0 or omit for auto-pagination.
- status: Array of status values (triage, active, resolved, closed)
- severity: Array of severity IDs (e.g., ["01HXYZ..."]) - Use list_severities first to get valid IDs

EXAMPLES:
- List all active incidents: {"status": ["active"]}
- List critical incidents: First call list_severities, then use severity ID like {"severity": ["01HXYZ..."]}
- List active high-severity incidents: {"status": ["active"], "severity": ["sev_1", "sev_2"]}

IMPORTANT: Severity parameter requires severity IDs, not severity names. Always call list_severities first to discover available severity IDs.`
}

func (t *ListIncidentsTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"page_size": map[string]interface{}{
				"type":        "integer",
				"description": "Number of results per page (max 250, default 25). Set to 0 or omit for automatic pagination through all results.",
				"default":     25,
			},
			"status": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Filter by incident status values. Valid values: triage, active, resolved, closed. Multiple values will match any of them (OR logic). Example: [\"active\", \"triage\"]",
			},
			"severity": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Filter by severity IDs (NOT severity names). IMPORTANT: Call 'list_severities' tool first to discover available severity IDs. Example: [\"sev_1\", \"sev_2\"] or [\"01HXYZ...\"]. Multiple IDs will match any of them (OR logic).",
			},
		},
	}
}

func (t *ListIncidentsTool) Execute(args map[string]interface{}) (string, error) {
	opts := &incidentio.ListIncidentsOptions{}

	if pageSize, ok := args["page_size"].(float64); ok {
		opts.PageSize = int(pageSize)
	}

	if statuses, ok := args["status"].([]interface{}); ok {
		for _, s := range statuses {
			if str, ok := s.(string); ok {
				opts.Status = append(opts.Status, str)
			}
		}
	}

	if severities, ok := args["severity"].([]interface{}); ok {
		for _, s := range severities {
			if str, ok := s.(string); ok {
				opts.Severity = append(opts.Severity, str)
			}
		}
	}

	resp, err := t.client.ListIncidents(opts)
	if err != nil {
		return "", err
	}

	result, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	return string(result), nil
}

// GetIncidentTool retrieves a specific incident
type GetIncidentTool struct {
	client *incidentio.Client
}

func NewGetIncidentTool(client *incidentio.Client) *GetIncidentTool {
	return &GetIncidentTool{client: client}
}

func (t *GetIncidentTool) Name() string {
	return "get_incident"
}

func (t *GetIncidentTool) Description() string {
	return "Get details of a specific incident by ID"
}

func (t *GetIncidentTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"incident_id": map[string]interface{}{
				"type":        "string",
				"description": "The incident ID",
			},
		},
		"required":             []interface{}{"incident_id"},
		"additionalProperties": false,
	}
}

func (t *GetIncidentTool) Execute(args map[string]interface{}) (string, error) {
	id, ok := args["incident_id"].(string)
	if !ok || id == "" {
		argDetails := make(map[string]interface{})
		for key, value := range args {
			argDetails[key] = value
		}
		return "", fmt.Errorf("incident_id parameter is required and must be a non-empty string. Received parameters: %+v", argDetails)
	}

	incident, err := t.client.GetIncident(id)
	if err != nil {
		return "", err
	}

	result, err := json.MarshalIndent(incident, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	return string(result), nil
}

// CreateIncidentTool creates a new incident
type CreateIncidentTool struct {
	client *incidentio.Client
}

func NewCreateIncidentTool(client *incidentio.Client) *CreateIncidentTool {
	return &CreateIncidentTool{client: client}
}

func (t *CreateIncidentTool) Name() string {
	return "create_incident"
}

func (t *CreateIncidentTool) Description() string {
	return "Create a new incident in incident.io"
}

func (t *CreateIncidentTool) InputSchema() map[string]interface{} {
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
			"status": map[string]interface{}{
				"type":        "string",
				"description": "Initial status (triage, active, resolved, closed)",
				"default":     "triage",
			},
			"severity_id": map[string]interface{}{
				"type":        "string",
				"description": "The severity ID",
			},
			"incident_type_id": map[string]interface{}{
				"type":        "string",
				"description": "The incident type ID",
			},
			"incident_status_id": map[string]interface{}{
				"type":        "string",
				"description": "The incident status ID",
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

func (t *CreateIncidentTool) Execute(args map[string]interface{}) (string, error) {
	name, ok := args["name"].(string)
	if !ok {
		return "", fmt.Errorf("name parameter is required")
	}

	// Generate idempotency key using timestamp and name
	idempotencyKey := fmt.Sprintf("mcp-%d-%s", time.Now().UnixNano(), name)

	req := &incidentio.CreateIncidentRequest{
		IdempotencyKey: idempotencyKey,
		Name:           name,
		Mode:           "standard", // Default to standard mode
		Visibility:     "public",   // Default to public visibility
	}

	if summary, ok := args["summary"].(string); ok {
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

	// Check if critical fields are missing and provide helpful suggestions
	var suggestions []string

	if req.SeverityID == "" {
		suggestions = append(suggestions, "severity_id is not set. Use list_severities to see available options.")
	}

	if req.IncidentTypeID == "" {
		suggestions = append(suggestions, "incident_type_id is not set. Use list_incident_types to see available options.")
	}

	if req.IncidentStatusID == "" {
		suggestions = append(suggestions, "incident_status_id is not set. Use list_incident_statuses to see available options.")
	}

	incident, err := t.client.CreateIncident(req)
	if err != nil {
		// If the error is related to missing required fields, provide more helpful error message
		errMsg := err.Error()
		if len(suggestions) > 0 && (strings.Contains(errMsg, "severity") || strings.Contains(errMsg, "incident_type") || strings.Contains(errMsg, "incident_status")) {
			return "", fmt.Errorf("%s\n\nSuggestions:\n%s", errMsg, strings.Join(suggestions, "\n"))
		}
		return "", err
	}

	// Include suggestions in successful response if fields were missing
	result, err := json.MarshalIndent(incident, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	if len(suggestions) > 0 {
		return fmt.Sprintf("%s\n\nNote: Incident created with defaults. %s", result, strings.Join(suggestions, " ")), nil
	}

	return string(result), nil
}

// UpdateIncidentTool updates an existing incident
type UpdateIncidentTool struct {
	client *incidentio.Client
}

func NewUpdateIncidentTool(client *incidentio.Client) *UpdateIncidentTool {
	return &UpdateIncidentTool{client: client}
}

func (t *UpdateIncidentTool) Name() string {
	return "update_incident"
}

func (t *UpdateIncidentTool) Description() string {
	return "Update an existing incident"
}

func (t *UpdateIncidentTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"incident_id": map[string]interface{}{
				"type":        "string",
				"description": "The incident ID to update",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Update the incident name",
			},
			"summary": map[string]interface{}{
				"type":        "string",
				"description": "Update the incident summary",
			},
			"incident_status_id": map[string]interface{}{
				"type":        "string",
				"description": "Update the incident status ID",
			},
			"severity_id": map[string]interface{}{
				"type":        "string",
				"description": "Update the severity ID",
			},
		},
		"required":             []interface{}{"incident_id"},
		"additionalProperties": false,
	}
}

func (t *UpdateIncidentTool) Execute(args map[string]interface{}) (string, error) {

	id, ok := args["incident_id"].(string)
	if !ok || id == "" {
		argDetails := make(map[string]interface{})
		for key, value := range args {
			argDetails[key] = value
		}
		return "", fmt.Errorf("incident_id parameter is required and must be a non-empty string. Received parameters: %+v", argDetails)
	}

	req := &incidentio.UpdateIncidentRequest{}
	hasUpdate := false

	if name, ok := args["name"].(string); ok {
		req.Name = name
		hasUpdate = true
	}
	if summary, ok := args["summary"].(string); ok {
		req.Summary = summary
		hasUpdate = true
	}
	if statusID, ok := args["incident_status_id"].(string); ok {
		req.IncidentStatusID = statusID
		hasUpdate = true
	}
	if severityID, ok := args["severity_id"].(string); ok {
		req.SeverityID = severityID
		hasUpdate = true
	}

	if !hasUpdate {
		return "", fmt.Errorf("at least one field to update must be provided")
	}

	incident, err := t.client.UpdateIncident(id, req)
	if err != nil {
		return "", err
	}

	result, err := json.MarshalIndent(incident, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	return string(result), nil
}
