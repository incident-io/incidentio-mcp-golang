package handlers

import (
	"strings"
	"testing"
	"time"

	"github.com/incident-io/incidentio-mcp-golang/internal/client"
)

// --- Tool metadata tests ---

func TestListSchedulesTool_Metadata(t *testing.T) {
	tool := &ListSchedulesTool{}

	if tool.Name() != "list_schedules" {
		t.Errorf("Expected name 'list_schedules', got %s", tool.Name())
	}

	schema := tool.InputSchema()
	props := schema["properties"].(map[string]interface{})
	if _, ok := props["name"]; !ok {
		t.Error("Schema should have 'name' property")
	}
	if _, ok := props["page_size"]; !ok {
		t.Error("Schema should have 'page_size' property")
	}
	if _, ok := props["after"]; !ok {
		t.Error("Schema should have 'after' property")
	}
}

func TestGetScheduleTool_Metadata(t *testing.T) {
	tool := &GetScheduleTool{}

	if tool.Name() != "get_schedule" {
		t.Errorf("Expected name 'get_schedule', got %s", tool.Name())
	}

	schema := tool.InputSchema()
	required := schema["required"].([]interface{})
	if len(required) != 1 || required[0] != "id" {
		t.Errorf("Expected required=[id], got %v", required)
	}
}

func TestListScheduleEntriesTool_Metadata(t *testing.T) {
	tool := &ListScheduleEntriesTool{}

	if tool.Name() != "list_schedule_entries" {
		t.Errorf("Expected name 'list_schedule_entries', got %s", tool.Name())
	}

	schema := tool.InputSchema()
	props := schema["properties"].(map[string]interface{})
	if _, ok := props["timezone"]; !ok {
		t.Error("Schema should have 'timezone' property")
	}

	required := schema["required"].([]interface{})
	if len(required) != 3 {
		t.Errorf("Expected 3 required params, got %d", len(required))
	}
}

// --- Parameter validation tests ---

func TestGetScheduleTool_MissingID(t *testing.T) {
	tool := &GetScheduleTool{}

	_, err := tool.Execute(map[string]interface{}{})
	if err == nil {
		t.Fatal("Expected error for missing id")
	}
	if err.Error() != "id parameter is required" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestListScheduleEntriesTool_MissingParams(t *testing.T) {
	tool := &ListScheduleEntriesTool{}

	tests := []struct {
		name        string
		args        map[string]interface{}
		expectedErr string
	}{
		{
			name:        "missing schedule_id",
			args:        map[string]interface{}{"entry_window_start": "2026-01-01T00:00:00Z", "entry_window_end": "2026-02-01T00:00:00Z"},
			expectedErr: "schedule_id parameter is required",
		},
		{
			name:        "missing entry_window_start",
			args:        map[string]interface{}{"schedule_id": "abc", "entry_window_end": "2026-02-01T00:00:00Z"},
			expectedErr: "entry_window_start parameter is required",
		},
		{
			name:        "missing entry_window_end",
			args:        map[string]interface{}{"schedule_id": "abc", "entry_window_start": "2026-01-01T00:00:00Z"},
			expectedErr: "entry_window_end parameter is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tool.Execute(tt.args)
			if err == nil {
				t.Fatal("Expected error")
			}
			if err.Error() != tt.expectedErr {
				t.Errorf("Expected %q, got %q", tt.expectedErr, err.Error())
			}
		})
	}
}

func TestListScheduleEntriesTool_InvalidTimezone(t *testing.T) {
	tool := &ListScheduleEntriesTool{}

	args := map[string]interface{}{
		"schedule_id":        "abc",
		"entry_window_start": "2026-01-01T00:00:00Z",
		"entry_window_end":   "2026-02-01T00:00:00Z",
		"timezone":           "Not/A/Timezone",
	}

	_, err := tool.Execute(args)
	if err == nil {
		t.Fatal("Expected error for invalid timezone")
	}
	if got := err.Error(); !strings.Contains(got, "invalid timezone") {
		t.Errorf("Expected error containing 'invalid timezone', got %q", got)
	}
}

// --- Pure helper function tests ---

