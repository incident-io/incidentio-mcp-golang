package client

import (
	"net/http"
	"testing"
)

func TestListIncidents(t *testing.T) {
	tests := []struct {
		name           string
		params         *ListIncidentsOptions
		mockResponse   string
		mockStatusCode int
		wantError      bool
		expectedCount  int
	}{
		{
			name: "successful list incidents",
			params: &ListIncidentsOptions{
				PageSize: 10,
				Status:   []string{"active", "resolved"},
			},
			mockResponse: `{
				"incidents": [
					{
						"id": "inc_123",
						"reference": "INC-123",
						"name": "Database outage",
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
			name:           "empty incidents list",
			params:         nil,
			mockResponse:   `{"incidents": [], "pagination_info": {"page_size": 25}}`,
			mockStatusCode: http.StatusOK,
			wantError:      false,
			expectedCount:  0,
		},
		{
			name: "filter by incident type",
			params: &ListIncidentsOptions{
				IncidentTypeOneOf: []string{"type_security"},
			},
			mockResponse: `{
				"incidents": [
					{
						"id": "inc_sec",
						"reference": "INC-100",
						"name": "Security incident",
						"incident_status": {
							"id": "status_active",
							"name": "Active"
						},
						"severity": {
							"id": "sev_1",
							"name": "Critical"
						},
						"incident_type": {
							"id": "type_security",
							"name": "Security"
						},
						"created_at": "2024-01-01T00:00:00Z",
						"updated_at": "2024-01-01T01:00:00Z"
					}
				],
				"pagination_info": {
					"page_size": 25
				}
			}`,
			mockStatusCode: http.StatusOK,
			wantError:      false,
			expectedCount:  1,
		},
		{
			name: "filter by severity",
			params: &ListIncidentsOptions{
				Severity: []string{"sev_1", "sev_2"},
			},
			mockResponse: `{
				"incidents": [
					{
						"id": "inc_456",
						"reference": "INC-456",
						"name": "API performance degradation",
						"incident_status": {
							"id": "status_investigating",
							"name": "Investigating"
						},
						"severity": {
							"id": "sev_2",
							"name": "High"
						},
						"created_at": "2024-01-02T00:00:00Z",
						"updated_at": "2024-01-02T00:30:00Z"
					}
				],
				"pagination_info": {
					"page_size": 25
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

					// Check query parameters
					if tt.params != nil {
						if tt.params.PageSize > 0 {
							assertEqual(t, "10", req.URL.Query().Get("page_size"))
						}
						if len(tt.params.Status) > 0 {
							// Status params should be present with bracket notation
							statusValues := req.URL.Query()["status[one_of]"]
							if len(statusValues) != len(tt.params.Status) {
								t.Errorf("expected %d status values, got %d", len(tt.params.Status), len(statusValues))
							}
						}
						if len(tt.params.IncidentTypeOneOf) > 0 {
							typeValues := req.URL.Query()["incident_type[one_of]"]
							if len(typeValues) != len(tt.params.IncidentTypeOneOf) {
								t.Errorf("expected %d incident_type values, got %d", len(tt.params.IncidentTypeOneOf), len(typeValues))
							}
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

			if tt.expectedCount > 0 {
				incident := result.Incidents[0]
				switch tt.name {
				case "successful list incidents":
					assertEqual(t, "inc_123", incident.ID)
					assertEqual(t, "INC-123", incident.Reference)
					assertEqual(t, "Database outage", incident.Name)
				case "filter by incident type":
					assertEqual(t, "inc_sec", incident.ID)
					assertEqual(t, "type_security", incident.IncidentType.ID)
				case "filter by severity":
					assertEqual(t, "inc_456", incident.ID)
					assertEqual(t, "sev_2", incident.Severity.ID)
				}
			}
		})
	}
}

func TestGetIncident(t *testing.T) {
	tests := []struct {
		name           string
		incidentID     string
		mockResponse   string
		mockStatusCode int
		wantError      bool
	}{
		{
			name:       "successful get incident",
			incidentID: "inc_123",
			mockResponse: `{
				"incident": {
					"id": "inc_123",
					"reference": "INC-123",
					"name": "Database outage",
					"summary": "Primary database cluster is experiencing connectivity issues",
					"incident_status": {
						"id": "status_active",
						"name": "Active"
					},
					"severity": {
						"id": "sev_1",
						"name": "Critical"
					},
					"incident_role_assignments": [
						{
							"role": {
								"id": "role_commander",
								"name": "Incident Commander"
							},
							"assignee": {
								"id": "user_123",
								"name": "John Doe",
								"email": "john@example.com"
							}
						}
					],
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-01T01:00:00Z"
				}
			}`,
			mockStatusCode: http.StatusOK,
			wantError:      false,
		},
		{
			name:           "incident not found",
			incidentID:     "inc_nonexistent",
			mockResponse:   `{"error": "Incident not found"}`,
			mockStatusCode: http.StatusNotFound,
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					assertEqual(t, "GET", req.Method)
					assertEqual(t, "/incidents/"+tt.incidentID, req.URL.Path)
					return mockResponse(tt.mockStatusCode, tt.mockResponse), nil
				},
			}

			client := NewTestClient(mockClient)
			incident, err := client.GetIncident(tt.incidentID)

			if tt.wantError {
				assertError(t, err)
				return
			}

			assertNoError(t, err)
			assertEqual(t, tt.incidentID, incident.ID)
			assertEqual(t, "INC-123", incident.Reference)
			assertEqual(t, "Database outage", incident.Name)

			// Verify role assignments
			if len(incident.IncidentRoleAssignments) > 0 {
				assignment := incident.IncidentRoleAssignments[0]
				assertEqual(t, "role_commander", assignment.Role.ID)
				assertEqual(t, "user_123", assignment.Assignee.ID)
			}
		})
	}
}

