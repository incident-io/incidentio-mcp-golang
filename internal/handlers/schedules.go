package handlers

import (
	"fmt"

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
	return `List all on-call schedules in the organization.

Returns schedule IDs, names, timezones, and current on-call shifts.
Use this to find the schedule ID for a specific team before querying schedule entries.

Supports pagination via page_size and after parameters.`
}

func (t *ListSchedulesTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
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
	opts := &client.ListSchedulesOptions{
		PageSize: GetIntArg(args, "page_size", 25),
		After:    GetStringArg(args, "after"),
	}

	resp, err := t.apiClient.ListSchedules(opts)
	if err != nil {
		return "", fmt.Errorf("failed to list schedules: %w", err)
	}

	response := map[string]interface{}{
		"schedules":       resp.Schedules,
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
		return "", fmt.Errorf("schedule ID is required")
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
- final: The computed timeline of who was actually on-call (combining rotations and overrides). This is what you should use.
- overrides: Override entries only — when someone temporarily took over another person's shift.
- scheduled: Base rotation entries only (what would have happened without overrides).

Each entry includes exact start/end timestamps (ISO 8601) and the on-call user.
To identify overrides in the final timeline, compare final entries against the overrides list.

Use cases:
- Payroll review: query a month's entries to see who was on-call each day, including partial-day overrides
- Override audit: check the overrides array to see all manual shift changes
- On-call history: get the complete on-call timeline for any date range

Parameters:
- schedule_id: Get this from list_schedules
- entry_window_start: ISO 8601 datetime (e.g., "2026-02-01T00:00:00Z")
- entry_window_end: ISO 8601 datetime (e.g., "2026-03-01T00:00:00Z")`
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

	opts := &client.ListScheduleEntriesOptions{
		ScheduleID:       scheduleID,
		EntryWindowStart: windowStart,
		EntryWindowEnd:   windowEnd,
	}

	resp, err := t.apiClient.ListScheduleEntries(opts)
	if err != nil {
		return "", fmt.Errorf("failed to list schedule entries: %w", err)
	}

	response := map[string]interface{}{
		"schedule_id":        scheduleID,
		"entry_window_start": windowStart,
		"entry_window_end":   windowEnd,
		"final": map[string]interface{}{
			"entries": resp.Final,
			"count":   len(resp.Final),
			"note":    "The computed timeline of who was actually on-call. Use this for payroll and audit.",
		},
		"overrides": map[string]interface{}{
			"entries": resp.Overrides,
			"count":   len(resp.Overrides),
			"note":    "Override entries only. Compare against final entries to identify which shifts were overrides.",
		},
		"scheduled": map[string]interface{}{
			"entries": resp.Scheduled,
			"count":   len(resp.Scheduled),
			"note":    "Base rotation entries (what would have happened without overrides).",
		},
	}

	return FormatJSONResponse(response)
}
