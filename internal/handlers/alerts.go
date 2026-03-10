package handlers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/incident-io/incidentio-mcp-golang/internal/client"
)

// ListAlertsTool lists alerts from incident.io
type ListAlertsTool struct {
	apiClient *client.Client
}

func NewListAlertsTool(c *client.Client) *ListAlertsTool {
	return &ListAlertsTool{apiClient: c}
}

func (t *ListAlertsTool) Name() string {
	return "list_alerts"
}

func (t *ListAlertsTool) Description() string {
	return "List alerts from incident.io with optional filters. Returns paginated results.\n\n" +
		"TITLE SEARCH:\n" +
		"Use 'title' for case-insensitive substring search on alert titles. " +
		"Auto-paginates all pages and returns up to 50 matches — no manual pagination needed. " +
		"Combine with alert_source_id or date filters to narrow server-side results.\n\n" +
		"🚨 CRITICAL PAGINATION REQUIREMENT (when not using title search):\n" +
		"1. Start with page_size=25 (API default)\n" +
		"2. Check pagination_meta.after in response\n" +
		"3. If 'after' exists, you MUST call again with after parameter\n" +
		"4. Continue until pagination_meta.after is empty\n" +
		"5. NEVER assume you have all results from just one page!\n\n" +
		"⚠️  WARNING: Claude often stops after first page and reports incomplete data!\n" +
		"📊 Always check if there are more pages before concluding your analysis!\n\n" +
		"FILTERING:\n" +
		"- Use status to filter by alert status (firing, resolved, etc.)\n" +
		"- Use alert_source_id to filter by alert source (server-side)\n" +
		"- Use created_at_gte/created_at_lte to filter by creation date\n" +
		"- Use created_at_date_range for date range filtering (format: '2024-12-02~2024-12-08')\n" +
		"- Use deduplication_key to filter by specific deduplication key\n" +
		"- Multiple status values can be provided as an array\n" +
		"- Combine multiple filters for precise results\n\n" +
		"DATE FILTERING EXAMPLES:\n" +
		"- Single day (e.g., 'August 29th 2025'): created_at_date_range=\"2025-08-29~2025-08-29\" (PREFERRED)\n" +
		"- Date range: created_at_date_range=\"2025-01-01~2025-01-15\"\n" +
		"- Past week: created_at_gte=\"2025-01-08\" (calculate 7 days ago)\n" +
		"- Past month: created_at_gte=\"2025-01-15\" (calculate 30 days ago)\n" +
		"- Before date: created_at_lte=\"2025-01-15\"\n" +
		"- Date format: '2025-08-29' (date only, no time)\n\n" +
		"PAGINATION EXAMPLES:\n" +
		"Title search: list_alerts({\"title\": \"Sentry\", \"created_at_gte\": \"2025-01-08\"})\n" +
		"Manual pagination:\n" +
		"1. list_alerts({\"created_at_gte\": \"2025-01-08\", \"page_size\": 25})\n" +
		"2. If pagination_meta.after exists, call again with after parameter\n" +
		"3. Continue until pagination_meta.after is empty\n" +
		"4. Combine all results for complete analysis\n\n" +
		"IMPORTANT: This endpoint returns ALL alerts in your organization. Use date and status filtering to reduce results."
}

func (t *ListAlertsTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"title": map[string]interface{}{
				"type":        "string",
				"description": "Case-insensitive substring search on alert title. Auto-paginates all pages and returns up to 50 matches. When set, page_size and after are ignored.",
			},
			"page_size": map[string]interface{}{
				"type":        "integer",
				"description": "Number of results per page (1-50). Default is 25 (API default). Use smaller values if you need fewer results.",
				"default":     25,
				"minimum":     1,
				"maximum":     50,
			},
			"after": map[string]interface{}{
				"type":        "string",
				"description": "Pagination cursor from previous response's 'pagination_meta.after' field. Use this to fetch the next page of results.",
			},
			"status": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Filter by alert status. Common values: 'firing', 'resolved'. Multiple values can be provided.",
			},
			"deduplication_key": map[string]interface{}{
				"type":        "string",
				"description": "Filter by deduplication key. Exact match required.",
			},
			"alert_source_id": map[string]interface{}{
				"type":        "string",
				"description": "Filter by alert source ID (exact match, server-side).",
			},
			"created_at_gte": map[string]interface{}{
				"type":        "string",
				"description": "Filter alerts created on or after this date. Format: '2025-01-15' (date only, no time)",
			},
			"created_at_lte": map[string]interface{}{
				"type":        "string",
				"description": "Filter alerts created on or before this date. Format: '2025-01-15' (date only, no time)",
			},
			"created_at_date_range": map[string]interface{}{
				"type":        "string",
				"description": "Filter alerts created within date range. Format: '2025-08-29~2025-08-29' (tilde-separated dates). PREFERRED for single day queries.",
			},
		},
	}
}

