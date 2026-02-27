package handlers

import (
	"testing"
)

// Test tools that were refactored to use helper functions
// These tests ensure the refactoring didn't break functionality

func TestListIncidentUpdatesTool_Schema(t *testing.T) {
	tool := &ListIncidentUpdatesTool{}

	// Test Name
	if tool.Name() != "list_incident_updates" {
		t.Errorf("Expected name 'list_incident_updates', got %s", tool.Name())
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
	if _, ok := properties["incident_id"]; !ok {
		t.Error("Schema should have 'incident_id' property")
	}
	if _, ok := properties["page_size"]; !ok {
		t.Error("Schema should have 'page_size' property")
	}
}

func TestGetIncidentUpdateTool_Schema(t *testing.T) {
	tool := &GetIncidentUpdateTool{}

	// Test Name
	if tool.Name() != "get_incident_update" {
		t.Errorf("Expected name 'get_incident_update', got %s", tool.Name())
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

	required := schema["required"].([]interface{})
	if len(required) != 1 || required[0] != "id" {
		t.Error("Schema should require only 'id'")
	}
}

func TestListWorkflowsTool_Schema(t *testing.T) {
	tool := &ListWorkflowsTool{}

	// Test Name
	if tool.Name() != "list_workflows" {
		t.Errorf("Expected name 'list_workflows', got %s", tool.Name())
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
	if _, ok := properties["page_size"]; !ok {
		t.Error("Schema should have 'page_size' property")
	}
	if _, ok := properties["after"]; !ok {
		t.Error("Schema should have 'after' property")
	}
}

func TestGetWorkflowTool_Schema(t *testing.T) {
	tool := &GetWorkflowTool{}

	// Test Name
	if tool.Name() != "get_workflow" {
		t.Errorf("Expected name 'get_workflow', got %s", tool.Name())
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

func TestListActionsTool_Schema(t *testing.T) {
	tool := &ListActionsTool{}

	// Test Name
	if tool.Name() != "list_actions" {
		t.Errorf("Expected name 'list_actions', got %s", tool.Name())
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
	expectedProps := []string{"page_size", "after", "incident_id", "status"}
	for _, prop := range expectedProps {
		if _, ok := properties[prop]; !ok {
			t.Errorf("Schema should have '%s' property", prop)
		}
	}
}

func TestGetActionTool_Schema(t *testing.T) {
	tool := &GetActionTool{}

	// Test Name
	if tool.Name() != "get_action" {
		t.Errorf("Expected name 'get_action', got %s", tool.Name())
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

func TestListAlertsTool_Schema(t *testing.T) {
	tool := &ListAlertsTool{}

	// Test Name
	if tool.Name() != "list_alerts" {
		t.Errorf("Expected name 'list_alerts', got %s", tool.Name())
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
	expectedProps := []string{"title", "page_size", "after", "status", "deduplication_key", "alert_source_id", "created_at_gte", "created_at_lte", "created_at_date_range"}
	for _, prop := range expectedProps {
		if _, ok := properties[prop]; !ok {
			t.Errorf("Schema should have '%s' property", prop)
		}
	}
}

func TestListIncidentStatusesTool_Schema(t *testing.T) {
	tool := &ListIncidentStatusesTool{}

	// Test Name
	if tool.Name() != "list_incident_statuses" {
		t.Errorf("Expected name 'list_incident_statuses', got %s", tool.Name())
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

	// This tool should have no properties
	properties := schema["properties"].(map[string]interface{})
	if len(properties) != 0 {
		t.Error("Schema should have no properties")
	}
}
