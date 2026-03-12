package handlers

import (
	"context"
	"testing"
)

func TestListCustomFieldsTool_Schema(t *testing.T) {
	tool := &ListCustomFieldsTool{}

	// Test Name
	if tool.Name() != "list_custom_fields" {
		t.Errorf("Expected name 'list_custom_fields', got %s", tool.Name())
	}

	// Test Description
	desc := tool.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}

	// Test InputSchema
	schema := tool.InputSchema()
	if schema["type"] != "object" {
		t.Error("Schema type should be 'object'")
	}
}

func TestGetCustomFieldTool_Execute(t *testing.T) {
	tool := &GetCustomFieldTool{}

	// Test missing required id parameter
	t.Run("missing required id", func(t *testing.T) {
		args := map[string]interface{}{}

		_, err := tool.Execute(context.Background(), args)
		if err == nil {
			t.Error("Expected error for missing id parameter")
		}
		if err.Error() != "id parameter is required" {
			t.Errorf("Expected 'id parameter is required' error, got: %v", err)
		}
	})

	// Test id parameter with wrong type
	t.Run("id parameter wrong type", func(t *testing.T) {
		args := map[string]interface{}{
			"id": 123, // Not a string
		}

		_, err := tool.Execute(context.Background(), args)
		if err == nil {
			t.Error("Expected error for wrong type id parameter")
		}
	})

	// Test empty id parameter
	t.Run("empty id parameter", func(t *testing.T) {
		args := map[string]interface{}{
			"id": "",
		}

		_, err := tool.Execute(context.Background(), args)
		if err == nil {
			t.Error("Expected error for empty id parameter")
		}
	})
}

func TestGetCustomFieldTool_Schema(t *testing.T) {
	tool := &GetCustomFieldTool{}

	// Test Name
	if tool.Name() != "get_custom_field" {
		t.Errorf("Expected name 'get_custom_field', got %s", tool.Name())
	}

	// Test Description
	desc := tool.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}

	// Test InputSchema
	schema := tool.InputSchema()
	if schema["type"] != "object" {
		t.Error("Schema type should be 'object'")
	}

	properties := schema["properties"].(map[string]interface{})
	if _, ok := properties["id"]; !ok {
		t.Error("Schema should have 'id' property")
	}

	required := schema["required"].([]string)
	if len(required) != 1 || required[0] != "id" {
		t.Error("Schema should require only 'id'")
	}
}

func TestSearchCustomFieldsTool_Execute(t *testing.T) {
	// Note: We can't test full execution without a real client,
	// but SearchCustomFieldsTool doesn't validate required parameters
	// since both query and field_type are optional
}

func TestSearchCustomFieldsTool_Schema(t *testing.T) {
	tool := &SearchCustomFieldsTool{}

	// Test Name
	if tool.Name() != "search_custom_fields" {
		t.Errorf("Expected name 'search_custom_fields', got %s", tool.Name())
	}

	// Test Description
	desc := tool.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}

	// Test InputSchema
	schema := tool.InputSchema()
	if schema["type"] != "object" {
		t.Error("Schema type should be 'object'")
	}

	properties := schema["properties"].(map[string]interface{})
	if _, ok := properties["query"]; !ok {
		t.Error("Schema should have 'query' property")
	}
	if _, ok := properties["field_type"]; !ok {
		t.Error("Schema should have 'field_type' property")
	}
}

func TestCreateCustomFieldTool_Execute(t *testing.T) {
	tool := &CreateCustomFieldTool{}

	// Test missing required name parameter
	t.Run("missing required name", func(t *testing.T) {
		args := map[string]interface{}{
			"description": "Test Description",
			"field_type":  "text",
		}

		_, err := tool.Execute(context.Background(), args)
		if err == nil {
			t.Error("Expected error for missing name parameter")
		}
		if err.Error() != "name is required" {
			t.Errorf("Expected 'name is required' error, got: %v", err)
		}
	})

	// Test missing required description parameter
	t.Run("missing required description", func(t *testing.T) {
		args := map[string]interface{}{
			"name":       "Test Field",
			"field_type": "text",
		}

		_, err := tool.Execute(context.Background(), args)
		if err == nil {
			t.Error("Expected error for missing description parameter")
		}
		if err.Error() != "description is required" {
			t.Errorf("Expected 'description is required' error, got: %v", err)
		}
	})

	// Test missing required field_type parameter
	t.Run("missing required field_type", func(t *testing.T) {
		args := map[string]interface{}{
			"name":        "Test Field",
			"description": "Test Description",
		}

		_, err := tool.Execute(context.Background(), args)
		if err == nil {
			t.Error("Expected error for missing field_type parameter")
		}
		if err.Error() != "field_type is required" {
			t.Errorf("Expected 'field_type is required' error, got: %v", err)
		}
	})

	// Test name parameter with wrong type
	t.Run("name parameter wrong type", func(t *testing.T) {
		args := map[string]interface{}{
			"name":        123, // Not a string
			"description": "Test Description",
			"field_type":  "text",
		}

		_, err := tool.Execute(context.Background(), args)
		if err == nil {
			t.Error("Expected error for wrong type name parameter")
		}
	})
}

