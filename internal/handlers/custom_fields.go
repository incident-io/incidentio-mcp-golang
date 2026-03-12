package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/incident-io/incidentio-mcp-golang/internal/client"
)

// ListCustomFieldsTool lists all custom fields
type ListCustomFieldsTool struct {
	apiClient *client.Client
}

func NewListCustomFieldsTool(c *client.Client) *ListCustomFieldsTool {
	return &ListCustomFieldsTool{apiClient: c}
}

func (t *ListCustomFieldsTool) Name() string {
	return "list_custom_fields"
}

func (t *ListCustomFieldsTool) Description() string {
	return "List all custom fields configured in incident.io. Use this to discover what custom fields exist (like \"Affected Team\", \"Priority\", etc.) before filtering incidents by them.\n\n" +
		"WHEN TO USE: If you need to filter incidents by a custom attribute (team, department, priority, etc.) but don't know the exact field name, call this first to see what custom fields are available, then use search_custom_fields or use the ID directly in list_incidents."
}

func (t *ListCustomFieldsTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

func (t *ListCustomFieldsTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	resp, err := t.apiClient.ListCustomFields(ctx)
	if err != nil {
		return "", err
	}

	result, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	return string(result), nil
}

// GetCustomFieldTool retrieves a specific custom field
type GetCustomFieldTool struct {
	apiClient *client.Client
}

func NewGetCustomFieldTool(c *client.Client) *GetCustomFieldTool {
	return &GetCustomFieldTool{apiClient: c}
}

func (t *GetCustomFieldTool) Name() string {
	return "get_custom_field"
}

func (t *GetCustomFieldTool) Description() string {
	return "Get details of a specific custom field by ID, including its options, type, and configuration"
}

func (t *GetCustomFieldTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "The custom field ID",
			},
		},
		"required": []string{"id"},
	}
}

func (t *GetCustomFieldTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("id parameter is required")
	}

	field, err := t.apiClient.GetCustomField(ctx, id)
	if err != nil {
		return "", err
	}

	result, err := json.MarshalIndent(field, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	return string(result), nil
}

// SearchCustomFieldsTool searches for custom fields
type SearchCustomFieldsTool struct {
	apiClient *client.Client
}

func NewSearchCustomFieldsTool(c *client.Client) *SearchCustomFieldsTool {
	return &SearchCustomFieldsTool{apiClient: c}
}

func (t *SearchCustomFieldsTool) Name() string {
	return "search_custom_fields"
}

func (t *SearchCustomFieldsTool) Description() string {
	return "Search for custom fields by name to get their IDs AND options for filtering.\n\n" +
		"IMPORTANT: When you get the custom field, check its 'options' array. If it's a select field:\n" +
		"- The custom_field_value must be the OPTION ID (e.g., '01ABC...'), NOT the option label\n" +
		"- Look for the option with matching 'value' field, then use its 'id'\n\n" +
		"WORKFLOW for team/department filtering:\n" +
		"1. search_custom_fields({\"query\": \"team\"}) → get field and its options\n" +
		"2. Find the option where value=\"Engineering\" → get its id (e.g., '01XYZ...')\n" +
		"3. list_incidents({\"custom_field_id\": \"cf_123\", \"custom_field_value\": \"01XYZ...\"})  ← Use option ID!\n\n" +
		"Example:\n" +
		"User: 'show Engineering team incidents'\n" +
		"→ search_custom_fields({\"query\": \"team\"})\n" +
		"  Returns: {id: 'cf_123', options: [{id: '01ABC', value: 'Engineering'}, ...]}\n" +
		"→ list_incidents({\"custom_field_id\": \"cf_123\", \"custom_field_value\": \"01ABC\"})"
}

func (t *SearchCustomFieldsTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Search query to match against custom field names and descriptions",
			},
			"field_type": map[string]interface{}{
				"type":        "string",
				"description": "Filter by field type: single_select, multi_select, text, link, numeric, etc.",
			},
		},
	}
}

func (t *SearchCustomFieldsTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	query := GetStringArg(args, "query")
	fieldType := GetStringArg(args, "field_type")

	fields, err := t.apiClient.SearchCustomFields(ctx, query, fieldType)
	if err != nil {
		return "", err
	}

	response := CreateSimpleResponse(fields, "")
	response["custom_fields"] = fields

	return FormatJSONResponse(response)
}

// CreateCustomFieldTool creates a new custom field
type CreateCustomFieldTool struct {
	apiClient *client.Client
}

func NewCreateCustomFieldTool(c *client.Client) *CreateCustomFieldTool {
	return &CreateCustomFieldTool{apiClient: c}
}

