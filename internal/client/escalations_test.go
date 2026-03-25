package client

import (
	"io"
	"net/http"
	"testing"
)

func TestListEscalationPaths(t *testing.T) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			assertEqual(t, "GET", req.Method)
			assertEqual(t, "25", req.URL.Query().Get("page_size"))
			assertEqual(t, "after_cursor", req.URL.Query().Get("after"))
			body := `{"escalation_paths":[{"id":"p1","name":"Path"}],"pagination_meta":{"after":"next","page_size":25}}`
			return mockResponse(http.StatusOK, body), nil
		},
	}
	c := NewTestClient(mockClient)
	res, err := c.ListEscalationPaths(&ListEscalationPathsParams{PageSize: 25, After: "after_cursor"})
	assertNoError(t, err)
	if len(res.EscalationPaths) != 1 {
		t.Fatalf("paths: got %d", len(res.EscalationPaths))
	}
	if res.PaginationMeta.After != "next" {
		t.Errorf("after: %q", res.PaginationMeta.After)
	}
}

func TestListEscalationPaths_CapsPageSize(t *testing.T) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			if req.URL.Query().Get("page_size") != "25" {
				t.Errorf("expected cap 25, got %q", req.URL.Query().Get("page_size"))
			}
			return mockResponse(http.StatusOK, `{"escalation_paths":[],"pagination_meta":{"page_size":25}}`), nil
		},
	}
	c := NewTestClient(mockClient)
	_, err := c.ListEscalationPaths(&ListEscalationPathsParams{PageSize: 100})
	assertNoError(t, err)
}

func TestDestroyEscalationPath(t *testing.T) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			assertEqual(t, "DELETE", req.Method)
			if req.URL.Path != "/escalation_paths/ep_1" {
				t.Errorf("path %s", req.URL.Path)
			}
			if req.Body != nil {
				b, _ := io.ReadAll(req.Body)
				if len(b) != 0 {
					t.Errorf("unexpected body")
				}
			}
			return mockResponse(http.StatusNoContent, ""), nil
		},
	}
	c := NewTestClient(mockClient)
	err := c.DestroyEscalationPath("ep_1")
	assertNoError(t, err)
}

func TestListEscalations_Query(t *testing.T) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			q := req.URL.Query()
			assertEqual(t, "10", q.Get("page_size"))
			if q.Get("status[one_of]") != "triggered" {
				t.Errorf("status filter: %v", q["status[one_of]"])
			}
			body := `{"escalations":[],"pagination_meta":{"page_size":10}}`
			return mockResponse(http.StatusOK, body), nil
		},
	}
	c := NewTestClient(mockClient)
	_, err := c.ListEscalations(&ListEscalationsParams{
		PageSize:    10,
		StatusOneOf: []string{"triggered"},
	})
	assertNoError(t, err)
}

func TestUpdateEscalationPath_PUT(t *testing.T) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			assertEqual(t, "PUT", req.Method)
			body := `{"escalation_path":{"id":"ep_1","name":"N"}}`
			return mockResponse(http.StatusOK, body), nil
		},
	}
	c := NewTestClient(mockClient)
	raw, err := c.UpdateEscalationPath("ep_1", map[string]interface{}{"name": "N"})
	assertNoError(t, err)
	if string(raw) != `{"id":"ep_1","name":"N"}` {
		t.Errorf("got %s", string(raw))
	}
}

func TestCreateEscalation(t *testing.T) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			assertEqual(t, "POST", req.Method)
			if req.URL.Path != "/escalations" {
				t.Errorf("path %s", req.URL.Path)
			}
			body := `{"escalation":{"id":"e1","title":"T"}}`
			return mockResponse(http.StatusCreated, body), nil
		},
	}
	c := NewTestClient(mockClient)
	raw, err := c.CreateEscalation(&CreateEscalationRequest{
		Title:            "T",
		EscalationPathID: "p1",
	})
	assertNoError(t, err)
	if string(raw) != `{"id":"e1","title":"T"}` {
		t.Errorf("got %s", string(raw))
	}
}
