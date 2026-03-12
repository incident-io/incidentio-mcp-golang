package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// ListCatalogTypes returns all catalog types
func (c *Client) ListCatalogTypes(ctx context.Context) (*ListCatalogTypesResponse, error) {
	respBody, err := c.doRequestWithBase(ctx, BaseURLV3, "GET", "/catalog_types", nil, nil)
	if err != nil {
		return nil, err
	}

	var response ListCatalogTypesResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// ListCatalogEntriesOptions represents options for listing catalog entries
type ListCatalogEntriesOptions struct {
	CatalogTypeID string
	PageSize      int
	After         string
	Identifier    string
}

// ListCatalogEntries returns catalog entries for a given type
func (c *Client) ListCatalogEntries(ctx context.Context, opts ListCatalogEntriesOptions) (*ListCatalogEntriesResponse, error) {
	// Set default page_size if not provided (required by API)
	pageSize := opts.PageSize
	if pageSize == 0 {
		pageSize = 25 // Default page size
	}

	params := url.Values{}
	if opts.CatalogTypeID != "" {
		params.Set("catalog_type_id", opts.CatalogTypeID)
	}
	params.Set("page_size", fmt.Sprintf("%d", pageSize)) // Always set page_size (required)
	if opts.After != "" {
		params.Set("after", opts.After)
	}
	if opts.Identifier != "" {
		params.Set("identifier", opts.Identifier)
	}

	respBody, err := c.doRequestWithBase(ctx, BaseURLV3, "GET", "/catalog_entries", params, nil)
	if err != nil {
		return nil, err
	}

	var response ListCatalogEntriesResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// UpdateCatalogEntry updates a catalog entry by ID
func (c *Client) UpdateCatalogEntry(ctx context.Context, id string, req UpdateCatalogEntryRequest) (*CatalogEntry, error) {
	respBody, err := c.doRequestWithBase(ctx, BaseURLV3, "PUT", fmt.Sprintf("/catalog_entries/%s", id), nil, req)
	if err != nil {
		return nil, err
	}

	var response struct {
		CatalogEntry CatalogEntry `json:"catalog_entry"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response.CatalogEntry, nil
}

// GetCatalogEntry retrieves a specific catalog entry by ID
func (c *Client) GetCatalogEntry(ctx context.Context, id string) (*CatalogEntry, error) {
	respBody, err := c.doRequestWithBase(ctx, BaseURLV3, "GET", fmt.Sprintf("/catalog_entries/%s", id), nil, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		CatalogEntry CatalogEntry `json:"catalog_entry"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response.CatalogEntry, nil
}
