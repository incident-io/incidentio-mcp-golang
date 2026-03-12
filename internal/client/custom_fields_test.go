package client

import (
	"context"
	"net/http"
	"testing"
)

func TestListCustomFields(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   string
		mockStatusCode int
		wantError      bool
		expectedCount  int
	}{
		{
			name: "successful list custom fields",
			mockResponse: `{
				"custom_fields": [
					{
						"id": "cf_123",
						"name": "Priority",
						"description": "Incident priority level",
						"field_type": "single_select",
						"required": "never",
						"show_before_closure": false,
						"show_before_creation": true,
						"show_before_update": false,
						"options": [
							{
								"id": "opt_1",
								"value": "High",
								"sort_key": 1,
								"created_at": "2024-01-01T00:00:00Z",
								"updated_at": "2024-01-01T00:00:00Z"
							}
						],
						"created_at": "2024-01-01T00:00:00Z",
						"updated_at": "2024-01-01T00:00:00Z"
					}
				],
				"pagination_meta": {
					"page_size": 25
				}
			}`,
			mockStatusCode: http.StatusOK,
			wantError:      false,
			expectedCount:  1,
		},
		{
			name:           "empty custom fields list",
			mockResponse:   `{"custom_fields": [], "pagination_meta": {"page_size": 25}}`,
			mockStatusCode: http.StatusOK,
			wantError:      false,
			expectedCount:  0,
		},
		{
			name:           "API error",
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
					assertEqual(t, "GET", req.Method)
					assertEqual(t, "Bearer test-api-key", req.Header.Get("Authorization"))
					return mockResponse(tt.mockStatusCode, tt.mockResponse), nil
				},
			}

			client := NewTestClient(mockClient)
			result, err := client.ListCustomFields(context.Background())

			if tt.wantError {
				assertError(t, err)
				return
			}

			assertNoError(t, err)
			if len(result.CustomFields) != tt.expectedCount {
				t.Errorf("expected %d custom fields, got %d", tt.expectedCount, len(result.CustomFields))
			}

			if tt.expectedCount > 0 {
				assertEqual(t, "cf_123", result.CustomFields[0].ID)
				assertEqual(t, "Priority", result.CustomFields[0].Name)
				assertEqual(t, "single_select", result.CustomFields[0].FieldType)
			}
		})
	}
}

func TestGetCustomField(t *testing.T) {
	tests := []struct {
		name           string
		fieldID        string
		mockResponse   string
		mockStatusCode int
		wantError      bool
	}{
		{
			name:    "successful get custom field",
			fieldID: "cf_123",
			mockResponse: `{
				"custom_field": {
					"id": "cf_123",
					"name": "Priority",
					"description": "Incident priority level",
					"field_type": "single_select",
					"required": "before_closure",
					"show_before_closure": true,
					"show_before_creation": true,
					"show_before_update": false,
					"options": [
						{
							"id": "opt_1",
							"value": "High",
							"sort_key": 1,
							"created_at": "2024-01-01T00:00:00Z",
							"updated_at": "2024-01-01T00:00:00Z"
						},
						{
							"id": "opt_2",
							"value": "Low",
							"sort_key": 2,
							"created_at": "2024-01-01T00:00:00Z",
							"updated_at": "2024-01-01T00:00:00Z"
						}
					],
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-01T00:00:00Z"
				}
			}`,
			mockStatusCode: http.StatusOK,
			wantError:      false,
		},
		{
			name:           "custom field not found",
			fieldID:        "cf_nonexistent",
			mockResponse:   `{"error": "Custom field not found"}`,
			mockStatusCode: http.StatusNotFound,
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					assertEqual(t, "GET", req.Method)
					assertEqual(t, "/custom_fields/"+tt.fieldID, req.URL.Path)
					return mockResponse(tt.mockStatusCode, tt.mockResponse), nil
				},
			}

			client := NewTestClient(mockClient)
			field, err := client.GetCustomField(context.Background(), tt.fieldID)

			if tt.wantError {
				assertError(t, err)
				return
			}

			assertNoError(t, err)
			assertEqual(t, tt.fieldID, field.ID)
			assertEqual(t, "Priority", field.Name)
			assertEqual(t, "single_select", field.FieldType)
			if len(field.Options) != 2 {
				t.Errorf("expected 2 options, got %d", len(field.Options))
			}
		})
	}
}