func TestCreateCustomFieldTool_Schema(t *testing.T) {
	tool := &CreateCustomFieldTool{}

	// Test Name
	if tool.Name() != "create_custom_field" {
		t.Errorf("Expected name 'create_custom_field', got %s", tool.Name())
	}

	// Test Description
	desc := tool.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}

	// Test InputSchema
	schema := tool.InputSchema()
	if schema["type"] != "object" {
		t.Error("Schema type should be 'object'")
	}

	properties := schema["properties"].(map[string]interface{})
	if _, ok := properties["name"]; !ok {
		t.Error("Schema should have 'name' property")
	}
	if _, ok := properties["description"]; !ok {
		t.Error("Schema should have 'description' property")
	}
	if _, ok := properties["field_type"]; !ok {
		t.Error("Schema should have 'field_type' property")
	}

	required := schema["required"].([]string)
	if len(required) != 3 {
		t.Errorf("Schema should require 3 fields, got %d", len(required))
	}
}

func TestUpdateCustomFieldTool_Execute(t *testing.T) {
	tool := &UpdateCustomFieldTool{}

	// Test missing required id parameter
	t.Run("missing required id", func(t *testing.T) {
		args := map[string]interface{}{
			"name": "Updated Name",
		}

		_, err := tool.Execute(context.Background(), args)
		if err == nil {
			t.Error("Expected error for missing id parameter")
		}
		if err.Error() != "id is required" {
			t.Errorf("Expected 'id is required' error, got: %v", err)
		}
	})

	// Test id parameter with wrong type
	t.Run("id parameter wrong type", func(t *testing.T) {
		args := map[string]interface{}{
			"id":   123, // Not a string
			"name": "Updated Name",
		}

		_, err := tool.Execute(context.Background(), args)
		if err == nil {
			t.Error("Expected error for wrong type id parameter")
		}
	})

	// Test empty id parameter
	t.Run("empty id parameter", func(t *testing.T) {
		args := map[string]interface{}{
			"id":   "",
			"name": "Updated Name",
		}

		_, err := tool.Execute(context.Background(), args)
		if err == nil {
			t.Error("Expected error for empty id parameter")
		}
	})
}

func TestUpdateCustomFieldTool_Schema(t *testing.T) {
	tool := &UpdateCustomFieldTool{}

	// Test Name
	if tool.Name() != "update_custom_field" {
		t.Errorf("Expected name 'update_custom_field', got %s", tool.Name())
	}

	// Test Description
	desc := tool.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}

	// Test InputSchema
	schema := tool.InputSchema()
	if schema["type"] != "object" {
		t.Error("Schema type should be 'object'")
	}

	properties := schema["properties"].(map[string]interface{})
	if _, ok := properties["id"]; !ok {
		t.Error("Schema should have 'id' property")
	}

	required := schema["required"].([]string)
	if len(required) != 1 || required[0] != "id" {
		t.Error("Schema should require only 'id'")
	}
}

func TestDeleteCustomFieldTool_Execute(t *testing.T) {
	tool := &DeleteCustomFieldTool{}

	// Test missing required id parameter
	t.Run("missing required id", func(t *testing.T) {
		args := map[string]interface{}{}

		_, err := tool.Execute(context.Background(), args)
		if err == nil {
			t.Error("Expected error for missing id parameter")
		}
		if err.Error() != "id is required" {
			t.Errorf("Expected 'id is required' error, got: %v", err)
		}
	})

	// Test id parameter with wrong type
	t.Run("id parameter wrong type", func(t *testing.T) {
		args := map[string]interface{}{
			"id": 123, // Not a string
		}

		_, err := tool.Execute(context.Background(), args)
		if err == nil {
			t.Error("Expected error for wrong type id parameter")
		}
	})

	// Test empty id parameter
	t.Run("empty id parameter", func(t *testing.T) {
		args := map[string]interface{}{
			"id": "",
		}

		_, err := tool.Execute(context.Background(), args)
		if err == nil {
			t.Error("Expected error for empty id parameter")
		}
	})
}

