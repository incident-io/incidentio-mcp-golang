package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/incident-io/incidentio-mcp-golang/internal/client"
)

// ListSchedulesTool lists all on-call schedules
type ListSchedulesTool struct {
	apiClient *client.Client
}

func NewListSchedulesTool(c *client.Client) *ListSchedulesTool {
	return &ListSchedulesTool{apiClient: c}
}

func (t *ListSchedulesTool) Name() string {
	return "list_schedules"
}

func (t *ListSchedulesTool) Description() string {
	return `List on-call schedules in the organization.

Returns schedule summaries: id, name, timezone, and current on-call user.
Use the optional name filter to search for schedules by name (case-insensitive substring match).
Use this to find the schedule ID for a specific team before querying schedule entries.

Supports pagination via page_size and after parameters. When using the name filter, all pages are auto-fetched (up to 5000 schedules).`
}

func (t *ListSchedulesTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Filter schedules by name (case-insensitive substring match). When provided, all pages are fetched automatically (up to 5000 schedules).",
			},
			"page_size": map[string]interface{}{
				"type":        "integer",
				"description": "Number of results per page (default 25)",
				"default":     25,
				"minimum":     1,
				"maximum":     250,
			},
			"after": map[string]interface{}{
				"type":        "string",
				"description": "Pagination cursor from a previous response to fetch the next page",
			},
		},
		"additionalProperties": false,
	}
}

func (t *ListSchedulesTool) Execute(args map[string]interface{}) (string, error) {
	nameFilter := GetStringArg(args, "name")

	if nameFilter != "" {
		return t.executeWithNameFilter(nameFilter)
	}

	opts := &client.ListSchedulesOptions{
		PageSize: GetIntArg(args, "page_size", 25),
		After:    GetStringArg(args, "after"),
	}
	resp, err := t.apiClient.ListSchedules(opts)
	if err != nil {
		return "", fmt.Errorf("failed to list schedules: %w", err)
	}

	response := map[string]interface{}{
		"schedules":       buildScheduleSummaries(resp.Schedules),
		"count":           len(resp.Schedules),
		"pagination_meta": resp.PaginationMeta,
	}
	if resp.PaginationMeta.After != "" {
		response["fetch_next_page"] = map[string]interface{}{
			"action":  "Call list_schedules again with after parameter",
			"after":   resp.PaginationMeta.After,
			"message": "More schedules available. Fetch next page to get complete results.",
		}
	}
	return FormatJSONResponse(response)
}

const maxPaginationPages = 20 // safety cap: 20 pages × 250 = 5000 schedules max

func (t *ListSchedulesTool) executeWithNameFilter(nameFilter string) (string, error) {
	var allSchedules []client.Schedule
	after := ""
	for page := 0; page < maxPaginationPages; page++ {
		opts := &client.ListSchedulesOptions{
			PageSize: 250,
			After:    after,
		}
		resp, err := t.apiClient.ListSchedules(opts)
		if err != nil {
			return "", fmt.Errorf("failed to list schedules: %w", err)
		}
		allSchedules = append(allSchedules, resp.Schedules...)
		if resp.PaginationMeta.After == "" {
			break
		}
		after = resp.PaginationMeta.After
	}

	nameLower := strings.ToLower(nameFilter)
	var filtered []client.Schedule
	for _, s := range allSchedules {
		if strings.Contains(strings.ToLower(s.Name), nameLower) {
			filtered = append(filtered, s)
		}
	}

	response := map[string]interface{}{
		"schedules": buildScheduleSummaries(filtered),
		"count":     len(filtered),
	}
	if after != "" {
		response["truncated"] = true
		response["warning"] = "Results may be incomplete. Name filter scanned the maximum of 5000 schedules."
	}
	return FormatJSONResponse(response)
}

