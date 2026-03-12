package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// ListPostmortems retrieves all postmortem documents for an organization
func (c *Client) ListPostmortems(ctx context.Context, opts *ListPostmortemsOptions) (*ListPostmortemsResponse, error) {
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

	respBody, err := c.doRequestWithBase(ctx, BaseURLV1, "GET", "/postmortem_documents", params, nil)
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
func (c *Client) GetPostmortem(ctx context.Context, id string) (*PostmortemDocument, error) {
	respBody, err := c.doRequestWithBase(ctx, BaseURLV1, "GET", fmt.Sprintf("/postmortem_documents/%s", id), nil, nil)
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
func (c *Client) GetPostmortemContent(ctx context.Context, id string) (*PostmortemContentResponse, error) {
	respBody, err := c.doRequestWithBase(ctx, BaseURLV1, "GET", fmt.Sprintf("/postmortem_documents/%s/content", id), nil, nil)
	if err != nil {
		return nil, err
	}

	var response PostmortemContentResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}