func TestFormatTime_UTC(t *testing.T) {
	ts := time.Date(2026, 2, 15, 10, 30, 0, 0, time.UTC)

	result := formatTime(ts, nil)
	expected := "2026-02-15T10:30:00Z"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestFormatTime_WithTimezone(t *testing.T) {
	ts := time.Date(2026, 2, 15, 10, 0, 0, 0, time.UTC)
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatalf("Failed to load timezone: %v", err)
	}

	result := formatTime(ts, loc)
	// February in Berlin is CET (UTC+1)
	expected := "2026-02-15T11:00:00+01:00"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestFormatTime_WithTimezone_DST(t *testing.T) {
	// July is CEST (UTC+2) in Berlin
	ts := time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC)
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatalf("Failed to load timezone: %v", err)
	}

	result := formatTime(ts, loc)
	expected := "2026-07-15T12:00:00+02:00"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestBuildEntryMap_AllFields(t *testing.T) {
	entry := client.ScheduleEntry{
		StartAt:     time.Date(2026, 2, 15, 10, 0, 0, 0, time.UTC),
		EndAt:       time.Date(2026, 2, 15, 18, 0, 0, 0, time.UTC),
		RotationID:  "rot-1",
		LayerID:     "layer-1",
		Fingerprint: "fp-123",
		User: client.User{
			ID:    "user-1",
			Name:  "Alice",
			Email: "alice@example.com",
		},
	}

	m := buildEntryMap(entry, nil)

	if m["start_at"] != "2026-02-15T10:00:00Z" {
		t.Errorf("Unexpected start_at: %v", m["start_at"])
	}
	if m["end_at"] != "2026-02-15T18:00:00Z" {
		t.Errorf("Unexpected end_at: %v", m["end_at"])
	}
	if m["rotation_id"] != "rot-1" {
		t.Errorf("Unexpected rotation_id: %v", m["rotation_id"])
	}
	if m["layer_id"] != "layer-1" {
		t.Errorf("Unexpected layer_id: %v", m["layer_id"])
	}
	if m["fingerprint"] != "fp-123" {
		t.Errorf("Unexpected fingerprint: %v", m["fingerprint"])
	}

	user := m["user"].(map[string]interface{})
	if user["id"] != "user-1" {
		t.Errorf("Unexpected user id: %v", user["id"])
	}
	if user["name"] != "Alice" {
		t.Errorf("Unexpected user name: %v", user["name"])
	}
	if user["email"] != "alice@example.com" {
		t.Errorf("Unexpected user email: %v", user["email"])
	}
	// Ensure we don't leak slack_user_id or role
	if _, ok := user["slack_user_id"]; ok {
		t.Error("User should not contain slack_user_id")
	}
	if _, ok := user["role"]; ok {
		t.Error("User should not contain role")
	}
}

func TestBuildEntryMap_OmitsEmptyOptionalFields(t *testing.T) {
	entry := client.ScheduleEntry{
		StartAt: time.Date(2026, 2, 15, 10, 0, 0, 0, time.UTC),
		EndAt:   time.Date(2026, 2, 15, 18, 0, 0, 0, time.UTC),
		User:    client.User{ID: "user-1", Name: "Alice", Email: "alice@example.com"},
		// RotationID, LayerID, Fingerprint all empty
	}

	m := buildEntryMap(entry, nil)

	if _, ok := m["rotation_id"]; ok {
		t.Error("Empty rotation_id should be omitted")
	}
	if _, ok := m["layer_id"]; ok {
		t.Error("Empty layer_id should be omitted")
	}
	if _, ok := m["fingerprint"]; ok {
		t.Error("Empty fingerprint should be omitted")
	}
}

func TestBuildEntryMap_WithTimezone(t *testing.T) {
	entry := client.ScheduleEntry{
		StartAt: time.Date(2026, 2, 15, 10, 0, 0, 0, time.UTC),
		EndAt:   time.Date(2026, 2, 15, 18, 0, 0, 0, time.UTC),
		User:    client.User{ID: "user-1", Name: "Alice"},
	}

	loc, _ := time.LoadLocation("Europe/Berlin")
	m := buildEntryMap(entry, loc)

	if m["start_at"] != "2026-02-15T11:00:00+01:00" {
		t.Errorf("Expected Berlin time, got %v", m["start_at"])
	}
	if m["end_at"] != "2026-02-15T19:00:00+01:00" {
		t.Errorf("Expected Berlin time, got %v", m["end_at"])
	}
}

func TestBuildEntryMaps(t *testing.T) {
	entries := []client.ScheduleEntry{
		{StartAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), EndAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), User: client.User{Name: "Alice"}},
		{StartAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), EndAt: time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC), User: client.User{Name: "Bob"}},
	}

	result := buildEntryMaps(entries, nil)
	if len(result) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(result))
	}

	user0 := result[0]["user"].(map[string]interface{})
	user1 := result[1]["user"].(map[string]interface{})
	if user0["name"] != "Alice" {
		t.Errorf("Expected Alice, got %v", user0["name"])
	}
	if user1["name"] != "Bob" {
		t.Errorf("Expected Bob, got %v", user1["name"])
	}
}