func (t *ListAlertsTool) Execute(args map[string]interface{}) (string, error) {
	titleFilter := strings.ToLower(GetStringArg(args, "title"))

	// Server-side filters
	opts := &client.ListAlertsOptions{}
	opts.Status = GetStringArrayArg(args, "status")
	if dk := GetStringArg(args, "deduplication_key"); dk != "" {
		opts.DeduplicationKey = dk
	}
	if asi := GetStringArg(args, "alert_source_id"); asi != "" {
		opts.AlertSourceID = asi
	}
	opts.CreatedAtGte = GetStringArg(args, "created_at_gte")
	opts.CreatedAtLte = GetStringArg(args, "created_at_lte")
	opts.CreatedAtDateRange = GetStringArg(args, "created_at_date_range")

	// Auto-paginating title search
	if titleFilter != "" {
		opts.PageSize = 50 // max API page size
		var matched []client.Alert
		pagesScanned := 0

		for pagesScanned < maxSearchPages {
			resp, err := t.apiClient.ListAlerts(opts)
			if err != nil {
				return "", err
			}
			pagesScanned++

			for i := range resp.Alerts {
				if strings.Contains(strings.ToLower(resp.Alerts[i].Title), titleFilter) {
					matched = append(matched, resp.Alerts[i])
					if len(matched) >= maxSearchResults {
						break
					}
				}
			}

			if len(matched) >= maxSearchResults || resp.PaginationMeta.After == "" {
				break
			}
			opts.After = resp.PaginationMeta.After
		}

		response := map[string]interface{}{
			"alerts":        matched,
			"count":         len(matched),
			"title_filter":  titleFilter,
			"pages_scanned": pagesScanned,
		}
		if len(matched) >= maxSearchResults {
			response["truncated"] = true
			response["note"] = fmt.Sprintf("Result limit (%d) reached. Narrow with date/status filters.", maxSearchResults)
		} else if pagesScanned >= maxSearchPages {
			response["scan_truncated"] = true
			response["note"] = fmt.Sprintf("Page limit (%d) reached. Use date or alert_source_id filters to narrow results.", maxSearchPages)
		}
		return FormatJSONResponse(response)
	}

	// Standard single-page mode
	opts.PageSize = GetIntArg(args, "page_size", 25)
	opts.After = GetStringArg(args, "after")

	resp, err := t.apiClient.ListAlerts(opts)
	if err != nil {
		return "", err
	}

	// Create response with prominent pagination info
	response := map[string]interface{}{
		"alerts":          resp.Alerts,
		"pagination_meta": resp.PaginationMeta,
		"count":           len(resp.Alerts),
	}

	// Add prominent pagination warnings
	if resp.PaginationMeta.After != "" {
		response["🚨 PAGINATION WARNING"] = "MORE RESULTS AVAILABLE - This is NOT the complete dataset!"
		response["📊 FETCH_NEXT_PAGE"] = map[string]interface{}{
			"action":  "Call list_alerts again with after parameter",
			"after":   resp.PaginationMeta.After,
			"message": "You MUST fetch all pages to get complete results!",
		}
		response["⚠️  INCOMPLETE_DATA"] = fmt.Sprintf("Only showing %d alerts from this page. Total results likely much higher.", len(resp.Alerts))
	} else {
		response["✅ COMPLETE"] = "No more pages - this appears to be the complete dataset"
	}

	return FormatJSONResponse(response)
}

// GetAlertTool retrieves a specific alert
type GetAlertTool struct {
	apiClient *client.Client
}

func NewGetAlertTool(c *client.Client) *GetAlertTool {
	return &GetAlertTool{apiClient: c}
}

func (t *GetAlertTool) Name() string {
	return "get_alert"
}

func (t *GetAlertTool) Description() string {
	return "Get details of a specific alert by ID"
}

func (t *GetAlertTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "The alert ID",
			},
		},
		"required": []string{"id"},
	}
}