func buildScheduleSummaries(schedules []client.Schedule) []map[string]interface{} {
	summaries := make([]map[string]interface{}, len(schedules))
	for i, s := range schedules {
		summary := map[string]interface{}{
			"id":       s.ID,
			"name":     s.Name,
			"timezone": s.Timezone,
		}
		if len(s.CurrentShifts) > 0 {
			names := make([]string, len(s.CurrentShifts))
			for j, shift := range s.CurrentShifts {
				names[j] = shift.User.Name
			}
			summary["current_on_call"] = names
		}
		summaries[i] = summary
	}
	return summaries
}

// GetScheduleTool gets details of a specific schedule
type GetScheduleTool struct {
	apiClient *client.Client
}

func NewGetScheduleTool(c *client.Client) *GetScheduleTool {
	return &GetScheduleTool{apiClient: c}
}

func (t *GetScheduleTool) Name() string {
	return "get_schedule"
}

func (t *GetScheduleTool) Description() string {
	return `Get details of a specific on-call schedule by ID.

Returns the full schedule configuration including rotations, handoff times, and users in rotation.
Use list_schedules first to find the schedule ID.`
}

func (t *GetScheduleTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "The schedule ID (use list_schedules to find this)",
				"minLength":   1,
			},
		},
		"required":             []interface{}{"id"},
		"additionalProperties": false,
	}
}

func (t *GetScheduleTool) Execute(args map[string]interface{}) (string, error) {
	id := GetStringArg(args, "id")
	if id == "" {
		return "", fmt.Errorf("id parameter is required")
	}

	schedule, err := t.apiClient.GetSchedule(id)
	if err != nil {
		return "", fmt.Errorf("failed to get schedule: %w", err)
	}

	return FormatJSONResponse(schedule)
}

// ListScheduleEntriesTool lists on-call entries for a schedule within a time window
type ListScheduleEntriesTool struct {
	apiClient *client.Client
}

func NewListScheduleEntriesTool(c *client.Client) *ListScheduleEntriesTool {
	return &ListScheduleEntriesTool{apiClient: c}
}

func (t *ListScheduleEntriesTool) Name() string {
	return "list_schedule_entries"
}

func (t *ListScheduleEntriesTool) Description() string {
	return `Get on-call entries for a schedule within a date range. This is the primary tool for answering "who was on-call when?"

Returns three sets of entries:
- final: The computed timeline of who was actually on-call (combining rotations and overrides). Each entry includes an is_override flag.
- overrides: Override entries only — when someone temporarily took over another person's shift.
- scheduled: Base rotation entries only (what would have happened without overrides).

Each entry includes exact start/end timestamps (ISO 8601), the on-call user, and (for final entries) is_override: true/false.

Use cases:
- Payroll review: query a month's entries to see who was on-call each day, including partial-day overrides
- Override audit: check the overrides array or filter final entries by is_override
- On-call history: get the complete on-call timeline for any date range

Parameters:
- schedule_id: Get this from list_schedules
- entry_window_start: ISO 8601 datetime (e.g., "2026-02-01T00:00:00Z")
- entry_window_end: ISO 8601 datetime (e.g., "2026-03-01T00:00:00Z")
- timezone: Optional IANA timezone (e.g., "Europe/Berlin") to convert timestamps from UTC`
}

func (t *ListScheduleEntriesTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"schedule_id": map[string]interface{}{
				"type":        "string",
				"description": "The schedule ID to get entries for (use list_schedules to find this)",
				"minLength":   1,
			},
			"entry_window_start": map[string]interface{}{
				"type":        "string",
				"description": "Start of the time window in ISO 8601 format (e.g., 2026-02-01T00:00:00Z)",
				"minLength":   1,
			},
			"entry_window_end": map[string]interface{}{
				"type":        "string",
				"description": "End of the time window in ISO 8601 format (e.g., 2026-03-01T00:00:00Z)",
				"minLength":   1,
			},
			"timezone": map[string]interface{}{
				"type":        "string",
				"description": "IANA timezone to convert timestamps to (e.g., \"Europe/Berlin\"). If omitted, timestamps remain in UTC.",
			},
		},
		"required":             []interface{}{"schedule_id", "entry_window_start", "entry_window_end"},
		"additionalProperties": false,
	}
}

