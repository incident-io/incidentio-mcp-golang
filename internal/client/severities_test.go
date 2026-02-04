package client

import (
	"net/http"
	"testing"
	"time"
)

func TestListSeverities_Caching(t *testing.T) {
	callCount := 0
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			assertEqual(t, "GET", req.Method)
			return mockResponse(http.StatusOK, `{
				"severities": [
					{
						"id": "sev_1",
						"name": "Critical",
						"rank": 1,
						"created_at": "2024-01-01T00:00:00Z",
						"updated_at": "2024-01-01T00:00:00Z"
					}
				]
			}`), nil
		},
	}

	client := NewTestClient(mockClient)

	// First call should hit the API
	resp1, err := client.ListSeverities()
	assertNoError(t, err)
	if len(resp1.Severities) != 1 {
		t.Errorf("Expected 1 severity, got %d", len(resp1.Severities))
	}
	if callCount != 1 {
		t.Errorf("Expected 1 API call, got %d", callCount)
	}

	// Second call should use cache
	resp2, err := client.ListSeverities()
	assertNoError(t, err)
	if len(resp2.Severities) != 1 {
		t.Errorf("Expected 1 severity, got %d", len(resp2.Severities))
	}
	if callCount != 1 {
		t.Errorf("Expected still 1 API call (cached), got %d", callCount)
	}

	// Verify same data
	if resp1.Severities[0].ID != resp2.Severities[0].ID {
		t.Error("Cached response should match original")
	}
}

func TestGetSeverity_Caching(t *testing.T) {
	callCount := 0
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			assertEqual(t, "GET", req.Method)
			return mockResponse(http.StatusOK, `{
				"severity": {
					"id": "sev_1",
					"name": "Critical",
					"rank": 1,
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-01T00:00:00Z"
				}
			}`), nil
		},
	}

	client := NewTestClient(mockClient)

	// First call should hit the API
	sev1, err := client.GetSeverity("sev_1")
	assertNoError(t, err)
	assertEqual(t, "sev_1", sev1.ID)
	if callCount != 1 {
		t.Errorf("Expected 1 API call, got %d", callCount)
	}

	// Second call should use cache
	sev2, err := client.GetSeverity("sev_1")
	assertNoError(t, err)
	assertEqual(t, "sev_1", sev2.ID)
	if callCount != 1 {
		t.Errorf("Expected still 1 API call (cached), got %d", callCount)
	}
}

func TestSeverities_CacheExpiration(t *testing.T) {
	callCount := 0
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			return mockResponse(http.StatusOK, `{
				"severities": [
					{
						"id": "sev_1",
						"name": "Critical",
						"rank": 1,
						"created_at": "2024-01-01T00:00:00Z",
						"updated_at": "2024-01-01T00:00:00Z"
					}
				]
			}`), nil
		},
	}

	client := NewTestClient(mockClient)
	// Override cache with short TTL for testing
	client.cache = NewCache(100 * time.Millisecond)

	// First call
	_, err := client.ListSeverities()
	assertNoError(t, err)
	if callCount != 1 {
		t.Errorf("Expected 1 API call, got %d", callCount)
	}

	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)

	// Second call should hit API again
	_, err = client.ListSeverities()
	assertNoError(t, err)
	if callCount != 2 {
		t.Errorf("Expected 2 API calls after expiration, got %d", callCount)
	}
}

// Made with Bob