func TestCreateIncident(t *testing.T) {
	tests := []struct {
		name           string
		request        *CreateIncidentRequest
		mockResponse   string
		mockStatusCode int
		wantError      bool
	}{
		{
			name: "successful create incident",
			request: &CreateIncidentRequest{
				Name:    "New production issue",
				Summary: "Users reporting 500 errors on checkout",
			},
			mockResponse: `{
				"incident": {
					"id": "inc_new",
					"reference": "INC-789",
					"name": "New production issue",
					"summary": "Users reporting 500 errors on checkout",
					"incident_status": {
						"id": "status_triage",
						"name": "Triage"
					},
					"severity": {
						"id": "sev_3",
						"name": "Medium"
					},
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-01T00:00:00Z"
				}
			}`,
			mockStatusCode: http.StatusCreated,
			wantError:      false,
		},
		{
			name: "create with severity and status",
			request: &CreateIncidentRequest{
				Name:             "Critical system failure",
				Summary:          "Complete system outage affecting all users",
				IncidentStatusID: "status_active",
				SeverityID:       "sev_1",
			},
			mockResponse: `{
				"incident": {
					"id": "inc_critical",
					"reference": "INC-999",
					"name": "Critical system failure",
					"summary": "Complete system outage affecting all users",
					"incident_status": {
						"id": "status_active",
						"name": "Active"
					},
					"severity": {
						"id": "sev_1",
						"name": "Critical"
					},
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-01T00:00:00Z"
				}
			}`,
			mockStatusCode: http.StatusCreated,
			wantError:      false,
		},
		{
			name: "create with role assignments",
			request: &CreateIncidentRequest{
				Name:    "Security incident",
				Summary: "Potential data breach detected",
				IncidentRoleAssignments: []CreateRoleAssignmentRequest{
					{
						IncidentRoleID: "role_commander",
						UserID:         "user_456",
					},
				},
			},
			mockResponse: `{
				"incident": {
					"id": "inc_security",
					"reference": "INC-1001",
					"name": "Security incident",
					"summary": "Potential data breach detected",
					"incident_status": {
						"id": "status_triage",
						"name": "Triage"
					},
					"incident_role_assignments": [
						{
							"role": {
								"id": "role_commander",
								"name": "Incident Commander"
							},
							"assignee": {
								"id": "user_456",
								"name": "Jane Smith",
								"email": "jane@example.com"
							}
						}
					],
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-01T00:00:00Z"
				}
			}`,
			mockStatusCode: http.StatusCreated,
			wantError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					assertEqual(t, "POST", req.Method)
					assertEqual(t, "/incidents", req.URL.Path)
					return mockResponse(tt.mockStatusCode, tt.mockResponse), nil
				},
			}

			client := NewTestClient(mockClient)
			incident, err := client.CreateIncident(tt.request)

			if tt.wantError {
				assertError(t, err)
				return
			}

			assertNoError(t, err)
			assertEqual(t, tt.request.Name, incident.Name)
			assertEqual(t, tt.request.Summary, incident.Summary)

			// Verify severity and status if specified
			if tt.request.SeverityID != "" {
				assertEqual(t, tt.request.SeverityID, incident.Severity.ID)
			}
			if tt.request.IncidentStatusID != "" {
				assertEqual(t, tt.request.IncidentStatusID, incident.IncidentStatus.ID)
			}
		})
	}
}

