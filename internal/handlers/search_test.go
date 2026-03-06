package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/incident-io/incidentio-mcp-golang/internal/client"
)

func newTestClient(t *testing.T, srv *httptest.Server) *client.Client {
	t.Helper()
	t.Setenv("INCIDENT_IO_API_KEY", "test-key")
	c, err := client.NewClient()
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	c.SetBaseURL(srv.URL)
	return c
}

func unmarshalResult(t *testing.T, result string) map[string]interface{} {
	t.Helper()
	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	return resp
}

func fakeAlert(title string) client.Alert {
	return client.Alert{
		ID:        "alert-" + title,
		Title:     title,
		Status:    "firing",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func fakeIncident(name string) client.Incident {
	return client.Incident{
		ID:             "inc-" + name,
		Reference:      "INC-1",
		Name:           name,
		Permalink:      "https://app.incident.io/incidents/inc-" + name,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		IncidentStatus: client.IncidentStatus{Name: "active"},
		Severity:       client.Severity{Name: "minor"},
	}
}

type alertPage struct {
	alerts []client.Alert
	after  string
}

type incidentPage struct {
	incidents  []client.Incident
	after      string
	totalCount int
}

func pageIndex(after string, cursors []string) int {
	if after != "" {
		for i := range cursors {
			if i > 0 && cursors[i-1] == after {
				return i
			}
		}
	}
	return 0
}

func alertsHandler(pages []alertPage) http.HandlerFunc {
	cursors := make([]string, len(pages))
	for i, p := range pages {
		cursors[i] = p.after
	}
	return func(w http.ResponseWriter, r *http.Request) {
		idx := pageIndex(r.URL.Query().Get("after"), cursors)
		page := pages[idx]
		resp := map[string]interface{}{
			"alerts": page.alerts,
			"pagination_meta": map[string]interface{}{
				"after":              page.after,
				"page_size":          len(page.alerts),
				"total_record_count": 0,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func incidentsHandler(pages []incidentPage) http.HandlerFunc {
	cursors := make([]string, len(pages))
	for i, p := range pages {
		cursors[i] = p.after
	}
	return func(w http.ResponseWriter, r *http.Request) {
		idx := pageIndex(r.URL.Query().Get("after"), cursors)
		page := pages[idx]
		resp := map[string]interface{}{
			"incidents": page.incidents,
			"pagination_meta": map[string]interface{}{
				"after":              page.after,
				"page_size":          len(page.incidents),
				"total_record_count": page.totalCount,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func TestListAlerts_TitleSearch_AcrossPages(t *testing.T) {
	pages := []alertPage{
		{alerts: []client.Alert{fakeAlert("cpu spike"), fakeAlert("dw_devices_aggregation timeout")}, after: "cursor1"},
		{alerts: []client.Alert{fakeAlert("memory OOM"), fakeAlert("disk full")}, after: "cursor2"},
		{alerts: []client.Alert{fakeAlert("dw_devices_aggregation failure")}, after: ""},
	}
	srv := httptest.NewServer(alertsHandler(pages))
	defer srv.Close()

	tool := NewListAlertsTool(newTestClient(t, srv))
	result, err := tool.Execute(map[string]interface{}{
		"title": "dw_devices_aggregation",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	resp := unmarshalResult(t, result)

	count := int(resp["count"].(float64))
	if count != 2 {
		t.Errorf("expected 2 matches, got %d", count)
	}
	pagesScanned := int(resp["pages_scanned"].(float64))
	if pagesScanned != 3 {
		t.Errorf("expected 3 pages scanned, got %d", pagesScanned)
	}
}

func TestListAlerts_TitleSearch_CaseInsensitive(t *testing.T) {
	pages := []alertPage{
		{alerts: []client.Alert{fakeAlert("CPU Spike Alert"), fakeAlert("cpu spike warning")}, after: ""},
	}
	srv := httptest.NewServer(alertsHandler(pages))
	defer srv.Close()

	tool := NewListAlertsTool(newTestClient(t, srv))
	result, err := tool.Execute(map[string]interface{}{
		"title": "CPU SPIKE",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	resp := unmarshalResult(t, result)

	if int(resp["count"].(float64)) != 2 {
		t.Errorf("expected 2 case-insensitive matches, got %v", resp["count"])
	}
}

func TestListAlerts_TitleSearch_StopsAtMaxResults(t *testing.T) {
	var pages []alertPage
	for i := 0; i < maxSearchPages; i++ {
		var alerts []client.Alert
		for j := 0; j < 10; j++ {
			alerts = append(alerts, fakeAlert(fmt.Sprintf("match-%d-%d", i, j)))
		}
		cursor := ""
		if i < maxSearchPages-1 {
			cursor = fmt.Sprintf("cursor%d", i)
		}
		pages = append(pages, alertPage{alerts: alerts, after: cursor})
	}
	srv := httptest.NewServer(alertsHandler(pages))
	defer srv.Close()

	tool := NewListAlertsTool(newTestClient(t, srv))
	result, err := tool.Execute(map[string]interface{}{
		"title": "match",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	resp := unmarshalResult(t, result)

	count := int(resp["count"].(float64))
	if count != maxSearchResults {
		t.Errorf("expected %d (maxSearchResults), got %d", maxSearchResults, count)
	}
	if resp["truncated"] != true {
		t.Error("expected truncated=true")
	}
}

func TestListAlerts_TitleSearch_ScanTruncated(t *testing.T) {
	var pages []alertPage
	for i := 0; i < maxSearchPages; i++ {
		alerts := []client.Alert{fakeAlert(fmt.Sprintf("noise-%d", i))}
		cursor := ""
		if i < maxSearchPages-1 {
			cursor = fmt.Sprintf("cursor%d", i)
		}
		pages = append(pages, alertPage{alerts: alerts, after: cursor})
	}
	srv := httptest.NewServer(alertsHandler(pages))
	defer srv.Close()

	tool := NewListAlertsTool(newTestClient(t, srv))
	result, err := tool.Execute(map[string]interface{}{
		"title": "nonexistent",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	resp := unmarshalResult(t, result)

	if int(resp["count"].(float64)) != 0 {
		t.Errorf("expected 0 matches, got %v", resp["count"])
	}
	if resp["scan_truncated"] != true {
		t.Error("expected scan_truncated=true when page limit reached")
	}
	if resp["truncated"] != nil {
		t.Error("truncated should not be set when result limit not reached")
	}
}

func TestListAlerts_NoTitle_SinglePage(t *testing.T) {
	pages := []alertPage{
		{alerts: []client.Alert{fakeAlert("a1"), fakeAlert("a2")}, after: "next"},
	}
	srv := httptest.NewServer(alertsHandler(pages))
	defer srv.Close()

	tool := NewListAlertsTool(newTestClient(t, srv))
	result, err := tool.Execute(map[string]interface{}{
		"page_size": float64(25),
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	resp := unmarshalResult(t, result)

	if _, ok := resp["🚨 PAGINATION WARNING"]; !ok {
		t.Error("expected pagination warning in standard mode with more pages")
	}
	fetchNext, ok := resp["📊 FETCH_NEXT_PAGE"].(map[string]interface{})
	if !ok {
		t.Error("expected FETCH_NEXT_PAGE in standard mode")
	} else if fetchNext["after"] != "next" {
		t.Errorf("expected FETCH_NEXT_PAGE.after='next', got %v", fetchNext["after"])
	}
	if _, ok := resp["title_filter"]; ok {
		t.Error("title_filter should not be present in standard mode")
	}
}

func TestListIncidents_Search_AcrossPages(t *testing.T) {
	pages := []incidentPage{
		{incidents: []client.Incident{fakeIncident("other-1"), fakeIncident("promtail failure")}, after: "c1", totalCount: 4},
		{incidents: []client.Incident{fakeIncident("promtail restart"), fakeIncident("other-2")}, after: "", totalCount: 4},
	}
	srv := httptest.NewServer(incidentsHandler(pages))
	defer srv.Close()

	tool := NewListIncidentsTool(newTestClient(t, srv))
	result, err := tool.Execute(map[string]interface{}{
		"search": "promtail",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	resp := unmarshalResult(t, result)

	count := int(resp["count"].(float64))
	if count != 2 {
		t.Errorf("expected 2 matches, got %d", count)
	}
	if int(resp["pages_scanned"].(float64)) != 2 {
		t.Errorf("expected 2 pages scanned, got %v", resp["pages_scanned"])
	}
}

func TestListIncidents_Search_StopsAtMaxResults(t *testing.T) {
	var pages []incidentPage
	for i := 0; i < maxSearchPages; i++ {
		var incidents []client.Incident
		for j := 0; j < 10; j++ {
			incidents = append(incidents, fakeIncident(fmt.Sprintf("match-%d-%d", i, j)))
		}
		cursor := ""
		if i < maxSearchPages-1 {
			cursor = fmt.Sprintf("c%d", i)
		}
		pages = append(pages, incidentPage{incidents: incidents, after: cursor, totalCount: maxSearchPages * 10})
	}
	srv := httptest.NewServer(incidentsHandler(pages))
	defer srv.Close()

	tool := NewListIncidentsTool(newTestClient(t, srv))
	result, err := tool.Execute(map[string]interface{}{
		"search": "match",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	resp := unmarshalResult(t, result)

	count := int(resp["count"].(float64))
	if count != maxSearchResults {
		t.Errorf("expected %d, got %d", maxSearchResults, count)
	}
	if resp["truncated"] != true {
		t.Error("expected truncated=true")
	}
}

func TestListIncidents_Search_SummaryMode(t *testing.T) {
	pages := []incidentPage{
		{incidents: []client.Incident{fakeIncident("target incident")}, after: "", totalCount: 1},
	}
	srv := httptest.NewServer(incidentsHandler(pages))
	defer srv.Close()

	tool := NewListIncidentsTool(newTestClient(t, srv))
	result, err := tool.Execute(map[string]interface{}{
		"search": "target",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	resp := unmarshalResult(t, result)

	incidents := resp["incidents"].([]interface{})
	if len(incidents) != 1 {
		t.Fatalf("expected 1 incident, got %d", len(incidents))
	}
	inc := incidents[0].(map[string]interface{})
	if _, ok := inc["reference"]; !ok {
		t.Error("summary mode should include 'reference'")
	}
	if _, ok := inc["id"]; ok {
		t.Error("summary mode should NOT include 'id'")
	}
}

func TestListIncidents_Search_FullDetailMode(t *testing.T) {
	pages := []incidentPage{
		{incidents: []client.Incident{fakeIncident("target incident")}, after: "", totalCount: 1},
	}
	srv := httptest.NewServer(incidentsHandler(pages))
	defer srv.Close()

	tool := NewListIncidentsTool(newTestClient(t, srv))
	result, err := tool.Execute(map[string]interface{}{
		"search":  "target",
		"summary": false,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	resp := unmarshalResult(t, result)

	incidents := resp["incidents"].([]interface{})
	if len(incidents) != 1 {
		t.Fatalf("expected 1 incident, got %d", len(incidents))
	}
	inc := incidents[0].(map[string]interface{})
	if _, ok := inc["id"]; !ok {
		t.Error("full detail mode should include 'id'")
	}
	if _, ok := inc["reference"]; !ok {
		t.Error("full detail mode should include 'reference'")
	}
}

func TestListIncidents_NoSearch_SinglePage(t *testing.T) {
	pages := []incidentPage{
		{incidents: []client.Incident{fakeIncident("inc1")}, after: "next", totalCount: 100},
	}
	srv := httptest.NewServer(incidentsHandler(pages))
	defer srv.Close()

	tool := NewListIncidentsTool(newTestClient(t, srv))
	result, err := tool.Execute(map[string]interface{}{
		"page_size": float64(25),
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	resp := unmarshalResult(t, result)

	if resp["has_more_results"] != true {
		t.Error("expected has_more_results=true in standard mode")
	}
	fetchNext, ok := resp["FETCH_NEXT_PAGE"].(map[string]interface{})
	if !ok {
		t.Error("expected FETCH_NEXT_PAGE in standard mode")
	} else if fetchNext["after"] != "next" {
		t.Errorf("expected FETCH_NEXT_PAGE.after='next', got %v", fetchNext["after"])
	}
	if _, ok := resp["search_filter"]; ok {
		t.Error("search_filter should not be present in standard mode")
	}
}
