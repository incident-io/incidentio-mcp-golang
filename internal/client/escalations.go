package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// ListEscalationPathsParams is query options for GET /escalation_paths.
// See https://docs.incident.io/api-reference/escalations-v2/listpaths
type ListEscalationPathsParams struct {
	PageSize int
	After    string
}

// ListEscalationPathsResponse is the API response for listing escalation paths.
type ListEscalationPathsResponse struct {
	EscalationPaths []json.RawMessage `json:"escalation_paths"`
	PaginationMeta  struct {
		After            string `json:"after,omitempty"`
		PageSize         int    `json:"page_size"`
		TotalRecordCount int    `json:"total_record_count,omitempty"`
	} `json:"pagination_meta"`
}

// ListEscalationPaths lists escalation paths (On-call).
func (c *Client) ListEscalationPaths(params *ListEscalationPathsParams) (*ListEscalationPathsResponse, error) {
	pageSize := 25
	if params != nil && params.PageSize > 0 {
		pageSize = params.PageSize
	}
	if pageSize > 25 {
		pageSize = 25
	}

	v := url.Values{}
	v.Set("page_size", strconv.Itoa(pageSize))
	if params != nil && params.After != "" {
		v.Set("after", params.After)
	}

	endpoint := "/escalation_paths?" + v.Encode()
	respBody, err := c.doRequest("GET", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var out ListEscalationPathsResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("failed to unmarshal escalation paths list: %w", err)
	}
	return &out, nil
}

// GetEscalationPath returns one escalation path by ID.
func (c *Client) GetEscalationPath(id string) (json.RawMessage, error) {
	if id == "" {
		return nil, fmt.Errorf("escalation path id is required")
	}
	respBody, err := c.doRequest("GET", fmt.Sprintf("/escalation_paths/%s", id), nil, nil)
	if err != nil {
		return nil, err
	}
	var wrap struct {
		EscalationPath json.RawMessage `json:"escalation_path"`
	}
	if err := json.Unmarshal(respBody, &wrap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal escalation path: %w", err)
	}
	return wrap.EscalationPath, nil
}

// CreateEscalationPath creates an escalation path. Body matches the API payload
// (name, path, team_ids, repeat_config, working_hours, etc.).
func (c *Client) CreateEscalationPath(body map[string]interface{}) (json.RawMessage, error) {
	respBody, err := c.doRequest("POST", "/escalation_paths", nil, body)
	if err != nil {
		return nil, err
	}
	var wrap struct {
		EscalationPath json.RawMessage `json:"escalation_path"`
	}
	if err := json.Unmarshal(respBody, &wrap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal create escalation path response: %w", err)
	}
	return wrap.EscalationPath, nil
}

// UpdateEscalationPath replaces an escalation path (PUT).
func (c *Client) UpdateEscalationPath(id string, body map[string]interface{}) (json.RawMessage, error) {
	if id == "" {
		return nil, fmt.Errorf("escalation path id is required")
	}
	respBody, err := c.doRequest("PUT", fmt.Sprintf("/escalation_paths/%s", id), nil, body)
	if err != nil {
		return nil, err
	}
	var wrap struct {
		EscalationPath json.RawMessage `json:"escalation_path"`
	}
	if err := json.Unmarshal(respBody, &wrap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal update escalation path response: %w", err)
	}
	return wrap.EscalationPath, nil
}

// DestroyEscalationPath archives an escalation path (DELETE, 204).
func (c *Client) DestroyEscalationPath(id string) error {
	if id == "" {
		return fmt.Errorf("escalation path id is required")
	}
	_, err := c.doRequest("DELETE", fmt.Sprintf("/escalation_paths/%s", id), nil, nil)
	return err
}

// ListEscalationsParams filters for GET /escalations.
// See https://docs.incident.io/api-reference/escalations-v2/list
type ListEscalationsParams struct {
	PageSize int
	After    string

	EscalationPathOneOf []string
	EscalationPathNotIn []string
	StatusOneOf         []string
	StatusNotIn         []string
	AlertOneOf          []string
	AlertNotIn          []string

	CreatedAtGte       string
	CreatedAtLte       string
	CreatedAtDateRange string
	UpdatedAtGte       string
	UpdatedAtLte       string
	UpdatedAtDateRange string

	IdempotencyKeyIs         string
	IdempotencyKeyStartsWith string
}