func (t *GetAlertTool) Execute(args map[string]interface{}) (string, error) {
	id, ok := args["id"].(string)
	if !ok {
		return "", fmt.Errorf("id parameter is required")
	}

	alert, err := t.apiClient.GetAlert(id)
	if err != nil {
		return "", err
	}

	result, err := json.MarshalIndent(alert, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	return string(result), nil
}

// ListIncidentAlertsTool lists connections between incidents and alerts
type ListIncidentAlertsTool struct {
	apiClient *client.Client
}

func NewListIncidentAlertsTool(c *client.Client) *ListIncidentAlertsTool {
	return &ListIncidentAlertsTool{apiClient: c}
}

func (t *ListIncidentAlertsTool) Name() string {
	return "list_incident_alerts"
}

func (t *ListIncidentAlertsTool) Description() string {
	return "List connections between incidents and alerts. Returns paginated results.\n\n" +
		"PAGINATION: Start with page_size=25 (API default). If pagination_meta.after exists, call again with after parameter to get next page. Continue until pagination_meta.after is empty.\n\n" +
		"FILTERING: Optional filters to narrow results.\n" +
		"- Use incident_id to find all alerts attached to a specific incident (requires FULL incident ID)\n" +
		"- Use alert_id to find which incident a specific alert triggered\n" +
		"- If no filters provided, returns ALL incident-alert connections (may be many results)\n\n" +
		"IMPORTANT: This endpoint requires the FULL incident ID (01K3VHM0T0ZTMG9JPJ9GESB7XX), NOT the short reference (1691)!\n\n" +
		"WORKFLOW FOR INCIDENT REFERENCES:\n" +
		"1. User says: 'alerts for INC-1691'\n" +
		"2. FIRST: get_incident({\"incident_id\": \"1691\"}) to get the full incident details\n" +
		"3. Extract the full incident ID from the response (e.g., '01K3VHM0T0ZTMG9JPJ9GESB7XX')\n" +
		"4. THEN: list_incident_alerts({\"incident_id\": \"01K3VHM0T0ZTMG9JPJ9GESB7XX\"})"
}

func (t *ListIncidentAlertsTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"incident_id": map[string]interface{}{
				"type":        "string",
				"description": "Filter by incident ID to find all alerts attached to this incident. REQUIRES FULL incident ID (01K3VHM0T0ZTMG9JPJ9GESB7XX), NOT short reference (1691).",
			},
			"alert_id": map[string]interface{}{
				"type":        "string",
				"description": "Filter by alert ID to find which incident this alert triggered",
			},
			"page_size": map[string]interface{}{
				"type":        "integer",
				"description": "Number of results per page (1-50). Default is 25.",
				"default":     25,
				"minimum":     1,
				"maximum":     50,
			},
			"after": map[string]interface{}{
				"type":        "string",
				"description": "Pagination cursor from previous response's 'pagination_meta.after' field. Use this to fetch the next page of results.",
			},
		},
	}
}

func (t *ListIncidentAlertsTool) Execute(args map[string]interface{}) (string, error) {
	opts := &client.ListIncidentAlertsOptions{
		PageSize: 25, // Default page size
	}

	if pageSize, ok := args["page_size"].(float64); ok {
		opts.PageSize = int(pageSize)
	}

	if after, ok := args["after"].(string); ok && after != "" {
		opts.After = after
	}

	if incidentID, ok := args["incident_id"].(string); ok && incidentID != "" {
		opts.IncidentID = incidentID
	}

	if alertID, ok := args["alert_id"].(string); ok && alertID != "" {
		opts.AlertID = alertID
	}

	// Note: According to API docs, both incident_id and alert_id are optional
	// If neither is provided, all incident-alert connections will be returned

	resp, err := t.apiClient.ListIncidentAlerts(opts)
	if err != nil {
		return "", err
	}

	// Create response with prominent pagination info
	response := map[string]interface{}{
		"incident_alerts": resp.IncidentAlerts,
		"pagination_meta": resp.PaginationMeta,
		"count":           len(resp.IncidentAlerts),
	}

	// Add helpful message if there are more pages
	if resp.PaginationMeta.After != "" {
		response["next_page_hint"] = fmt.Sprintf("More results available. Use after='%s' to fetch the next page.", resp.PaginationMeta.After)
	}

	return FormatJSONResponse(response)
}
