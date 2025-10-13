package incidentio

import (
	"net/http"
	"testing"
)

// TestListIncidentsSeverityFilter specifically tests the severity filter functionality
// to ensure it uses the correct API parameter name: severity[one_of]
func TestListIncidentsSeverityFilter(t *testing.T) {
	tests := []struct {
		name                  string
		params                *ListIncidentsOptions
		expectedSeverityParam []string
		mockResponse          string
		mockStatusCode        int
		wantError             bool
		expectedCount         int
	}{
		{
			name: "single severity filter",
			params: &ListIncidentsOptions{
				PageSize: 10,
				Severity: []string{"sev_1"},
			},
			expectedSeverityParam: []string{"sev_1"},
			mockResponse: `{
				"incidents": [
					{
						"id": "inc_123",
						"reference": "INC-123",
						"name": "Critical database outage",
						"incident_status": {
							"id": "status_active",
							"name": "Active"
						},
						"severity": {
							"id": "sev_1",
							"name": "Critical"
						},
						"created_at": "2024-01-01T00:00:00Z",
						"updated_at": "2024-01-01T01:00:00Z"
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
			name: "multiple severity filter",
			params: &ListIncidentsOptions{
				PageSize: 25,
				Severity: []string{"sev_1", "sev_2", "sev_3"},
			},
			expectedSeverityParam: []string{"sev_1", "sev_2", "sev_3"},
			mockResponse: `{
				"incidents": [
					{
						"id": "inc_456",
						"reference": "INC-456",
						"name": "API performance degradation",
						"severity": {
							"id": "sev_2",
							"name": "High"
						},
						"created_at": "2024-01-02T00:00:00Z",
						"updated_at": "2024-01-02T00:30:00Z"
					},
					{
						"id": "inc_789",
						"reference": "INC-789",
						"name": "Database connection issues",
						"severity": {
							"id": "sev_1",
							"name": "Critical"
						},
						"created_at": "2024-01-03T00:00:00Z",
						"updated_at": "2024-01-03T00:15:00Z"
					}
				],
				"pagination_info": {
					"page_size": 25
				}
			}`,
			mockStatusCode: http.StatusOK,
			wantError:      false,
			expectedCount:  2,
		},
		{
			name: "severity and status combined filter",
			params: &ListIncidentsOptions{
				PageSize: 50,
				Status:   []string{"active", "triage"},
				Severity: []string{"sev_1", "sev_2"},
			},
			expectedSeverityParam: []string{"sev_1", "sev_2"},
			mockResponse: `{
				"incidents": [
					{
						"id": "inc_999",
						"reference": "INC-999",
						"name": "High priority incident",
						"incident_status": {
							"id": "status_active",
							"name": "Active"
						},
						"severity": {
							"id": "sev_2",
							"name": "High"
						},
						"created_at": "2024-01-04T00:00:00Z",
						"updated_at": "2024-01-04T00:30:00Z"
					}
				],
				"pagination_info": {
					"page_size": 50
				}
			}`,
			mockStatusCode: http.StatusOK,
			wantError:      false,
			expectedCount:  1,
		},
		{
			name: "severity filter with auto-pagination",
			params: &ListIncidentsOptions{
				Severity: []string{"sev_3"},
			},
			expectedSeverityParam: []string{"sev_3"},
			mockResponse: `{
				"incidents": [
					{
						"id": "inc_100",
						"reference": "INC-100",
						"name": "Medium severity incident",
						"severity": {
							"id": "sev_3",
							"name": "Medium"
						},
						"created_at": "2024-01-05T00:00:00Z",
						"updated_at": "2024-01-05T00:30:00Z"
					}
				],
				"pagination_info": {
					"page_size": 250,
					"after": ""
				}
			}`,
			mockStatusCode: http.StatusOK,
			wantError:      false,
			expectedCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					assertEqual(t, "GET", req.Method)
					assertEqual(t, "Bearer test-api-key", req.Header.Get("Authorization"))

					// Verify severity[one_of] parameter is used correctly
					if tt.expectedSeverityParam != nil {
						severityValues := req.URL.Query()["severity[one_of]"]
						if len(severityValues) != len(tt.expectedSeverityParam) {
							t.Errorf("expected %d severity[one_of] values, got %d", len(tt.expectedSeverityParam), len(severityValues))
						}

						// Verify each expected severity is present
						for i, expected := range tt.expectedSeverityParam {
							if i < len(severityValues) && severityValues[i] != expected {
								t.Errorf("expected severity[one_of][%d]=%s, got %s", i, expected, severityValues[i])
							}
						}

						// Ensure old "severity" parameter is NOT used
						if oldSeverity := req.URL.Query()["severity"]; len(oldSeverity) > 0 {
							t.Errorf("found deprecated 'severity' parameter, should use 'severity[one_of]' instead")
						}
					}

					// Verify status parameter if present
					if tt.params != nil && len(tt.params.Status) > 0 {
						statusValues := req.URL.Query()["status"]
						if len(statusValues) != len(tt.params.Status) {
							t.Errorf("expected %d status values, got %d", len(tt.params.Status), len(statusValues))
						}
					}

					return mockResponse(tt.mockStatusCode, tt.mockResponse), nil
				},
			}

			client := NewTestClient(mockClient)
			result, err := client.ListIncidents(tt.params)

			if tt.wantError {
				assertError(t, err)
				return
			}

			assertNoError(t, err)
			if len(result.Incidents) != tt.expectedCount {
				t.Errorf("expected %d incidents, got %d", tt.expectedCount, len(result.Incidents))
			}

			// Verify returned incidents match requested severity filter
			if tt.expectedCount > 0 && tt.params != nil && len(tt.params.Severity) > 0 {
				for _, incident := range result.Incidents {
					found := false
					for _, severity := range tt.params.Severity {
						if incident.Severity.ID == severity {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("incident %s has severity %s which is not in filter %v",
							incident.ID, incident.Severity.ID, tt.params.Severity)
					}
				}
			}
		})
	}
}

// TestListIncidentsSeverityFilterErrors tests error scenarios with severity filtering
func TestListIncidentsSeverityFilterErrors(t *testing.T) {
	tests := []struct {
		name           string
		params         *ListIncidentsOptions
		mockResponse   string
		mockStatusCode int
		wantError      bool
		errorContains  string
	}{
		{
			name: "invalid severity ID returns API error",
			params: &ListIncidentsOptions{
				Severity: []string{"invalid_severity_id"},
			},
			mockResponse:   `{"type":"validation_error","status":422,"message":"Invalid severity ID"}`,
			mockStatusCode: http.StatusUnprocessableEntity,
			wantError:      true,
			errorContains:  "422",
		},
		{
			name: "malformed severity ID",
			params: &ListIncidentsOptions{
				Severity: []string{"SEV-1"}, // Wrong format
			},
			mockResponse:   `{"type":"validation_error","status":422,"message":"Severity ID format invalid"}`,
			mockStatusCode: http.StatusUnprocessableEntity,
			wantError:      true,
			errorContains:  "422",
		},
		{
			name: "empty severity returns all incidents",
			params: &ListIncidentsOptions{
				PageSize: 10,
				Severity: []string{},
			},
			mockResponse: `{
				"incidents": [
					{
						"id": "inc_all",
						"reference": "INC-ALL",
						"name": "Any severity incident",
						"severity": {
							"id": "sev_4",
							"name": "Low"
						},
						"created_at": "2024-01-06T00:00:00Z",
						"updated_at": "2024-01-06T00:30:00Z"
					}
				],
				"pagination_info": {
					"page_size": 10
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
					// Verify severity[one_of] parameter is used even for invalid values
					if len(tt.params.Severity) > 0 {
						severityValues := req.URL.Query()["severity[one_of]"]
						if len(severityValues) != len(tt.params.Severity) {
							t.Errorf("expected %d severity[one_of] values, got %d", len(tt.params.Severity), len(severityValues))
						}
					}

					return mockResponse(tt.mockStatusCode, tt.mockResponse), nil
				},
			}

			client := NewTestClient(mockClient)
			_, err := client.ListIncidents(tt.params)

			if tt.wantError {
				assertError(t, err)
				if tt.errorContains != "" && err != nil {
					if !contains(err.Error(), tt.errorContains) {
						t.Errorf("expected error to contain %q, got %q", tt.errorContains, err.Error())
					}
				}
			} else {
				assertNoError(t, err)
			}
		})
	}
}

