package client

import (
	"net/http"
	"testing"
)

func TestListWorkflows(t *testing.T) {
	tests := []struct {
		name           string
		params         *ListWorkflowsParams
		mockResponse   string
		mockStatusCode int
		wantError      bool
		expectedCount  int
	}{
		{
			name:   "successful list workflows",
			params: &ListWorkflowsParams{PageSize: 10},
			mockResponse: `{
				"workflows": [
					{
						"id": "wf_123",
						"name": "Test Workflow",
						"trigger": {
							"name": "incident.created",
							"label": "Incident created"
						},
						"enabled": true,
						"created_at": "2024-01-01T00:00:00Z",
						"updated_at": "2024-01-01T00:00:00Z"
					}
				],
				"pagination_info": {
					"page_size": 10
				}
			}`,
			mockStatusCode: http.StatusOK,
			wantError:      false,
			expectedCount:  1,
		},
		{
			name:           "empty workflows list",
			params:         nil,
			mockResponse:   `{"workflows": [], "pagination_info": {"page_size": 25}}`,
			mockStatusCode: http.StatusOK,
			wantError:      false,
			expectedCount:  0,
		},
		{
			name:           "API error",
			params:         &ListWorkflowsParams{PageSize: 10},
			mockResponse:   `{"error": "Internal server error"}`,
			mockStatusCode: http.StatusInternalServerError,
			wantError:      true,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// Verify request
					assertEqual(t, "GET", req.Method)
					assertEqual(t, "Bearer test-api-key", req.Header.Get("Authorization"))

					// Check query params if provided
					if tt.params != nil && tt.params.PageSize > 0 {
						assertEqual(t, "10", req.URL.Query().Get("page_size"))
					}

					return mockResponse(tt.mockStatusCode, tt.mockResponse), nil
				},
			}

			client := NewTestClient(mockClient)
			result, err := client.ListWorkflows(tt.params)

			if tt.wantError {
				assertError(t, err)
				return
			}

			assertNoError(t, err)
			if len(result.Workflows) != tt.expectedCount {
				t.Errorf("expected %d workflows, got %d", tt.expectedCount, len(result.Workflows))
			}

			if tt.expectedCount > 0 {
				assertEqual(t, "wf_123", result.Workflows[0].ID)
				assertEqual(t, "Test Workflow", result.Workflows[0].Name)
				// Verify trigger is properly parsed as an object
				assertEqual(t, "incident.created", result.Workflows[0].Trigger.Name)
				assertEqual(t, "Incident created", result.Workflows[0].Trigger.Label)
			}
		})
	}
}

func TestGetWorkflow(t *testing.T) {
	tests := []struct {
		name           string
		workflowID     string
		mockResponse   string
		mockStatusCode int
		wantError      bool
	}{
		{
			name:       "successful get workflow",
			workflowID: "wf_123",
			mockResponse: `{
				"workflow": {
					"id": "wf_123",
					"name": "Test Workflow",
					"trigger": {
						"name": "incident.created",
						"label": "Incident created"
					},
					"enabled": true,
					"state": "active",
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-01T00:00:00Z"
				}
			}`,
			mockStatusCode: http.StatusOK,
			wantError:      false,
		},
		{
			name:           "workflow not found",
			workflowID:     "wf_nonexistent",
			mockResponse:   `{"error": "Workflow not found"}`,
			mockStatusCode: http.StatusNotFound,
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					assertEqual(t, "GET", req.Method)
					assertEqual(t, "/workflows/"+tt.workflowID, req.URL.Path)
					return mockResponse(tt.mockStatusCode, tt.mockResponse), nil
				},
			}

			client := NewTestClient(mockClient)
			workflow, err := client.GetWorkflow(tt.workflowID)

			if tt.wantError {
				assertError(t, err)
				return
			}

			assertNoError(t, err)
			assertEqual(t, tt.workflowID, workflow.ID)
			assertEqual(t, "Test Workflow", workflow.Name)
			// Verify trigger is properly parsed as an object
			assertEqual(t, "incident.created", workflow.Trigger.Name)
			assertEqual(t, "Incident created", workflow.Trigger.Label)
			// Verify state is a string
			assertEqual(t, "active", workflow.State)
		})
	}
}