func (t *CreateCustomFieldTool) Name() string {
	return "create_custom_field"
}

func (t *CreateCustomFieldTool) Description() string {
	return "Create a new custom field for incidents. Custom fields allow you to capture additional structured data on incidents."
}

func (t *CreateCustomFieldTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Name of the custom field",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Description of what this field is for",
			},
			"field_type": map[string]interface{}{
				"type":        "string",
				"description": "Type of field: single_select, multi_select, text, link, numeric",
			},
			"required": map[string]interface{}{
				"type":        "string",
				"description": "When this field is required: never, always, before_closure",
				"default":     "never",
			},
			"show_before_closure": map[string]interface{}{
				"type":        "boolean",
				"description": "Show this field before incident closure",
				"default":     false,
			},
			"show_before_creation": map[string]interface{}{
				"type":        "boolean",
				"description": "Show this field when creating an incident",
				"default":     false,
			},
			"show_before_update": map[string]interface{}{
				"type":        "boolean",
				"description": "Show this field when updating an incident",
				"default":     false,
			},
			"options": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Options for select fields (single_select or multi_select)",
			},
			"catalog_type_id": map[string]interface{}{
				"type":        "string",
				"description": "Catalog type ID if this is a catalog field",
			},
		},
		"required": []string{"name", "description", "field_type"},
	}
}

func (t *CreateCustomFieldTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return "", fmt.Errorf("name is required")
	}

	description, ok := args["description"].(string)
	if !ok || description == "" {
		return "", fmt.Errorf("description is required")
	}

	fieldType, ok := args["field_type"].(string)
	if !ok || fieldType == "" {
		return "", fmt.Errorf("field_type is required")
	}

	req := &client.CreateCustomFieldRequest{
		Name:               name,
		Description:        description,
		FieldType:          fieldType,
		Required:           "never",
		ShowBeforeClosure:  false,
		ShowBeforeCreation: false,
		ShowBeforeUpdate:   false,
	}

	if required, ok := args["required"].(string); ok {
		req.Required = required
	}

	if showBeforeClosure, ok := args["show_before_closure"].(bool); ok {
		req.ShowBeforeClosure = showBeforeClosure
	}

	if showBeforeCreation, ok := args["show_before_creation"].(bool); ok {
		req.ShowBeforeCreation = showBeforeCreation
	}

	if showBeforeUpdate, ok := args["show_before_update"].(bool); ok {
		req.ShowBeforeUpdate = showBeforeUpdate
	}

	if catalogTypeID, ok := args["catalog_type_id"].(string); ok && catalogTypeID != "" {
		req.CatalogTypeID = catalogTypeID
	}

	if optionsRaw, ok := args["options"].([]interface{}); ok {
		options := make([]string, 0, len(optionsRaw))
		for _, opt := range optionsRaw {
			if optStr, ok := opt.(string); ok {
				options = append(options, optStr)
			}
		}
		req.Options = options
	}

	field, err := t.apiClient.CreateCustomField(ctx, req)
	if err != nil {
		return "", err
	}

	result, err := json.MarshalIndent(field, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	return string(result), nil
}

// UpdateCustomFieldTool updates an existing custom field
type UpdateCustomFieldTool struct {
	apiClient *client.Client
}

func NewUpdateCustomFieldTool(c *client.Client) *UpdateCustomFieldTool {
	return &UpdateCustomFieldTool{apiClient: c}
}

func (t *UpdateCustomFieldTool) Name() string {
	return "update_custom_field"
}

func (t *UpdateCustomFieldTool) Description() string {
	return "Update an existing custom field's configuration, including its name, description, and display settings"
}

func (t *UpdateCustomFieldTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "The custom field ID to update",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "New name for the custom field",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "New description for the custom field",
			},
			"required": map[string]interface{}{
				"type":        "string",
				"description": "When this field is required: never, always, before_closure",
			},
			"show_before_closure": map[string]interface{}{
				"type":        "boolean",
				"description": "Show this field before incident closure",
			},
			"show_before_creation": map[string]interface{}{
				"type":        "boolean",
				"description": "Show this field when creating an incident",
			},
			"show_before_update": map[string]interface{}{
				"type":        "boolean",
				"description": "Show this field when updating an incident",
			},
			"options": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Updated options for select fields",
			},
		},
		"required": []string{"id"},
	}
}

