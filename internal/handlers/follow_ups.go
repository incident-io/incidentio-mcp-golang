package handlers

import (
	"context"
	"fmt"

	"github.com/incident-io/incidentio-mcp-golang/internal/client"
)

// ListFollowUpsTool lists follow-ups from incident.io
type ListFollowUpsTool struct {
	*BaseTool
}

func NewListFollowUpsTool(c *client.Client) *ListFollowUpsTool {
	return &ListFollowUpsTool{BaseTool: NewBaseTool(c)}
}

func (t *ListFollowUpsTool) Name() string {
	return "list_follow_ups"
}

func (t *ListFollowUpsTool) Description() string {
	return "List follow-ups for an organization. Follow-ups track work that should be done after an incident.\n\n" +
		"⚠️  CRITICAL WARNING: This API has very limited filtering options and WILL truncate large responses!\n" +
		"🚨 Claude cannot determine if responses are truncated, leading to incorrect data analysis!\n\n" +
		"CRITICAL: If user mentions an incident reference like 'INC-1691', you MUST first resolve it to the full incident ID!\n\n" +
		"WORKFLOW FOR INCIDENT REFERENCES:\n" +
		"1. User says: 'outstanding follow-ups for INC-1691'\n" +
		"2. FIRST: get_incident({\"incident_id\": \"1691\"}) to get the full incident details\n" +
		"3. Extract the full incident ID from the response (e.g., '01K3VHM0T0ZTMG9JPJ9GESB7XX')\n" +
		"4. THEN: list_follow_ups({\"incident_id\": \"01K3VHM0T0ZTMG9JPJ9GESB7XX\"})\n\n" +
		"IMPORTANT: This endpoint requires the FULL incident ID (01K3VHM0T0ZTMG9JPJ9GESB7XX), NOT the short reference (1691)!\n\n" +
		"FILTERING:\n" +
		"- Use incident_id to find all follow-ups for a specific incident (requires FULL incident ID, not reference)\n" +
		"- Use incident_mode to filter by incident type (standard, retrospective, test, tutorial, stream)\n\n" +
		"LIMITATIONS:\n" +
		"- This API does NOT support pagination (no page_size parameter)\n" +
		"- This API does NOT support status filtering (no 'status' parameter)\n" +
		"- Returns ALL follow-ups at once, which may cause response truncation for large datasets\n" +
		"- Only useful filters: incident_id (for specific incidents) and incident_mode (limited value)\n" +
		"- For queries like 'all outstanding follow-ups', Claude must fetch ALL follow-ups and filter client-side\n" +
		"- 🚨 CRITICAL: Claude cannot detect response truncation, so status counts will be WRONG!\n" +
		"- 🚨 NEVER trust status counts from this endpoint - they are likely incomplete due to truncation!\n" +
		"- ALWAYS use incident_id filter when possible to reduce response size\n" +
		"- If response is truncated, use get_follow_up with specific follow-up IDs\n\n" +
		"EXAMPLES:\n" +
		"- User: 'follow-ups for INC-1691' → get_incident({\"incident_id\": \"1691\"}) → list_follow_ups({\"incident_id\": \"01K3VHM0T0ZTMG9JPJ9GESB7XX\"})\n" +
		"- list_follow_ups({\"incident_mode\": \"standard\"}) - Get follow-ups from standard incidents only"
}

func (t *ListFollowUpsTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"incident_id": map[string]interface{}{
				"type":        "string",
				"description": "Filter follow-ups related to this incident ID. IMPORTANT: This must be the FULL incident ID (e.g., '01K3VHM0T0ZTMG9JPJ9GESB7XX'), NOT the short reference (e.g., '1691' from INC-1691). Use get_incident first to resolve references.",
			},
			"incident_mode": map[string]interface{}{
				"type":        "string",
				"description": "Filter by incident mode",
				"enum":        []string{"standard", "retrospective", "test", "tutorial", "stream"},
			},
		},
	}
}

func (t *ListFollowUpsTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	opts := &client.ListFollowUpsOptions{
		IncidentID:   t.ValidateOptionalString(args, "incident_id"),
		IncidentMode: t.ValidateOptionalString(args, "incident_mode"),
	}

	resp, err := t.GetClient().ListFollowUps(ctx, opts)
	if err != nil {
		return "", err
	}

	// Create response with helpful information
	message := "No follow-ups found"
	if len(resp.FollowUps) > 0 {
		message = fmt.Sprintf("Found %d follow-up(s) in this response", len(resp.FollowUps))

		// Add warning for large datasets
		if len(resp.FollowUps) > 50 {
			message += fmt.Sprintf("\n\n⚠️  CRITICAL WARNING: Large dataset detected (%d follow-ups returned).", len(resp.FollowUps))
			message += "\n🚨 RESPONSE LIKELY TRUNCATED - This is NOT the complete dataset!"
			message += "\n📊 The actual total number of follow-ups in your system is likely much higher."
			message += "\n🔍 For accurate results, use incident_id filter or get_follow_up for specific follow-ups."
			message += "\n❌ DO NOT trust status counts or totals from this response."
		} else if len(resp.FollowUps) > 25 {
			message += fmt.Sprintf("\n\n⚠️  WARNING: Moderate dataset (%d follow-ups). Response may be truncated.", len(resp.FollowUps))
			message += "\nConsider using incident_id filter to reduce response size."
		}
	}

	response := t.CreateSimpleResponse(resp.FollowUps, message)
	return t.FormatResponse(response)
}

// GetFollowUpTool retrieves a specific follow-up
type GetFollowUpTool struct {
	*BaseTool
}

func NewGetFollowUpTool(c *client.Client) *GetFollowUpTool {
	return &GetFollowUpTool{BaseTool: NewBaseTool(c)}
}

func (t *GetFollowUpTool) Name() string {
	return "get_follow_up"
}

func (t *GetFollowUpTool) Description() string {
	return "Get details of a specific follow-up by ID. Follow-ups track post-incident action items."
}

func (t *GetFollowUpTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "The follow-up ID",
			},
		},
		"required":             []interface{}{"id"},
		"additionalProperties": false,
	}
}

func (t *GetFollowUpTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	id, err := t.ValidateRequiredString(args, "id")
	if err != nil {
		return "", err
	}

	followUp, err := t.GetClient().GetFollowUp(ctx, id)
	if err != nil {
		return "", err
	}

	return t.FormatResponse(followUp)
}
