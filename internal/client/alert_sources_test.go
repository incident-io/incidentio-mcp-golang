package client

import (
	"context"
	"net/http"
	"testing"
)

func TestListAlertSources(t *testing.T) {
	tests := []struct {
		name           string
		params         *ListAlertSourcesParams
		mockResponse   string
		mockStatusCode int
		wantError      bool
		expectedCount  int
	}{
		{
			name:   "successful list alert sources",
			params: &ListAlertSourcesParams{PageSize: 10},
			mockResponse: `{
				"alert_sources": [
					{
						"id": "as_123",
						"name": "Production Monitoring",
						"type": "http",
						"config_type": "webhook",
						"created_at": "2024-01-01T00:00:00Z",
						"updated_at": "2024-01-01T00:00:00Z"
					},
					{
						"id": "as_456",
						"name": "Datadog Integration",
						"type": "datadog",
						"config_type": "integration",
						"created_at": "2024-01-02T00:00:00Z",
						"updated_at": "2024-01-02T00:00:00Z"
					}
				],
				"pagination_info": {
					"page_size": 10
				}
			}`,
			mockStatusCode: http.StatusOK,
			wantError:      false,
			expectedCount:  2,
		},
		{
			name:           "empty alert sources",
			params:         nil,
			mockResponse:   `{"alert_sources": [], "pagination_info": {"page_size": 25}}`,
			mockStatusCode: http.StatusOK,
			wantError:      false,
			expectedCount:  0,
		},
		{
			name:   "with pagination",
			params: &ListAlertSourcesParams{PageSize: 5, After: "cursor_123"},
			mockResponse: `{
				"alert_sources": [
					{
						"id": "as_789",
						"name": "PagerDuty Integration",
						"type": "pagerduty",
						"config_type": "integration",
						"created_at": "2024-01-03T00:00:00Z",
						"updated_at": "2024-01-03T00:00:00Z"
					}
				],
				"pagination_info": {
					"page_size": 5,
					"after": "cursor_456"
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
							switch tt.name {
							case "successful list alert sources":
								assertEqual(t, "10", req.URL.Query().Get("page_size"))
							case "with pagination":
								assertEqual(t, "5", req.URL.Query().Get("page_size"))
							}
						}
						if tt.params.After != "" {
							assertEqual(t, tt.params.After, req.URL.Query().Get("after"))
						}
					}

					return mockResponse(tt.mockStatusCode, tt.mockResponse), nil
				},
			}

			client := NewTestClient(mockClient)
			result, err := client.ListAlertSources(context.Background(), tt.params)

			if tt.wantError {
				assertError(t, err)
				return
			}

			assertNoError(t, err)
			if len(result.AlertSources) != tt.expectedCount {
				t.Errorf("expected %d alert sources, got %d", tt.expectedCount, len(result.AlertSources))
			}

			// Verify first alert source details
			if tt.expectedCount > 0 {
				source := result.AlertSources[0]
				switch tt.name {
				case "successful list alert sources":
					assertEqual(t, "as_123", source.ID)
					assertEqual(t, "Production Monitoring", source.Name)
					assertEqual(t, "http", source.Type)
					assertEqual(t, "webhook", source.ConfigType)
				case "with pagination":
					assertEqual(t, "as_789", source.ID)
					assertEqual(t, "PagerDuty Integration", source.Name)
					assertEqual(t, "pagerduty", source.Type)
				}
			}

			// Verify pagination info
			if tt.params != nil && tt.params.PageSize > 0 {
				if result.Pagination.PageSize != tt.params.PageSize {
					t.Errorf("expected page size %d, got %d", tt.params.PageSize, result.Pagination.PageSize)
				}
			}
		})
	}
}