func (t *UpdateCustomFieldTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("id is required")
	}

	req := &client.UpdateCustomFieldRequest{}

	if name, ok := args["name"].(string); ok && name != "" {
		req.Name = name
	}

	if description, ok := args["description"].(string); ok && description != "" {
		req.Description = description
	}

	if required, ok := args["required"].(string); ok && required != "" {
		req.Required = required
	}

	if showBeforeClosure, ok := args["show_before_closure"].(bool); ok {
		req.ShowBeforeClosure = &showBeforeClosure
	}

	if showBeforeCreation, ok := args["show_before_creation"].(bool); ok {
		req.ShowBeforeCreation = &showBeforeCreation
	}

	if showBeforeUpdate, ok := args["show_before_update"].(bool); ok {
		req.ShowBeforeUpdate = &showBeforeUpdate
	}

	if optionsRaw, ok := args["options"].([]interface{}); ok {
		options := make([]string, 0, len(optionsRaw))
		for _, opt := range optionsRaw {
			if optStr, ok := opt.(string); ok {
				options = append(options, optStr)
			}
		}
		req.Options = options
	}

	field, err := t.apiClient.UpdateCustomField(ctx, id, req)
	if err != nil {
		return "", err
	}

	result, err := json.MarshalIndent(field, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	return string(result), nil
}

// DeleteCustomFieldTool deletes a custom field
type DeleteCustomFieldTool struct {
	apiClient *client.Client
}

func NewDeleteCustomFieldTool(c *client.Client) *DeleteCustomFieldTool {
	return &DeleteCustomFieldTool{apiClient: c}
}

func (t *DeleteCustomFieldTool) Name() string {
	return "delete_custom_field"
}

func (t *DeleteCustomFieldTool) Description() string {
	return "Delete a custom field. Warning: This will remove the field from all incidents."
}

func (t *DeleteCustomFieldTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "The custom field ID to delete",
			},
		},
		"required": []string{"id"},
	}
}

func (t *DeleteCustomFieldTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("id is required")
	}

	err := t.apiClient.DeleteCustomField(ctx, id)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Custom field %s deleted successfully", id), nil
}

// ListCustomFieldOptionsTool lists all custom field options
type ListCustomFieldOptionsTool struct {
	apiClient *client.Client
}

func NewListCustomFieldOptionsTool(c *client.Client) *ListCustomFieldOptionsTool {
	return &ListCustomFieldOptionsTool{apiClient: c}
}

func (t *ListCustomFieldOptionsTool) Name() string {
	return "list_custom_field_options"
}

func (t *ListCustomFieldOptionsTool) Description() string {
	return "List all custom field options across all custom fields. Useful for understanding available values for select fields."
}

func (t *ListCustomFieldOptionsTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

func (t *ListCustomFieldOptionsTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	options, err := t.apiClient.ListCustomFieldOptions(ctx)
	if err != nil {
		return "", err
	}

	response := CreateSimpleResponse(options, "")
	response["custom_field_options"] = options

	return FormatJSONResponse(response)
}

// CreateCustomFieldOptionTool creates a new option for a select-type custom field
type CreateCustomFieldOptionTool struct {
	apiClient *client.Client
}

func NewCreateCustomFieldOptionTool(c *client.Client) *CreateCustomFieldOptionTool {
	return &CreateCustomFieldOptionTool{apiClient: c}
}

func (t *CreateCustomFieldOptionTool) Name() string {
	return "create_custom_field_option"
}

func (t *CreateCustomFieldOptionTool) Description() string {
	return "Add a new option to a single_select or multi_select custom field"
}

func (t *CreateCustomFieldOptionTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"custom_field_id": map[string]interface{}{
				"type":        "string",
				"description": "The ID of the custom field to add this option to",
			},
			"value": map[string]interface{}{
				"type":        "string",
				"description": "The value/label for this option",
			},
			"sort_key": map[string]interface{}{
				"type":        "integer",
				"description": "Sort order for this option (optional)",
			},
		},
		"required": []string{"custom_field_id", "value"},
	}
}

func (t *CreateCustomFieldOptionTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	customFieldID, ok := args["custom_field_id"].(string)
	if !ok || customFieldID == "" {
		return "", fmt.Errorf("custom_field_id is required")
	}

	value, ok := args["value"].(string)
	if !ok || value == "" {
		return "", fmt.Errorf("value is required")
	}

	req := &client.CreateCustomFieldOptionRequest{
		CustomFieldID: customFieldID,
		Value:         value,
	}

	if sortKey, ok := args["sort_key"].(float64); ok {
		req.SortKey = int(sortKey)
	}

	option, err := t.apiClient.CreateCustomFieldOption(ctx, req)
	if err != nil {
		return "", err
	}

	result, err := json.MarshalIndent(option, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	return string(result), nil
}