// ListEscalationsResponse is the API response for listing escalations.
type ListEscalationsResponse struct {
	Escalations    []json.RawMessage `json:"escalations"`
	PaginationMeta struct {
		After            string `json:"after,omitempty"`
		PageSize         int    `json:"page_size"`
		TotalRecordCount int    `json:"total_record_count,omitempty"`
	} `json:"pagination_meta"`
}

// ListEscalations lists escalations with optional filters.
func (c *Client) ListEscalations(params *ListEscalationsParams) (*ListEscalationsResponse, error) {
	pageSize := 25
	if params != nil && params.PageSize > 0 {
		pageSize = params.PageSize
	}
	if pageSize > 50 {
		pageSize = 50
	}

	v := url.Values{}
	v.Set("page_size", strconv.Itoa(pageSize))
	if params != nil {
		if params.After != "" {
			v.Set("after", params.After)
		}
		for _, id := range params.EscalationPathOneOf {
			v.Add("escalation_path[one_of]", id)
		}
		for _, id := range params.EscalationPathNotIn {
			v.Add("escalation_path[not_in]", id)
		}
		for _, s := range params.StatusOneOf {
			v.Add("status[one_of]", s)
		}
		for _, s := range params.StatusNotIn {
			v.Add("status[not_in]", s)
		}
		for _, id := range params.AlertOneOf {
			v.Add("alert[one_of]", id)
		}
		for _, id := range params.AlertNotIn {
			v.Add("alert[not_in]", id)
		}
		if params.CreatedAtGte != "" {
			v.Set("created_at[gte]", params.CreatedAtGte)
		}
		if params.CreatedAtLte != "" {
			v.Set("created_at[lte]", params.CreatedAtLte)
		}
		if params.CreatedAtDateRange != "" {
			v.Set("created_at[date_range]", params.CreatedAtDateRange)
		}
		if params.UpdatedAtGte != "" {
			v.Set("updated_at[gte]", params.UpdatedAtGte)
		}
		if params.UpdatedAtLte != "" {
			v.Set("updated_at[lte]", params.UpdatedAtLte)
		}
		if params.UpdatedAtDateRange != "" {
			v.Set("updated_at[date_range]", params.UpdatedAtDateRange)
		}
		if params.IdempotencyKeyIs != "" {
			v.Set("idempotency_key[is]", params.IdempotencyKeyIs)
		}
		if params.IdempotencyKeyStartsWith != "" {
			v.Set("idempotency_key[starts_with]", params.IdempotencyKeyStartsWith)
		}
	}

	endpoint := "/escalations?" + v.Encode()
	respBody, err := c.doRequest("GET", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var out ListEscalationsResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("failed to unmarshal escalations list: %w", err)
	}
	return &out, nil
}

// GetEscalation returns one escalation by ID.
func (c *Client) GetEscalation(id string) (json.RawMessage, error) {
	if id == "" {
		return nil, fmt.Errorf("escalation id is required")
	}
	respBody, err := c.doRequest("GET", fmt.Sprintf("/escalations/%s", id), nil, nil)
	if err != nil {
		return nil, err
	}
	var wrap struct {
		Escalation json.RawMessage `json:"escalation"`
	}
	if err := json.Unmarshal(respBody, &wrap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal escalation: %w", err)
	}
	return wrap.Escalation, nil
}

// CreateEscalationRequest is the body for POST /escalations.
// Provide either EscalationPathID or UserIDs, not both.
type CreateEscalationRequest struct {
	Title            string   `json:"title"`
	Description      string   `json:"description,omitempty"`
	EscalationPathID string   `json:"escalation_path_id,omitempty"`
	UserIDs          []string `json:"user_ids,omitempty"`
	IdempotencyKey   string   `json:"idempotency_key,omitempty"`
}

// CreateEscalation triggers a new escalation.
func (c *Client) CreateEscalation(req *CreateEscalationRequest) (json.RawMessage, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}
	respBody, err := c.doRequest("POST", "/escalations", nil, req)
	if err != nil {
		return nil, err
	}
	var wrap struct {
		Escalation json.RawMessage `json:"escalation"`
	}
	if err := json.Unmarshal(respBody, &wrap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal create escalation response: %w", err)
	}
	return wrap.Escalation, nil
}
