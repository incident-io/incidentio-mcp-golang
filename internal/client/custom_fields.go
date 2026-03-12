package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// ListCustomFields retrieves all custom fields
func (c *Client) ListCustomFields(ctx context.Context) (*ListCustomFieldsResponse, error) {
	respBody, err := c.doRequest(ctx, "GET", "/custom_fields", nil, nil)
	if err != nil {
		return nil, err
	}

	var response ListCustomFieldsResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// GetCustomField retrieves a specific custom field by ID
func (c *Client) GetCustomField(ctx context.Context, id string) (*CustomField, error) {
	respBody, err := c.doRequest(ctx, "GET", fmt.Sprintf("/custom_fields/%s", id), nil, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		CustomField CustomField `json:"custom_field"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response.CustomField, nil
}

// CreateCustomField creates a new custom field
func (c *Client) CreateCustomField(ctx context.Context, req *CreateCustomFieldRequest) (*CustomField, error) {
	respBody, err := c.doRequest(ctx, "POST", "/custom_fields", nil, req)
	if err != nil {
		return nil, err
	}

	var response struct {
		CustomField CustomField `json:"custom_field"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response.CustomField, nil
}

// UpdateCustomField updates an existing custom field
func (c *Client) UpdateCustomField(ctx context.Context, id string, req *UpdateCustomFieldRequest) (*CustomField, error) {
	respBody, err := c.doRequest(ctx, "PUT", fmt.Sprintf("/custom_fields/%s", id), nil, req)
	if err != nil {
		return nil, err
	}

	var response struct {
		CustomField CustomField `json:"custom_field"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response.CustomField, nil
}

// DeleteCustomField deletes a custom field
func (c *Client) DeleteCustomField(ctx context.Context, id string) error {
	_, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/custom_fields/%s", id), nil, nil)
	return err
}

// ListCustomFieldOptions retrieves all options for custom fields
func (c *Client) ListCustomFieldOptions(ctx context.Context) ([]CustomFieldOption, error) {
	respBody, err := c.doRequest(ctx, "GET", "/custom_field_options", nil, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		CustomFieldOptions []CustomFieldOption `json:"custom_field_options"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response.CustomFieldOptions, nil
}

// GetCustomFieldOption retrieves a specific custom field option by ID
func (c *Client) GetCustomFieldOption(ctx context.Context, id string) (*CustomFieldOption, error) {
	respBody, err := c.doRequest(ctx, "GET", fmt.Sprintf("/custom_field_options/%s", id), nil, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		CustomFieldOption CustomFieldOption `json:"custom_field_option"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response.CustomFieldOption, nil
}

// CreateCustomFieldOption creates a new custom field option
func (c *Client) CreateCustomFieldOption(ctx context.Context, req *CreateCustomFieldOptionRequest) (*CustomFieldOption, error) {
	respBody, err := c.doRequest(ctx, "POST", "/custom_field_options", nil, req)
	if err != nil {
		return nil, err
	}

	var response struct {
		CustomFieldOption CustomFieldOption `json:"custom_field_option"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response.CustomFieldOption, nil
}

// UpdateCustomFieldOption updates an existing custom field option
func (c *Client) UpdateCustomFieldOption(ctx context.Context, id string, req *UpdateCustomFieldOptionRequest) (*CustomFieldOption, error) {
	respBody, err := c.doRequest(ctx, "PUT", fmt.Sprintf("/custom_field_options/%s", id), nil, req)
	if err != nil {
		return nil, err
	}

	var response struct {
		CustomFieldOption CustomFieldOption `json:"custom_field_option"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response.CustomFieldOption, nil
}

// DeleteCustomFieldOption deletes a custom field option
func (c *Client) DeleteCustomFieldOption(ctx context.Context, id string) error {
	_, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/custom_field_options/%s", id), nil, nil)
	return err
}

// SearchCustomFields searches for custom fields by name or field type
func (c *Client) SearchCustomFields(ctx context.Context, query string, fieldType string) ([]CustomField, error) {
	params := url.Values{}
	if query != "" {
		params.Set("query", query)
	}
	if fieldType != "" {
		params.Set("field_type", fieldType)
	}

	respBody, err := c.doRequest(ctx, "GET", "/custom_fields", params, nil)
	if err != nil {
		return nil, err
	}

	var response ListCustomFieldsResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Filter results if query provided (client-side filtering as fallback)
	if query != "" {
		filtered := []CustomField{}
		for _, field := range response.CustomFields {
			if contains(field.Name, query) || contains(field.Description, query) {
				filtered = append(filtered, field)
			}
		}
		return filtered, nil
	}

	return response.CustomFields, nil
}

// Helper function for case-insensitive string search
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	// Simple case-insensitive contains
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