func TestUpdateWorkflow(t *testing.T) {
	tests := []struct {
		name           string
		workflowID     string
		request        *UpdateWorkflowRequest
		mockResponse   string
		mockStatusCode int
		wantError      bool
	}{
		{
			name:       "successful update workflow",
			workflowID: "wf_123",
			request: &UpdateWorkflowRequest{
				Name:    "Updated Workflow",
				Enabled: boolPtr(false),
			},
			mockResponse: `{
				"workflow": {
					"id": "wf_123",
					"name": "Updated Workflow",
					"trigger": {
						"name": "incident.created",
						"label": "Incident created"
					},
					"enabled": false,
					"state": "disabled",
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-02T00:00:00Z"
				}
			}`,
			mockStatusCode: http.StatusOK,
			wantError:      false,
		},
		{
			name:       "enable workflow",
			workflowID: "wf_123",
			request: &UpdateWorkflowRequest{
				Enabled: boolPtr(true),
			},
			mockResponse: `{
				"workflow": {
					"id": "wf_123",
					"name": "Test Workflow",
					"trigger": {
						"name": "incident.created",
						"label": "Incident created"
					},
					"enabled": true,
					"state": "active",
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-02T00:00:00Z"
				}
			}`,
			mockStatusCode: http.StatusOK,
			wantError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					assertEqual(t, "PATCH", req.Method)
					assertEqual(t, "/workflows/"+tt.workflowID, req.URL.Path)
					return mockResponse(tt.mockStatusCode, tt.mockResponse), nil
				},
			}

			client := NewTestClient(mockClient)
			workflow, err := client.UpdateWorkflow(tt.workflowID, tt.request)

			if tt.wantError {
				assertError(t, err)
				return
			}

			assertNoError(t, err)
			assertEqual(t, tt.workflowID, workflow.ID)

			// Verify updates were applied
			if tt.request.Name != "" {
				assertEqual(t, tt.request.Name, workflow.Name)
			}
			if tt.request.Enabled != nil {
				if workflow.Enabled != *tt.request.Enabled {
					t.Errorf("expected enabled to be %v, got %v", *tt.request.Enabled, workflow.Enabled)
				}
			}
		})
	}
}

// TestWorkflowTriggerObjectParsing is a regression test for bugs where:
// 1. workflow.trigger was incorrectly typed as string instead of an object
// 2. workflow.state was incorrectly typed as map[string]interface{} instead of string
// The incident.io API returns:
//   - trigger as: {"name": "...", "label": "..."}
//   - state as: "active" | "disabled" | etc.
// See: https://github.com/incident-io/incidentio-mcp-golang/pull/20
func TestWorkflowTriggerObjectParsing(t *testing.T) {
	// This test uses a real-world API response structure to ensure
	// we correctly parse the trigger object and state string
	mockResp := `{
		"workflows": [
			{
				"id": "wf_oncall_123",
				"name": "Notify on on-call change",
				"trigger": {
					"name": "schedule.currently-on-call-changed",
					"label": "An On-call schedule shift changes"
				},
				"enabled": true,
				"state": "active",
				"created_at": "2024-01-01T00:00:00Z",
				"updated_at": "2024-01-01T00:00:00Z"
			},
			{
				"id": "wf_incident_456",
				"name": "Auto-assign incident lead",
				"trigger": {
					"name": "incident.created",
					"label": "An incident is created"
				},
				"enabled": false,
				"state": "disabled",
				"created_at": "2024-01-01T00:00:00Z",
				"updated_at": "2024-01-01T00:00:00Z"
			}
		],
		"pagination_info": {
			"page_size": 25
		}
	}`

	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return mockResponse(http.StatusOK, mockResp), nil
		},
	}

	client := NewTestClient(mockClient)
	result, err := client.ListWorkflows(nil)

	assertNoError(t, err)
	if len(result.Workflows) != 2 {
		t.Fatalf("expected 2 workflows, got %d", len(result.Workflows))
	}

	// Verify first workflow (on-call schedule, active)
	wf1 := result.Workflows[0]
	assertEqual(t, "wf_oncall_123", wf1.ID)
	assertEqual(t, "schedule.currently-on-call-changed", wf1.Trigger.Name)
	assertEqual(t, "An On-call schedule shift changes", wf1.Trigger.Label)
	assertEqual(t, "active", wf1.State)

	// Verify second workflow (incident created, disabled)
	wf2 := result.Workflows[1]
	assertEqual(t, "wf_incident_456", wf2.ID)
	assertEqual(t, "incident.created", wf2.Trigger.Name)
	assertEqual(t, "An incident is created", wf2.Trigger.Label)
	assertEqual(t, "disabled", wf2.State)
}

// Helper function to create a bool pointer
func boolPtr(b bool) *bool {
	return &b
}