func TestDeleteCustomFieldTool_Schema(t *testing.T) {
	tool := &DeleteCustomFieldTool{}

	// Test Name
	if tool.Name() != "delete_custom_field" {
		t.Errorf("Expected name 'delete_custom_field', got %s", tool.Name())
	}

	// Test Description
	desc := tool.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}

	// Test InputSchema
	schema := tool.InputSchema()
	if schema["type"] != "object" {
		t.Error("Schema type should be 'object'")
	}

	properties := schema["properties"].(map[string]interface{})
	if _, ok := properties["id"]; !ok {
		t.Error("Schema should have 'id' property")
	}

	required := schema["required"].([]string)
	if len(required) != 1 || required[0] != "id" {
		t.Error("Schema should require only 'id'")
	}
}

func TestListCustomFieldOptionsTool_Schema(t *testing.T) {
	tool := &ListCustomFieldOptionsTool{}

	// Test Name
	if tool.Name() != "list_custom_field_options" {
		t.Errorf("Expected name 'list_custom_field_options', got %s", tool.Name())
	}

	// Test Description
	desc := tool.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}

	// Test InputSchema
	schema := tool.InputSchema()
	if schema["type"] != "object" {
		t.Error("Schema type should be 'object'")
	}
}

func TestCreateCustomFieldOptionTool_Execute(t *testing.T) {
	tool := &CreateCustomFieldOptionTool{}

	// Test missing required custom_field_id parameter
	t.Run("missing required custom_field_id", func(t *testing.T) {
		args := map[string]interface{}{
			"value": "Option Value",
		}

		_, err := tool.Execute(context.Background(), args)
		if err == nil {
			t.Error("Expected error for missing custom_field_id parameter")
		}
		if err.Error() != "custom_field_id is required" {
			t.Errorf("Expected 'custom_field_id is required' error, got: %v", err)
		}
	})

	// Test missing required value parameter
	t.Run("missing required value", func(t *testing.T) {
		args := map[string]interface{}{
			"custom_field_id": "cf_123",
		}

		_, err := tool.Execute(context.Background(), args)
		if err == nil {
			t.Error("Expected error for missing value parameter")
		}
		if err.Error() != "value is required" {
			t.Errorf("Expected 'value is required' error, got: %v", err)
		}
	})

	// Test custom_field_id parameter with wrong type
	t.Run("custom_field_id parameter wrong type", func(t *testing.T) {
		args := map[string]interface{}{
			"custom_field_id": 123, // Not a string
			"value":           "Option Value",
		}

		_, err := tool.Execute(context.Background(), args)
		if err == nil {
			t.Error("Expected error for wrong type custom_field_id parameter")
		}
	})

	// Test empty custom_field_id parameter
	t.Run("empty custom_field_id parameter", func(t *testing.T) {
		args := map[string]interface{}{
			"custom_field_id": "",
			"value":           "Option Value",
		}

		_, err := tool.Execute(context.Background(), args)
		if err == nil {
			t.Error("Expected error for empty custom_field_id parameter")
		}
	})
}

func TestCreateCustomFieldOptionTool_Schema(t *testing.T) {
	tool := &CreateCustomFieldOptionTool{}

	// Test Name
	if tool.Name() != "create_custom_field_option" {
		t.Errorf("Expected name 'create_custom_field_option', got %s", tool.Name())
	}

	// Test Description
	desc := tool.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}

	// Test InputSchema
	schema := tool.InputSchema()
	if schema["type"] != "object" {
		t.Error("Schema type should be 'object'")
	}

	properties := schema["properties"].(map[string]interface{})
	if _, ok := properties["custom_field_id"]; !ok {
		t.Error("Schema should have 'custom_field_id' property")
	}
	if _, ok := properties["value"]; !ok {
		t.Error("Schema should have 'value' property")
	}

	required := schema["required"].([]string)
	if len(required) != 2 {
		t.Errorf("Schema should require 2 fields, got %d", len(required))
	}
}