func TestUpdateIncident(t *testing.T) {
	tests := []struct {
		name           string
		incidentID     string
		request        *UpdateIncidentRequest
		mockResponse   string
		mockStatusCode int
		wantError      bool
	}{
		{
			name:       "successful update incident",
			incidentID: "inc_123",
			request: &UpdateIncidentRequest{
				Name:    "Updated incident name",
				Summary: "Updated summary with more details",
			},
			mockResponse: `{
				"incident": {
					"id": "inc_123",
					"reference": "INC-123",
					"name": "Updated incident name",
					"summary": "Updated summary with more details",
					"incident_status": {
						"id": "status_active",
						"name": "Active"
					},
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-01T02:00:00Z"
				}
			}`,
			mockStatusCode: http.StatusOK,
			wantError:      false,
		},
		{
			name:       "update status",
			incidentID: "inc_123",
			request: &UpdateIncidentRequest{
				IncidentStatusID: "status_resolved",
			},
			mockResponse: `{
				"incident": {
					"id": "inc_123",
					"reference": "INC-123",
					"name": "Database outage",
					"incident_status": {
						"id": "status_resolved",
						"name": "Resolved"
					},
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-01T03:00:00Z"
				}
			}`,
			mockStatusCode: http.StatusOK,
			wantError:      false,
		},
		{
			name:       "update severity",
			incidentID: "inc_456",
			request: &UpdateIncidentRequest{
				SeverityID: "sev_1",
			},
			mockResponse: `{
				"incident": {
					"id": "inc_456",
					"reference": "INC-456",
					"name": "API issue",
					"severity": {
						"id": "sev_1",
						"name": "Critical"
					},
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-01T04:00:00Z"
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
					assertEqual(t, "POST", req.Method)
					assertEqual(t, "/incidents/"+tt.incidentID+"/actions/edit", req.URL.Path)
					return mockResponse(tt.mockStatusCode, tt.mockResponse), nil
				},
			}

			client := NewTestClient(mockClient)
			incident, err := client.UpdateIncident(tt.incidentID, tt.request)

			if tt.wantError {
				assertError(t, err)
				return
			}

			assertNoError(t, err)
			assertEqual(t, tt.incidentID, incident.ID)

			// Verify updates
			if tt.request.Name != "" {
				assertEqual(t, tt.request.Name, incident.Name)
			}
			if tt.request.Summary != "" {
				assertEqual(t, tt.request.Summary, incident.Summary)
			}
			if tt.request.IncidentStatusID != "" {
				assertEqual(t, tt.request.IncidentStatusID, incident.IncidentStatus.ID)
			}
			if tt.request.SeverityID != "" {
				assertEqual(t, tt.request.SeverityID, incident.Severity.ID)
			}
		})
	}
}
