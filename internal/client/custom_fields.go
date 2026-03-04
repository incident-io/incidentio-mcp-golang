package client

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// ListCustomFields retrieves all custom fields (cached for 5 minutes)
func (c *Client) ListCustomFields() (*ListCustomFieldsResponse, error) {
	// Check cache first
	cacheKey := "custom_fields:list"
	if cached, found := c.cache.Get(cacheKey); found {
		return cached.(*ListCustomFieldsResponse), nil
	}

	respBody, err := c.doRequest("GET", "/custom_fields", nil, nil)
	if err != nil {
		return nil, err
	}

	var response ListCustomFieldsResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Cache the response
	c.cache.Set(cacheKey, &response)

	return &response, nil
}

// GetCustomField retrieves a specific custom field by ID (cached for 5 minutes)
func (c *Client) GetCustomField(id string) (*CustomField, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("custom_field:%s", id)
	if cached, found := c.cache.Get(cacheKey); found {
		return cached.(*CustomField), nil
	}

	respBody, err := c.doRequest("GET", fmt.Sprintf("/custom_fields/%s", id), nil, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		CustomField CustomField `json:"custom_field"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Cache the response
	c.cache.Set(cacheKey, &response.CustomField)

	return &response.CustomField, nil
}

// CreateCustomField creates a new custom field and invalidates cache
func (c *Client) CreateCustomField(req *CreateCustomFieldRequest) (*CustomField, error) {
	respBody, err := c.doRequest("POST", "/custom_fields", nil, req)
	if err != nil {
		return nil, err
	}

	// Invalidate list cache since we created a new field
	c.cache.Delete("custom_fields:list")

	var response struct {
		CustomField CustomField `json:"custom_field"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response.CustomField, nil
}

// UpdateCustomField updates an existing custom field
func (c *Client) UpdateCustomField(id string, req *UpdateCustomFieldRequest) (*CustomField, error) {
	respBody, err := c.doRequest("PUT", fmt.Sprintf("/custom_fields/%s", id), nil, req)
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
func (c *Client) DeleteCustomField(id string) error {
	_, err := c.doRequest("DELETE", fmt.Sprintf("/custom_fields/%s", id), nil, nil)
	return err
}

// ListCustomFieldOptions retrieves all options for custom fields
func (c *Client) ListCustomFieldOptions() ([]CustomFieldOption, error) {
	respBody, err := c.doRequest("GET", "/custom_field_options", nil, nil)
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
func (c *Client) GetCustomFieldOption(id string) (*CustomFieldOption, error) {
	respBody, err := c.doRequest("GET", fmt.Sprintf("/custom_field_options/%s", id), nil, nil)
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
func (c *Client) CreateCustomFieldOption(req *CreateCustomFieldOptionRequest) (*CustomFieldOption, error) {
	respBody, err := c.doRequest("POST", "/custom_field_options", nil, req)
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
func (c *Client) UpdateCustomFieldOption(id string, req *UpdateCustomFieldOptionRequest) (*CustomFieldOption, error) {
	respBody, err := c.doRequest("PUT", fmt.Sprintf("/custom_field_options/%s", id), nil, req)
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
func (c *Client) DeleteCustomFieldOption(id string) error {
	_, err := c.doRequest("DELETE", fmt.Sprintf("/custom_field_options/%s", id), nil, nil)
	return err
}

// SearchCustomFields searches for custom fields by name or field type
func (c *Client) SearchCustomFields(query string, fieldType string) ([]CustomField, error) {
	params := url.Values{}
	if query != "" {
		params.Set("query", query)
	}
	if fieldType != "" {
		params.Set("field_type", fieldType)
	}

	respBody, err := c.doRequest("GET", "/custom_fields", params, nil)
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