func TestCreateCustomField(t *testing.T) {
	tests := []struct {
		name           string
		request        *CreateCustomFieldRequest
		mockResponse   string
		mockStatusCode int
		wantError      bool
	}{
		{
			name: "successful create text custom field",
			request: &CreateCustomFieldRequest{
				Name:               "Impact Summary",
				Description:        "Summary of incident impact",
				FieldType:          "text",
				Required:           "never",
				ShowBeforeClosure:  false,
				ShowBeforeCreation: true,
				ShowBeforeUpdate:   false,
			},
			mockResponse: `{
				"custom_field": {
					"id": "cf_456",
					"name": "Impact Summary",
					"description": "Summary of incident impact",
					"field_type": "text",
					"required": "never",
					"show_before_closure": false,
					"show_before_creation": true,
					"show_before_update": false,
					"options": [],
					"created_at": "2024-01-02T00:00:00Z",
					"updated_at": "2024-01-02T00:00:00Z"
				}
			}`,
			mockStatusCode: http.StatusCreated,
			wantError:      false,
		},
		{
			name: "successful create select field with options",
			request: &CreateCustomFieldRequest{
				Name:               "Severity Level",
				Description:        "Incident severity",
				FieldType:          "single_select",
				Required:           "always",
				ShowBeforeClosure:  true,
				ShowBeforeCreation: true,
				ShowBeforeUpdate:   true,
				Options:            []string{"Critical", "High", "Medium", "Low"},
			},
			mockResponse: `{
				"custom_field": {
					"id": "cf_789",
					"name": "Severity Level",
					"description": "Incident severity",
					"field_type": "single_select",
					"required": "always",
					"show_before_closure": true,
					"show_before_creation": true,
					"show_before_update": true,
					"options": [
						{"id": "opt_1", "value": "Critical", "sort_key": 1, "created_at": "2024-01-02T00:00:00Z", "updated_at": "2024-01-02T00:00:00Z"},
						{"id": "opt_2", "value": "High", "sort_key": 2, "created_at": "2024-01-02T00:00:00Z", "updated_at": "2024-01-02T00:00:00Z"}
					],
					"created_at": "2024-01-02T00:00:00Z",
					"updated_at": "2024-01-02T00:00:00Z"
				}
			}`,
			mockStatusCode: http.StatusCreated,
			wantError:      false,
		},
		{
			name: "validation error",
			request: &CreateCustomFieldRequest{
				Name:        "",
				Description: "Missing name",
				FieldType:   "text",
				Required:    "never",
			},
			mockResponse:   `{"error": "Name is required"}`,
			mockStatusCode: http.StatusBadRequest,
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					assertEqual(t, "POST", req.Method)
					assertEqual(t, "/custom_fields", req.URL.Path)
					return mockResponse(tt.mockStatusCode, tt.mockResponse), nil
				},
			}

			client := NewTestClient(mockClient)
			field, err := client.CreateCustomField(context.Background(), tt.request)

			if tt.wantError {
				assertError(t, err)
				return
			}

			assertNoError(t, err)
			assertEqual(t, tt.request.Name, field.Name)
			assertEqual(t, tt.request.FieldType, field.FieldType)
		})
	}
}