func (t *ListScheduleEntriesTool) Execute(args map[string]interface{}) (string, error) {
	scheduleID := GetStringArg(args, "schedule_id")
	if scheduleID == "" {
		return "", fmt.Errorf("schedule_id parameter is required")
	}

	windowStart := GetStringArg(args, "entry_window_start")
	if windowStart == "" {
		return "", fmt.Errorf("entry_window_start parameter is required")
	}

	windowEnd := GetStringArg(args, "entry_window_end")
	if windowEnd == "" {
		return "", fmt.Errorf("entry_window_end parameter is required")
	}

	// Load timezone if provided
	tzName := GetStringArg(args, "timezone")
	var loc *time.Location
	if tzName != "" {
		var err error
		loc, err = time.LoadLocation(tzName)
		if err != nil {
			return "", fmt.Errorf("invalid timezone %q: %w", tzName, err)
		}
	}

	opts := &client.ListScheduleEntriesOptions{
		ScheduleID:       scheduleID,
		EntryWindowStart: windowStart,
		EntryWindowEnd:   windowEnd,
	}

	resp, err := t.apiClient.ListScheduleEntries(opts)
	if err != nil {
		return "", fmt.Errorf("failed to list schedule entries: %w", err)
	}

	// Build override fingerprint set for is_override flag
	overrideFingerprints := make(map[string]bool, len(resp.Overrides))
	for _, o := range resp.Overrides {
		if o.Fingerprint != "" {
			overrideFingerprints[o.Fingerprint] = true
		}
	}

	// Build final entries with is_override flag
	finalEntries := make([]map[string]interface{}, len(resp.Final))
	for i, entry := range resp.Final {
		finalEntries[i] = buildEntryMap(entry, loc)
		finalEntries[i]["is_override"] = overrideFingerprints[entry.Fingerprint]
	}

	response := map[string]interface{}{
		"schedule_id":        scheduleID,
		"entry_window_start": windowStart,
		"entry_window_end":   windowEnd,
		"final": map[string]interface{}{
			"entries": finalEntries,
			"count":   len(finalEntries),
			"note":    "The computed timeline of who was actually on-call. Each entry includes is_override flag.",
		},
		"overrides": map[string]interface{}{
			"entries": buildEntryMaps(resp.Overrides, loc),
			"count":   len(resp.Overrides),
			"note":    "Override entries only.",
		},
		"scheduled": map[string]interface{}{
			"entries": buildEntryMaps(resp.Scheduled, loc),
			"count":   len(resp.Scheduled),
			"note":    "Base rotation entries (what would have happened without overrides).",
		},
	}

	if tzName != "" {
		response["timezone"] = tzName
	}

	return FormatJSONResponse(response)
}

func formatTime(t time.Time, loc *time.Location) string {
	if loc != nil {
		return t.In(loc).Format(time.RFC3339)
	}
	return t.Format(time.RFC3339)
}

func buildEntryMap(entry client.ScheduleEntry, loc *time.Location) map[string]interface{} {
	m := make(map[string]interface{}, 6)
	m["start_at"] = formatTime(entry.StartAt, loc)
	m["end_at"] = formatTime(entry.EndAt, loc)
	m["user"] = map[string]interface{}{
		"id":    entry.User.ID,
		"name":  entry.User.Name,
		"email": entry.User.Email,
	}
	if entry.RotationID != "" {
		m["rotation_id"] = entry.RotationID
	}
	if entry.LayerID != "" {
		m["layer_id"] = entry.LayerID
	}
	if entry.Fingerprint != "" {
		m["fingerprint"] = entry.Fingerprint
	}
	return m
}

func buildEntryMaps(entries []client.ScheduleEntry, loc *time.Location) []map[string]interface{} {
	result := make([]map[string]interface{}, len(entries))
	for i, entry := range entries {
		result[i] = buildEntryMap(entry, loc)
	}
	return result
}
