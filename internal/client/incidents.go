package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// ListIncidentsOptions represents options for listing incidents
// Per API docs: https://api-docs.incident.io/tag/Incidents-V2#operation/Incidents%20V2_List
type ListIncidentsOptions struct {
	PageSize           int
	After              string
	Status             []string
	Severity           []string // Deprecated: Use SeverityOneOf, SeverityGte, or SeverityLte
	SeverityOneOf      []string // Filter by exact severity IDs (severity[one_of]=ID)
	SeverityGte        string   // Filter by severity rank >= this ID (severity[gte]=ID)
	SeverityLte        string   // Filter by severity rank <= this ID (severity[lte]=ID)
	IncidentTypeOneOf  []string // Filter by incident type IDs (incident_type[one_of]=ID)
	CreatedAtGte       string   // Greater than or equal to (format: "2024-12-02")
	CreatedAtLte       string   // Less than or equal to (format: "2024-12-02")
	CreatedAtDateRange string   // Date range (format: "2024-12-02~2024-12-08")
	UpdatedAtGte       string   // Greater than or equal to
	UpdatedAtLte       string   // Less than or equal to
	UpdatedAtDateRange string   // Date range (format: "2024-12-02~2024-12-08")
	CustomFieldOneOf   map[string]string // custom_field[ID][one_of]=option_id
	CustomFieldNotIn   map[string]string // custom_field[ID][not_in]=option_id
}

// ListIncidentsResponse represents the response from listing incidents
type ListIncidentsResponse struct {
	Incidents []Incident `json:"incidents"`
	ListResponse
}

// ListIncidents retrieves a single page of incidents
// Pagination is controlled by the caller using PageSize and After parameters
func (c *Client) ListIncidents(opts *ListIncidentsOptions) (*ListIncidentsResponse, error) {
	pageSize := 10 // Conservative default to avoid exceeding MCP client limits
	after := ""

	if opts != nil {
		if opts.PageSize > 0 {
			pageSize = opts.PageSize
		}
		if opts.After != "" {
			after = opts.After
		}
	}

	params := url.Values{}
	params.Set("page_size", strconv.Itoa(pageSize))

	if after != "" {
		params.Set("after", after)
	}

	if opts != nil {
		for _, status := range opts.Status {
			params.Add("status[one_of]", status)
		}
		for _, severity := range opts.SeverityOneOf {
			params.Add("severity[one_of]", severity)
		}
		if opts.SeverityGte != "" {
			params.Add("severity[gte]", opts.SeverityGte)
		}
		if opts.SeverityLte != "" {
			params.Add("severity[lte]", opts.SeverityLte)
		}
		for _, typeID := range opts.IncidentTypeOneOf {
			params.Add("incident_type[one_of]", typeID)
		}
		if opts.CreatedAtGte != "" {
			params.Add("created_at[gte]", opts.CreatedAtGte)
		}
		if opts.CreatedAtLte != "" {
			params.Add("created_at[lte]", opts.CreatedAtLte)
		}
		if opts.CreatedAtDateRange != "" {
			params.Add("created_at[date_range]", opts.CreatedAtDateRange)
		}
		if opts.UpdatedAtGte != "" {
			params.Add("updated_at[gte]", opts.UpdatedAtGte)
		}
		if opts.UpdatedAtLte != "" {
			params.Add("updated_at[lte]", opts.UpdatedAtLte)
		}
		if opts.UpdatedAtDateRange != "" {
			params.Add("updated_at[date_range]", opts.UpdatedAtDateRange)
		}
		// Add custom field filters with correct format: custom_field[ID][one_of]=value
		for fieldID, value := range opts.CustomFieldOneOf {
			params.Add("custom_field["+fieldID+"][one_of]", value)
		}
		for fieldID, value := range opts.CustomFieldNotIn {
			params.Add("custom_field["+fieldID+"][not_in]", value)
		}
	}

	respBody, err := c.doRequest("GET", "/incidents", params, nil)
	if err != nil {
		return nil, err
	}

	var response ListIncidentsResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// GetIncident retrieves a specific incident by ID
func (c *Client) GetIncident(id string) (*Incident, error) {
	respBody, err := c.doRequest("GET", fmt.Sprintf("/incidents/%s", id), nil, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Incident Incident `json:"incident"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response.Incident, nil
}

// CreateIncident creates a new incident
func (c *Client) CreateIncident(req *CreateIncidentRequest) (*Incident, error) {
	respBody, err := c.doRequest("POST", "/incidents", nil, req)
	if err != nil {
		return nil, err
	}

	var response struct {
		Incident Incident `json:"incident"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response.Incident, nil
}

// UpdateIncident updates an existing incident using V2 actions/edit API
func (c *Client) UpdateIncident(id string, req *UpdateIncidentRequest) (*Incident, error) {
	// Use the correct V2 actions/edit endpoint
	editRequest := map[string]interface{}{
		"notify_incident_channel": true,
	}

	// Build the incident object with only the fields that are being updated
	incident := make(map[string]interface{})

	if req.Name != "" {
		incident["name"] = req.Name
	}
	if req.Summary != "" {
		incident["summary"] = req.Summary
	}
	if req.IncidentStatusID != "" {
		incident["incident_status_id"] = req.IncidentStatusID
	}
	if req.SeverityID != "" {
		incident["severity_id"] = req.SeverityID
	}
	if req.CallURL != "" {
		incident["call_url"] = req.CallURL
	}
	if req.SlackChannelNameOverride != "" {
		incident["slack_channel_name_override"] = req.SlackChannelNameOverride
	}
	if len(req.IncidentRoleAssignments) > 0 {
		incident["incident_role_assignments"] = req.IncidentRoleAssignments
	}
	if len(req.CustomFieldEntries) > 0 {
		incident["custom_field_entries"] = req.CustomFieldEntries
	}
	if len(req.IncidentTimestampValues) > 0 {
		incident["incident_timestamp_values"] = req.IncidentTimestampValues
	}

	// Only include incident object if there are fields to update
	if len(incident) > 0 {
		editRequest["incident"] = incident
	} else {
		return nil, fmt.Errorf("no fields to update")
	}

	respBody, err := c.doRequest("POST", fmt.Sprintf("/incidents/%s/actions/edit", id), nil, editRequest)
	if err != nil {
		return nil, err
	}

	var response struct {
		Incident Incident `json:"incident"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response.Incident, nil
}

// AssignIncidentRoleRequest represents a request to assign a role to a user
type AssignIncidentRoleRequest struct {
	IncidentRoleID string `json:"incident_role_id"`
	UserID         string `json:"user_id"`
}

// AssignIncidentRole assigns a specific role to a user for an incident
func (c *Client) AssignIncidentRole(incidentID string, req *AssignIncidentRoleRequest) (*Incident, error) {
	respBody, err := c.doRequest("PATCH", fmt.Sprintf("/incidents/%s", incidentID), nil, map[string]interface{}{
		"incident_role_assignments": []map[string]interface{}{
			{
				"incident_role_id": req.IncidentRoleID,
				"user_id":          req.UserID,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	var response struct {
		Incident Incident `json:"incident"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response.Incident, nil
}
