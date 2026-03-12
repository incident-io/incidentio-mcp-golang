package client

import (
	"context"
	"net/http"
	"testing"
)

func TestCreateAlertEvent(t *testing.T) {
	tests := []struct {
		name           string
		request        *CreateAlertEventRequest
		mockResponse   string
		mockStatusCode int
		wantError      bool
	}{
		{
			name: "successful create alert event",
			request: &CreateAlertEventRequest{
				AlertSourceID: "as_123",
				Title:         "Database connection failure",
				Description:   "Unable to connect to primary database",
				Status:        "firing",
			},
			mockResponse: `{
				"alert_event": {
					"id": "ae_123",
					"alert_source_id": "as_123",
					"title": "Database connection failure",
					"description": "Unable to connect to primary database",
					"status": "firing",
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-01T00:00:00Z"
				}
			}`,
			mockStatusCode: http.StatusCreated,
			wantError:      false,
		},
		{
			name: "create with deduplication key",
			request: &CreateAlertEventRequest{
				AlertSourceID:    "as_456",
				Title:            "High CPU usage",
				Description:      "CPU usage above 90% for 5 minutes",
				Status:           "firing",
				DeduplicationKey: "cpu-alert-prod-server-1",
			},
			mockResponse: `{
				"alert_event": {
					"id": "ae_456",
					"alert_source_id": "as_456",
					"title": "High CPU usage",
					"description": "CPU usage above 90% for 5 minutes",
					"status": "firing",
					"deduplication_key": "cpu-alert-prod-server-1",
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-01T00:00:00Z"
				}
			}`,
			mockStatusCode: http.StatusCreated,
			wantError:      false,
		},
		{
			name: "create with metadata",
			request: &CreateAlertEventRequest{
				AlertSourceID: "as_789",
				Title:         "Payment processing error",
				Description:   "Failed to process payment transaction",
				Status:        "firing",
				Metadata: map[string]interface{}{
					"transaction_id": "txn_12345",
					"amount":         99.99,
					"currency":       "USD",
					"error_code":     "INSUFFICIENT_FUNDS",
				},
			},
			mockResponse: `{
				"alert_event": {
					"id": "ae_789",
					"alert_source_id": "as_789",
					"title": "Payment processing error",
					"description": "Failed to process payment transaction",
					"status": "firing",
					"metadata": {
						"transaction_id": "txn_12345",
						"amount": 99.99,
						"currency": "USD",
						"error_code": "INSUFFICIENT_FUNDS"
					},
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-01T00:00:00Z"
				}
			}`,
			mockStatusCode: http.StatusCreated,
			wantError:      false,
		},
		{
			name: "resolve alert event",
			request: &CreateAlertEventRequest{
				AlertSourceID:    "as_123",
				Title:            "Database connection restored",
				Description:      "Connection to primary database has been restored",
				Status:           "resolved",
				DeduplicationKey: "db-alert-prod",
			},
			mockResponse: `{
				"alert_event": {
					"id": "ae_999",
					"alert_source_id": "as_123",
					"title": "Database connection restored",
					"description": "Connection to primary database has been restored",
					"status": "resolved",
					"deduplication_key": "db-alert-prod",
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-01T00:00:00Z"
				}
			}`,
			mockStatusCode: http.StatusCreated,
			wantError:      false,
		},
		{
			name: "invalid alert source",
			request: &CreateAlertEventRequest{
				AlertSourceID: "as_invalid",
				Title:         "Test alert",
				Status:        "firing",
			},
			mockResponse:   `{"error": "Alert source not found"}`,
			mockStatusCode: http.StatusNotFound,
			wantError:      true,
		},
		{
			name: "missing required fields",
			request: &CreateAlertEventRequest{
				AlertSourceID: "as_123",
				// Missing title
			},
			mockResponse:   `{"error": "Title is required"}`,
			mockStatusCode: http.StatusBadRequest,
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					assertEqual(t, "POST", req.Method)
					assertEqual(t, "/alert_events/http", req.URL.Path)
					assertEqual(t, "Bearer test-api-key", req.Header.Get("Authorization"))
					assertEqual(t, "application/json", req.Header.Get("Content-Type"))

					return mockResponse(tt.mockStatusCode, tt.mockResponse), nil
				},
			}

			client := NewTestClient(mockClient)
			event, err := client.CreateAlertEvent(context.Background(), tt.request)

			if tt.wantError {
				assertError(t, err)
				return
			}

			assertNoError(t, err)
			assertEqual(t, tt.request.AlertSourceID, event.AlertSourceID)
			assertEqual(t, tt.request.Title, event.Title)

			// Verify optional fields
			if tt.request.Description != "" {
				assertEqual(t, tt.request.Description, event.Description)
			}

			if tt.request.DeduplicationKey != "" {
				assertEqual(t, tt.request.DeduplicationKey, event.DeduplicationKey)
			}

			if tt.request.Status != "" {
				assertEqual(t, tt.request.Status, event.Status)
			}

			// Verify metadata
			if tt.request.Metadata != nil {
				if event.Metadata == nil {
					t.Error("expected metadata to be set")
				} else {
					// Check specific metadata fields based on test case
					if tt.name == "create with metadata" {
						if event.Metadata["transaction_id"] != "txn_12345" {
							t.Errorf("expected transaction_id to be txn_12345, got %v", event.Metadata["transaction_id"])
						}
						if event.Metadata["amount"] != 99.99 {
							t.Errorf("expected amount to be 99.99, got %v", event.Metadata["amount"])
						}
					}
				}
			}
		})
	}
}