// TestListIncidentsSeverityParameterMigration verifies the fix migrated from old parameter to new
func TestListIncidentsSeverityParameterMigration(t *testing.T) {
	t.Run("verify severity[one_of] is used not severity", func(t *testing.T) {
		params := &ListIncidentsOptions{
			PageSize: 10,
			Severity: []string{"sev_1", "sev_2"},
		}

		parameterUsed := false
		deprecatedParameterUsed := false

		mockClient := &MockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				// Check for correct parameter
				if values := req.URL.Query()["severity[one_of]"]; len(values) > 0 {
					parameterUsed = true
				}

				// Check for deprecated parameter
				if values := req.URL.Query()["severity"]; len(values) > 0 {
					deprecatedParameterUsed = true
				}

				return mockResponse(http.StatusOK, `{
					"incidents": [],
					"pagination_info": {"page_size": 10}
				}`), nil
			},
		}

		client := NewTestClient(mockClient)
		_, err := client.ListIncidents(params)

		assertNoError(t, err)

		if !parameterUsed {
			t.Error("Expected 'severity[one_of]' parameter to be used, but it was not")
		}

		if deprecatedParameterUsed {
			t.Error("Found deprecated 'severity' parameter, should only use 'severity[one_of]'")
		}
	})
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		   (len(s) > len(substr) && contains(s[1:], substr))
}
