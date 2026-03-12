package handlers

import (
	"context"
	"strings"
	"testing"
)

func TestCreateIncidentEnhancedTool_Execute(t *testing.T) {
	tool := &CreateIncidentEnhancedTool{}

	// Test missing required name parameter
	t.Run("missing required name", func(t *testing.T) {
		args := map[string]interface{}{
			"summary": "Test Summary",
		}

		_, err := tool.Execute(context.Background(), args)
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

		_, err := tool.Execute(context.Background(), args)
		if err == nil {
			t.Error("Expected error for wrong type name parameter")
		}
		if err.Error() != "name parameter is required" {
			t.Errorf("Expected 'name parameter is required' error, got: %v", err)
		}
	})
}

func TestCreateIncidentEnhancedTool_Schema(t *testing.T) {
	tool := &CreateIncidentEnhancedTool{}

	// Test Name
	if tool.Name() != "create_incident_smart" {
		t.Errorf("Expected name 'create_incident_smart', got %s", tool.Name())
	}

	// Test Description
	expectedDesc := "Create a new incident with smart defaults - automatically fetches first available severity, type, and status if not provided"
	if tool.Description() != expectedDesc {
		t.Errorf("Unexpected description: %s", tool.Description())
	}

	// Test InputSchema
	schema := tool.InputSchema()
	if schema["type"] != "object" {
		t.Error("Schema type should be 'object'")
	}

	properties := schema["properties"].(map[string]interface{})

	// Check that all expected properties are present
	expectedProps := []string{"name", "summary", "severity_id", "incident_type_id", "incident_status_id", "mode", "visibility"}
	for _, prop := range expectedProps {
		if _, ok := properties[prop]; !ok {
			t.Errorf("Schema should have '%s' property", prop)
		}
	}

	// Check that severity_id description mentions auto-selection
	severityProp := properties["severity_id"].(map[string]interface{})
	if !strings.Contains(severityProp["description"].(string), "auto-selected") {
		t.Error("severity_id description should mention auto-selection")
	}

	required := schema["required"].([]interface{})
	if len(required) != 1 || required[0] != "name" {
		t.Error("Schema should require only 'name'")
	}
}