func TestUpdateCustomField(t *testing.T) {
	tests := []struct {
		name           string
		fieldID        string
		request        *UpdateCustomFieldRequest
		mockResponse   string
		mockStatusCode int
		wantError      bool
	}{
		{
			name:    "successful update custom field name",
			fieldID: "cf_123",
			request: &UpdateCustomFieldRequest{
				Name:        "Updated Priority",
				Description: "Updated description",
			},
			mockResponse: `{
				"custom_field": {
					"id": "cf_123",
					"name": "Updated Priority",
					"description": "Updated description",
					"field_type": "single_select",
					"required": "never",
					"show_before_closure": false,
					"show_before_creation": true,
					"show_before_update": false,
					"options": [],
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-03T00:00:00Z"
				}
			}`,
			mockStatusCode: http.StatusOK,
			wantError:      false,
		},
		{
			name:    "update field requirements",
			fieldID: "cf_123",
			request: &UpdateCustomFieldRequest{
				Required:          "before_closure",
				ShowBeforeClosure: boolPtr(true),
			},
			mockResponse: `{
				"custom_field": {
					"id": "cf_123",
					"name": "Priority",
					"description": "Incident priority level",
					"field_type": "single_select",
					"required": "before_closure",
					"show_before_closure": true,
					"show_before_creation": true,
					"show_before_update": false,
					"options": [],
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-03T00:00:00Z"
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
					assertEqual(t, "PUT", req.Method)
					assertEqual(t, "/custom_fields/"+tt.fieldID, req.URL.Path)
					return mockResponse(tt.mockStatusCode, tt.mockResponse), nil
				},
			}

			client := NewTestClient(mockClient)
			field, err := client.UpdateCustomField(context.Background(), tt.fieldID, tt.request)

			if tt.wantError {
				assertError(t, err)
				return
			}

			assertNoError(t, err)
			assertEqual(t, tt.fieldID, field.ID)
		})
	}
}

func TestDeleteCustomField(t *testing.T) {
	tests := []struct {
		name           string
		fieldID        string
		mockStatusCode int
		wantError      bool
	}{
		{
			name:           "successful delete",
			fieldID:        "cf_123",
			mockStatusCode: http.StatusNoContent,
			wantError:      false,
		},
		{
			name:           "custom field not found",
			fieldID:        "cf_nonexistent",
			mockStatusCode: http.StatusNotFound,
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					assertEqual(t, "DELETE", req.Method)
					assertEqual(t, "/custom_fields/"+tt.fieldID, req.URL.Path)
					body := ""
					if tt.mockStatusCode >= 400 {
						body = `{"error": "Not found"}`
					}
					return mockResponse(tt.mockStatusCode, body), nil
				},
			}

			client := NewTestClient(mockClient)
			err := client.DeleteCustomField(context.Background(), tt.fieldID)

			if tt.wantError {
				assertError(t, err)
				return
			}

			assertNoError(t, err)
		})
	}
}

func TestListCustomFieldOptions(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   string
		mockStatusCode int
		wantError      bool
		expectedCount  int
	}{
		{
			name: "successful list custom field options",
			mockResponse: `{
				"custom_field_options": [
					{
						"id": "opt_1",
						"value": "High",
						"sort_key": 1,
						"created_at": "2024-01-01T00:00:00Z",
						"updated_at": "2024-01-01T00:00:00Z"
					},
					{
						"id": "opt_2",
						"value": "Low",
						"sort_key": 2,
						"created_at": "2024-01-01T00:00:00Z",
						"updated_at": "2024-01-01T00:00:00Z"
					}
				]
			}`,
			mockStatusCode: http.StatusOK,
			wantError:      false,
			expectedCount:  2,
		},
		{
			name:           "empty options list",
			mockResponse:   `{"custom_field_options": []}`,
			mockStatusCode: http.StatusOK,
			wantError:      false,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					assertEqual(t, "GET", req.Method)
					return mockResponse(tt.mockStatusCode, tt.mockResponse), nil
				},
			}

			client := NewTestClient(mockClient)
			options, err := client.ListCustomFieldOptions(context.Background())

			if tt.wantError {
				assertError(t, err)
				return
			}

			assertNoError(t, err)
			if len(options) != tt.expectedCount {
				t.Errorf("expected %d options, got %d", tt.expectedCount, len(options))
			}
		})
	}
}

func TestCreateCustomFieldOption(t *testing.T) {
	tests := []struct {
		name           string
		request        *CreateCustomFieldOptionRequest
		mockResponse   string
		mockStatusCode int
		wantError      bool
	}{
		{
			name: "successful create option",
			request: &CreateCustomFieldOptionRequest{
				CustomFieldID: "cf_123",
				Value:         "Medium",
				SortKey:       3,
			},
			mockResponse: `{
				"custom_field_option": {
					"id": "opt_3",
					"value": "Medium",
					"sort_key": 3,
					"created_at": "2024-01-02T00:00:00Z",
					"updated_at": "2024-01-02T00:00:00Z"
				}
			}`,
			mockStatusCode: http.StatusCreated,
			wantError:      false,
		},
		{
			name: "validation error",
			request: &CreateCustomFieldOptionRequest{
				CustomFieldID: "cf_123",
				Value:         "",
			},
			mockResponse:   `{"error": "Value is required"}`,
			mockStatusCode: http.StatusBadRequest,
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					assertEqual(t, "POST", req.Method)
					assertEqual(t, "/custom_field_options", req.URL.Path)
					return mockResponse(tt.mockStatusCode, tt.mockResponse), nil
				},
			}

			client := NewTestClient(mockClient)
			option, err := client.CreateCustomFieldOption(context.Background(), tt.request)

			if tt.wantError {
				assertError(t, err)
				return
			}

			assertNoError(t, err)
			assertEqual(t, tt.request.Value, option.Value)
		})
	}
}

func TestSearchCustomFields(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		fieldType      string
		mockResponse   string
		mockStatusCode int
		wantError      bool
		expectedCount  int
	}{
		{
			name:      "search by query",
			query:     "Priority",
			fieldType: "",
			mockResponse: `{
				"custom_fields": [
					{
						"id": "cf_123",
						"name": "Priority",
						"description": "Priority level",
						"field_type": "single_select",
						"required": "never",
						"show_before_closure": false,
						"show_before_creation": true,
						"show_before_update": false,
						"options": [],
						"created_at": "2024-01-01T00:00:00Z",
						"updated_at": "2024-01-01T00:00:00Z"
					}
				],
				"pagination_meta": {"page_size": 25}
			}`,
			mockStatusCode: http.StatusOK,
			wantError:      false,
			expectedCount:  1,
		},
		{
			name:      "filter by field type",
			query:     "",
			fieldType: "text",
			mockResponse: `{
				"custom_fields": [
					{
						"id": "cf_456",
						"name": "Impact Summary",
						"description": "Summary text",
						"field_type": "text",
						"required": "never",
						"show_before_closure": false,
						"show_before_creation": true,
						"show_before_update": false,
						"options": [],
						"created_at": "2024-01-01T00:00:00Z",
						"updated_at": "2024-01-01T00:00:00Z"
					}
				],
				"pagination_meta": {"page_size": 25}
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
					if tt.query != "" {
						assertEqual(t, tt.query, req.URL.Query().Get("query"))
					}
					if tt.fieldType != "" {
						assertEqual(t, tt.fieldType, req.URL.Query().Get("field_type"))
					}
					return mockResponse(tt.mockStatusCode, tt.mockResponse), nil
				},
			}

			client := NewTestClient(mockClient)
			fields, err := client.SearchCustomFields(context.Background(), tt.query, tt.fieldType)

			if tt.wantError {
				assertError(t, err)
				return
			}

			assertNoError(t, err)
			if len(fields) != tt.expectedCount {
				t.Errorf("expected %d fields, got %d", tt.expectedCount, len(fields))
			}
		})
	}
}