func TestBuildEntryMaps_WithTimezone(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Berlin")
	entries := []client.ScheduleEntry{
		{StartAt: time.Date(2026, 2, 15, 10, 0, 0, 0, time.UTC), EndAt: time.Date(2026, 2, 15, 18, 0, 0, 0, time.UTC), User: client.User{Name: "Alice"}},
		{StartAt: time.Date(2026, 2, 16, 10, 0, 0, 0, time.UTC), EndAt: time.Date(2026, 2, 16, 18, 0, 0, 0, time.UTC), User: client.User{Name: "Bob"}},
	}

	result := buildEntryMaps(entries, loc)
	// Both entries should have Berlin time (CET, UTC+1 in February)
	if result[0]["start_at"] != "2026-02-15T11:00:00+01:00" {
		t.Errorf("Expected Berlin time for entry 0, got %v", result[0]["start_at"])
	}
	if result[1]["start_at"] != "2026-02-16T11:00:00+01:00" {
		t.Errorf("Expected Berlin time for entry 1, got %v", result[1]["start_at"])
	}
}

func TestBuildEntryMaps_Empty(t *testing.T) {
	result := buildEntryMaps([]client.ScheduleEntry{}, nil)
	if len(result) != 0 {
		t.Errorf("Expected empty slice, got %d entries", len(result))
	}
}

func TestBuildScheduleSummaries_Basic(t *testing.T) {
	schedules := []client.Schedule{
		{
			ID:       "sched-1",
			Name:     "Engineering On-Call",
			Timezone: "Europe/Berlin",
		},
	}

	summaries := buildScheduleSummaries(schedules)
	if len(summaries) != 1 {
		t.Fatalf("Expected 1 summary, got %d", len(summaries))
	}

	s := summaries[0]
	if s["id"] != "sched-1" {
		t.Errorf("Unexpected id: %v", s["id"])
	}
	if s["name"] != "Engineering On-Call" {
		t.Errorf("Unexpected name: %v", s["name"])
	}
	if s["timezone"] != "Europe/Berlin" {
		t.Errorf("Unexpected timezone: %v", s["timezone"])
	}
	if _, ok := s["current_on_call"]; ok {
		t.Error("Should not have current_on_call when no shifts")
	}
}

func TestBuildScheduleSummaries_WithSingleShift(t *testing.T) {
	schedules := []client.Schedule{
		{
			ID:   "sched-1",
			Name: "Test",
			CurrentShifts: []client.ScheduleEntry{
				{User: client.User{Name: "Alice"}},
			},
		},
	}

	summaries := buildScheduleSummaries(schedules)
	names := summaries[0]["current_on_call"].([]string)
	if len(names) != 1 || names[0] != "Alice" {
		t.Errorf("Expected [Alice], got %v", names)
	}
}

func TestBuildScheduleSummaries_WithMultipleShifts(t *testing.T) {
	schedules := []client.Schedule{
		{
			ID:   "sched-1",
			Name: "Test",
			CurrentShifts: []client.ScheduleEntry{
				{User: client.User{Name: "Alice"}},
				{User: client.User{Name: "Bob"}},
			},
		},
	}

	summaries := buildScheduleSummaries(schedules)
	names := summaries[0]["current_on_call"].([]string)
	if len(names) != 2 {
		t.Fatalf("Expected 2 names, got %d", len(names))
	}
	if names[0] != "Alice" || names[1] != "Bob" {
		t.Errorf("Expected [Alice, Bob], got %v", names)
	}
}

func TestBuildScheduleSummaries_Empty(t *testing.T) {
	summaries := buildScheduleSummaries([]client.Schedule{})
	if summaries == nil {
		t.Error("Expected non-nil empty slice")
	}
	if len(summaries) != 0 {
		t.Errorf("Expected empty slice, got %d summaries", len(summaries))
	}
}

func TestBuildScheduleSummaries_DoesNotLeakExtraFields(t *testing.T) {
	schedules := []client.Schedule{
		{
			ID:       "sched-1",
			Name:     "Test",
			Timezone: "UTC",
			Config:   map[string]interface{}{"rotations": "should not appear"},
		},
	}

	summaries := buildScheduleSummaries(schedules)
	s := summaries[0]

	// Only id, name, timezone should be present (no current_on_call since no shifts)
	allowedKeys := map[string]bool{"id": true, "name": true, "timezone": true}
	for k := range s {
		if !allowedKeys[k] {
			t.Errorf("Unexpected key %q in summary", k)
		}
	}
}
