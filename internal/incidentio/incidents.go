package incidentio

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// ListIncidentsOptions represents options for listing incidents
type ListIncidentsOptions struct {
	PageSize int
	After    string
	Status   []string
	Severity []string
}

// ListIncidentsResponse represents the response from listing incidents
type ListIncidentsResponse struct {
	Incidents []Incident `json:"incidents"`
	ListResponse
}

// ListIncidents retrieves a list of incidents with automatic pagination
func (c *Client) ListIncidents(opts *ListIncidentsOptions) (*ListIncidentsResponse, error) {
	allIncidents := []Incident{}
	pageSize := 250 // Default max page size
	after := ""

	// If a specific page size is requested, respect it and don't paginate
	if opts != nil && opts.PageSize > 0 {
		params := url.Values{}
		params.Set("page_size", strconv.Itoa(opts.PageSize))

		if opts.After != "" {
			params.Set("after", opts.After)
		}

		for _, status := range opts.Status {
			params.Add("status", status)
		}
		for _, severity := range opts.Severity {
			params.Add("severity[one_of]", severity)
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

	// Set up base parameters for auto-pagination
	baseParams := url.Values{}
	if opts != nil {
		for _, status := range opts.Status {
			baseParams.Add("status", status)
		}
		for _, severity := range opts.Severity {
			baseParams.Add("severity[one_of]", severity)
		}
	}

	// Paginate through all results
	maxPages := 10 // Safety limit
	for page := 0; page < maxPages; page++ {
		params := url.Values{}
		// Copy base parameters
		for k, v := range baseParams {
			params[k] = v
		}

		params.Set("page_size", strconv.Itoa(pageSize))
		if after != "" {
			params.Set("after", after)
		}

		respBody, err := c.doRequest("GET", "/incidents", params, nil)
		if err != nil {
			return nil, err
		}

		var response ListIncidentsResponse
		if err := json.Unmarshal(respBody, &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}

		allIncidents = append(allIncidents, response.Incidents...)

		// Check if there are more pages
		if response.PaginationMeta.After == "" || len(response.Incidents) == 0 {
			break
		}
		after = response.PaginationMeta.After
	}

	// Return combined results
	return &ListIncidentsResponse{
		Incidents: allIncidents,
		ListResponse: ListResponse{
			PaginationMeta: struct {
				After      string `json:"after,omitempty"`
				PageSize   int    `json:"page_size"`
				TotalCount int    `json:"total_count"`
			}{
				PageSize:   pageSize,
				TotalCount: 0,
			},
		},
	}, nil
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
