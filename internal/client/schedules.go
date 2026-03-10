package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// Schedule represents an on-call schedule in incident.io
type Schedule struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Timezone      string                 `json:"timezone"`
	CurrentShifts []ScheduleEntry        `json:"current_shifts,omitempty"`
	Config        map[string]interface{} `json:"config,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// ScheduleEntry represents a single on-call entry with start/end times and user
type ScheduleEntry struct {
	EntryID     string    `json:"entry_id,omitempty"`
	StartAt     time.Time `json:"start_at"`
	EndAt       time.Time `json:"end_at"`
	RotationID  string    `json:"rotation_id,omitempty"`
	LayerID     string    `json:"layer_id,omitempty"`
	Fingerprint string    `json:"fingerprint,omitempty"`
	User        User      `json:"user"`
}

// ListSchedulesOptions contains optional parameters for listing schedules
type ListSchedulesOptions struct {
	PageSize int
	After    string
}

// ListSchedulesResponse represents the response from listing schedules
type ListSchedulesResponse struct {
	Schedules []Schedule `json:"schedules"`
	ListResponse
}

// ListScheduleEntriesOptions contains parameters for listing schedule entries
type ListScheduleEntriesOptions struct {
	ScheduleID       string
	EntryWindowStart string
	EntryWindowEnd   string
}

// ListScheduleEntriesResponse represents the response from listing schedule entries
type ListScheduleEntriesResponse struct {
	Scheduled []ScheduleEntry `json:"scheduled"`
	Overrides []ScheduleEntry `json:"overrides"`
	Final     []ScheduleEntry `json:"final"`
}

// ListSchedules returns all on-call schedules
func (c *Client) ListSchedules(opts *ListSchedulesOptions) (*ListSchedulesResponse, error) {
	pageSize := 25
	if opts != nil && opts.PageSize > 0 {
		pageSize = opts.PageSize
	}

	params := url.Values{}
	params.Set("page_size", strconv.Itoa(pageSize))
	if opts != nil && opts.After != "" {
		params.Set("after", opts.After)
	}

	respBody, err := c.doRequest("GET", "/schedules", params, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list schedules: %w", err)
	}

	var result ListSchedulesResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schedules response: %w", err)
	}

	return &result, nil
}

// GetSchedule returns a specific schedule by ID
func (c *Client) GetSchedule(id string) (*Schedule, error) {
	if id == "" {
		return nil, fmt.Errorf("schedule ID is required")
	}

	endpoint := fmt.Sprintf("/schedules/%s", url.PathEscape(id))

	respBody, err := c.doRequest("GET", endpoint, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule %s: %w", id, err)
	}

	var result struct {
		Schedule Schedule `json:"schedule"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schedule response: %w", err)
	}

	return &result.Schedule, nil
}

// ListScheduleEntries returns on-call entries for a schedule within a time window
func (c *Client) ListScheduleEntries(opts *ListScheduleEntriesOptions) (*ListScheduleEntriesResponse, error) {
	if opts == nil || opts.ScheduleID == "" {
		return nil, fmt.Errorf("schedule_id is required")
	}

	params := url.Values{}
	params.Set("schedule_id", opts.ScheduleID)
	if opts.EntryWindowStart != "" {
		params.Set("entry_window_start", opts.EntryWindowStart)
	}
	if opts.EntryWindowEnd != "" {
		params.Set("entry_window_end", opts.EntryWindowEnd)
	}

	respBody, err := c.doRequest("GET", "/schedule_entries", params, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list schedule entries: %w", err)
	}

	var result struct {
		ScheduleEntries ListScheduleEntriesResponse `json:"schedule_entries"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schedule entries response: %w", err)
	}

	// Normalize nil slices to empty slices for consistent JSON serialization
	resp := &result.ScheduleEntries
	if resp.Final == nil {
		resp.Final = []ScheduleEntry{}
	}
	if resp.Overrides == nil {
		resp.Overrides = []ScheduleEntry{}
	}
	if resp.Scheduled == nil {
		resp.Scheduled = []ScheduleEntry{}
	}

	return resp, nil
}
