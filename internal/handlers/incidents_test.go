package handlers

import (
	"testing"
)

func TestCreateIncidentTool_Execute(t *testing.T) {
	tool := &CreateIncidentTool{}

	// Test missing required name parameter
	t.Run("missing required name", func(t *testing.T) {
		args := map[string]interface{}{
			"summary": "Test Summary",
		}

		_, err := tool.Execute(args)
		if err == nil {
			t.Error("Expected error for missing name parameter")
		}
		if err.Error() != "name parameter is required" {
			t.Errorf("Expected 'name parameter is required' error, got: %v", err)
		}
	})

	// Test name parameter with wrong type
	t.Run("name parameter wrong type", func(t *testing.T) {
		args := map[string]interface{}{
			"name": 123, // Not a string
		}

		_, err := tool.Execute(args)
		if err == nil {
			t.Error("Expected error for wrong type name parameter")
		}
		if err.Error() != "name parameter is required" {
			t.Errorf("Expected 'name parameter is required' error, got: %v", err)
		}
	})

	// Note: We can't test the full execution without a real client,
	// but we can test the parameter validation and schema
}

func TestUpdateIncidentTool_Schema(t *testing.T) {
	tool := &UpdateIncidentTool{}

	// Test Name
	if tool.Name() != "update_incident" {
		t.Errorf("Expected name 'update_incident', got %s", tool.Name())
	}

	// Test InputSchema has custom_field_entries
	schema := tool.InputSchema()
	properties := schema["properties"].(map[string]interface{})

	if _, ok := properties["custom_field_entries"]; !ok {
		t.Error("Schema should have 'custom_field_entries' property")
	}

	// Verify custom_field_entries is an array type
	cfe := properties["custom_field_entries"].(map[string]interface{})
	if cfe["type"] != "array" {
		t.Errorf("custom_field_entries should be type 'array', got %v", cfe["type"])
	}

	// Verify items schema has required fields
	items := cfe["items"].(map[string]interface{})
	itemProps := items["properties"].(map[string]interface{})
	if _, ok := itemProps["custom_field_id"]; !ok {
		t.Error("custom_field_entries items should have 'custom_field_id' property")
	}
	if _, ok := itemProps["values"]; !ok {
		t.Error("custom_field_entries items should have 'values' property")
	}
}

func TestUpdateIncidentTool_Execute(t *testing.T) {
	tool := &UpdateIncidentTool{}

	// Test missing required incident_id parameter
	t.Run("missing incident_id", func(t *testing.T) {
		args := map[string]interface{}{
			"name": "Test",
		}
		_, err := tool.Execute(args)
		if err == nil {
			t.Error("Expected error for missing incident_id")
		}
	})

	// Test no update fields provided
	t.Run("no update fields", func(t *testing.T) {
		args := map[string]interface{}{
			"incident_id": "some-id",
		}
		_, err := tool.Execute(args)
		if err == nil {
			t.Error("Expected error when no update fields are provided")
		}
		if err.Error() != "at least one field to update must be provided" {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	// Test custom_field_entries is recognized as a valid update field
	// (will panic at API call since no client, but should get past validation)
	t.Run("custom_field_entries counts as update", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				// Expected: nil pointer dereference because apiClient is nil
				// This means we got past the "at least one field" validation
			}
		}()

		args := map[string]interface{}{
			"incident_id": "some-id",
			"custom_field_entries": []interface{}{
				map[string]interface{}{
					"custom_field_id": "field-123",
					"values": []interface{}{
						map[string]interface{}{
							"value_catalog_entry_id": "option-456",
						},
					},
				},
			},
		}
		_, err := tool.Execute(args)
		if err != nil && err.Error() == "at least one field to update must be provided" {
			t.Error("custom_field_entries should count as a valid update field")
		}
	})

	// Test custom_field_entries with empty entries is not counted
	t.Run("empty custom_field_entries not counted", func(t *testing.T) {
		args := map[string]interface{}{
			"incident_id":          "some-id",
			"custom_field_entries": []interface{}{},
		}
		_, err := tool.Execute(args)
		if err == nil {
			t.Error("Expected error when custom_field_entries is empty")
		}
		if err.Error() != "at least one field to update must be provided" {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	// Test custom_field_entries with invalid entries are skipped
	t.Run("invalid entries skipped", func(t *testing.T) {
		args := map[string]interface{}{
			"incident_id": "some-id",
			"custom_field_entries": []interface{}{
				map[string]interface{}{
					// Missing custom_field_id
					"values": []interface{}{},
				},
			},
		}
		_, err := tool.Execute(args)
		if err == nil {
			t.Error("Expected error when all entries are invalid")
		}
		if err.Error() != "at least one field to update must be provided" {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}

func TestCreateIncidentTool_Schema(t *testing.T) {
	tool := &CreateIncidentTool{}

	// Test Name
	if tool.Name() != "create_incident" {
		t.Errorf("Expected name 'create_incident', got %s", tool.Name())
	}

	// Test Description
	if tool.Description() != "Create a new incident in incident.io" {
		t.Errorf("Unexpected description: %s", tool.Description())
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

	required := schema["required"].([]interface{})
	if len(required) != 1 || required[0] != "name" {
		t.Error("Schema should require only 'name'")
	}
}
