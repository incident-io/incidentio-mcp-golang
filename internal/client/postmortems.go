package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// ListPostmortems retrieves all postmortem documents for an organization
func (c *Client) ListPostmortems(opts *ListPostmortemsOptions) (*ListPostmortemsResponse, error) {
	// Note: Postmortems are under V1 API, not V2
	// We need to temporarily change the base URL for this request
	originalBaseURL := c.BaseURL()
	c.SetBaseURL("https://api.incident.io/v1")
	defer func() { c.SetBaseURL(originalBaseURL) }()

	params := url.Values{}

	if opts != nil {
		if opts.PageSize > 0 {
			params.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.After != "" {
			params.Set("after", opts.After)
		}
		if opts.IncidentID != "" {
			params.Set("incident_id", opts.IncidentID)
		}
		if opts.SortBy != "" {
			params.Set("sort_by", opts.SortBy)
		}
	}

	respBody, err := c.doRequest("GET", "/postmortem_documents", params, nil)
	if err != nil {
		return nil, err
	}

	var response ListPostmortemsResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// GetPostmortem retrieves a specific postmortem document by ID
func (c *Client) GetPostmortem(id string) (*PostmortemDocument, error) {
	// Note: Postmortems are under V1 API, not V2
	// We need to temporarily change the base URL for this request
	originalBaseURL := c.BaseURL()
	c.SetBaseURL("https://api.incident.io/v1")
	defer func() { c.SetBaseURL(originalBaseURL) }()

	respBody, err := c.doRequest("GET", fmt.Sprintf("/postmortem_documents/%s", id), nil, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		PostmortemDocument PostmortemDocument `json:"postmortem_document"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response.PostmortemDocument, nil
}

// GetPostmortemContent retrieves the markdown content of a postmortem document
func (c *Client) GetPostmortemContent(id string) (*PostmortemContentResponse, error) {
	// Note: Postmortems are under V1 API, not V2
	// We need to temporarily change the base URL for this request
	originalBaseURL := c.BaseURL()
	c.SetBaseURL("https://api.incident.io/v1")
	defer func() { c.SetBaseURL(originalBaseURL) }()

	respBody, err := c.doRequest("GET", fmt.Sprintf("/postmortem_documents/%s/content", id), nil, nil)
	if err != nil {
		return nil, err
	}

	var response PostmortemContentResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}
