package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/incident-io/incidentio-mcp-golang/internal/client"
)

// ListIncidentRolesTool lists available incident roles
type ListIncidentRolesTool struct {
	apiClient *client.Client
}

func NewListIncidentRolesTool(c *client.Client) *ListIncidentRolesTool {
	return &ListIncidentRolesTool{apiClient: c}
}

func (t *ListIncidentRolesTool) Name() string {
	return "list_available_incident_roles"
}

func (t *ListIncidentRolesTool) Description() string {
	return "List all available incident roles that can be assigned to users"
}

func (t *ListIncidentRolesTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"page_size": map[string]interface{}{
				"type":        "integer",
				"description": "Number of results per page (max 250)",
				"default":     25,
			},
		},
	}
}

func (t *ListIncidentRolesTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	opts := &client.ListIncidentRolesOptions{}

	if pageSize, ok := args["page_size"].(float64); ok {
		opts.PageSize = int(pageSize)
	}

	resp, err := t.apiClient.ListIncidentRoles(ctx, opts)
	if err != nil {
		return "", err
	}

	result, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	return string(result), nil
}

// ListUsersTool lists available users for role assignment
type ListUsersTool struct {
	apiClient *client.Client
}

func NewListUsersTool(c *client.Client) *ListUsersTool {
	return &ListUsersTool{apiClient: c}
}

func (t *ListUsersTool) Name() string {
	return "list_users"
}

func (t *ListUsersTool) Description() string {
	return "List all users available for incident role assignment (automatically paginated to return ALL users)"
}

func (t *ListUsersTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"page_size": map[string]interface{}{
				"type":        "integer",
				"description": "Number of results per page (max 250)",
				"default":     250,
			},
			"email": map[string]interface{}{
				"type":        "string",
				"description": "Filter users by email address",
			},
		},
		"additionalProperties": false,
	}
}

func (t *ListUsersTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	opts := &client.ListUsersOptions{}

	if pageSize, ok := args["page_size"].(float64); ok {
		opts.PageSize = int(pageSize)
	}

	if email, ok := args["email"].(string); ok && email != "" {
		opts.Email = email
	}

	resp, err := t.apiClient.ListUsers(ctx, opts)
	if err != nil {
		return "", err
	}

	// Add a helpful message about the results
	var output string
	if opts.Email != "" {
		output = fmt.Sprintf("Users matching email '%s':\n", opts.Email)
	} else {
		output = fmt.Sprintf("Found %d users:\n", len(resp.Users))
	}

	// Format users in a more readable way
	for _, user := range resp.Users {
		output += fmt.Sprintf("\n- Name: %s\n  Email: %s\n  ID: %s\n  Role: %s\n",
			user.Name, user.Email, user.ID, user.Role)
	}

	// Also include the raw JSON
	jsonResult, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return output, nil // Return readable output even if JSON fails
	}

	output += "\n\nRaw JSON response:\n" + string(jsonResult)

	return output, nil
}

// AssignIncidentRoleTool assigns a role to a user for an incident
type AssignIncidentRoleTool struct {
	apiClient *client.Client
}

func NewAssignIncidentRoleTool(c *client.Client) *AssignIncidentRoleTool {
	return &AssignIncidentRoleTool{apiClient: c}
}

func (t *AssignIncidentRoleTool) Name() string {
	return "assign_incident_role"
}

func (t *AssignIncidentRoleTool) Description() string {
	return "Assign a specific incident role to a user for an incident"
}

func (t *AssignIncidentRoleTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "The incident ID",
			},
			"incident_role_id": map[string]interface{}{
				"type":        "string",
				"description": "The incident role ID to assign",
			},
			"user_id": map[string]interface{}{
				"type":        "string",
				"description": "The user ID to assign the role to",
			},
		},
		"required":             []interface{}{"id", "incident_role_id", "user_id"},
		"additionalProperties": false,
	}
}

func (t *AssignIncidentRoleTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	argDetails := make(map[string]interface{})
	for key, value := range args {
		argDetails[key] = value
	}

	if len(args) == 0 {
		return "", fmt.Errorf("no parameters provided")
	}

	incidentID, ok := args["id"].(string)
	if !ok || incidentID == "" {
		return "", fmt.Errorf("id parameter is required and must be a non-empty string. Received parameters: %+v", argDetails)
	}

	roleID, ok := args["incident_role_id"].(string)
	if !ok {
		return "", fmt.Errorf("incident_role_id parameter is required")
	}

	userID, ok := args["user_id"].(string)
	if !ok {
		return "", fmt.Errorf("user_id parameter is required")
	}

	// Create role assignment request using UpdateIncident
	req := &client.UpdateIncidentRequest{
		IncidentRoleAssignments: []client.CreateRoleAssignmentRequest{
			{
				IncidentRoleID: roleID,
				UserID:         userID,
			},
		},
	}

	incident, err := t.apiClient.UpdateIncident(ctx, incidentID, req)
	if err != nil {
		return "", err
	}

	// Return just the role assignments part for clarity
	roleAssignments := make([]map[string]interface{}, 0)
	for _, assignment := range incident.IncidentRoleAssignments {
		roleData := map[string]interface{}{
			"role": map[string]interface{}{
				"id":          assignment.Role.ID,
				"name":        assignment.Role.Name,
				"description": assignment.Role.Description,
				"role_type":   assignment.Role.RoleType,
			},
		}

		if assignment.Assignee != nil {
			roleData["assignee"] = map[string]interface{}{
				"id":    assignment.Assignee.ID,
				"name":  assignment.Assignee.Name,
				"email": assignment.Assignee.Email,
			}
		}

		roleAssignments = append(roleAssignments, roleData)
	}

	response := map[string]interface{}{
		"message":          fmt.Sprintf("Successfully assigned role to user for incident %s", incident.Name),
		"incident_id":      incident.ID,
		"incident_name":    incident.Name,
		"role_assignments": roleAssignments,
	}

	return FormatJSONResponse(response)
}
