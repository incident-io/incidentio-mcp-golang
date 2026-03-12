package handlers

import (
	"context"
	"fmt"

	"github.com/incident-io/incidentio-mcp-golang/internal/client"
)

// ListPostmortemsTool lists postmortem documents from incident.io
type ListPostmortemsTool struct {
	*BaseTool
}

func NewListPostmortemsTool(c *client.Client) *ListPostmortemsTool {
	return &ListPostmortemsTool{BaseTool: NewBaseTool(c)}
}

func (t *ListPostmortemsTool) Name() string {
	return "list_postmortems"
}

func (t *ListPostmortemsTool) Description() string {
	return "List postmortem documents for the organization. Returns metadata only (title, status, editors, etc.).\n\n" +
		"To get the full content of a postmortem, use get_postmortem_content with the document ID.\n\n" +
		"FILTERING:\n" +
		"- Use incident_id to find the postmortem for a specific incident\n" +
		"- Use sort_by to order results by creation date\n\n" +
		"STATUS VALUES:\n" +
		"- in_progress: Still being written\n" +
		"- in_review: Ready for review\n" +
		"- completed: Finalized\n\n" +
		"PAGINATION:\n" +
		"- Default page_size is 25, max is 250\n" +
		"- Use 'after' cursor from response to fetch next page\n\n" +
		"EXAMPLES:\n" +
		"- list_postmortems() - Get all postmortems\n" +
		"- list_postmortems({\"incident_id\": \"01ABC123...\"}) - Get postmortem for specific incident\n" +
		"- list_postmortems({\"sort_by\": \"created_at_oldest_first\"}) - Get oldest postmortems first"
}

func (t *ListPostmortemsTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"page_size": map[string]interface{}{
				"type":        "integer",
				"description": "Number of results per page (1-250). Default is 25.",
				"minimum":     1,
				"maximum":     250,
			},
			"after": map[string]interface{}{
				"type":        "string",
				"description": "Pagination cursor from previous response's 'pagination_meta.after' field.",
			},
			"incident_id": map[string]interface{}{
				"type":        "string",
				"description": "Filter postmortems by incident ID. Use the full incident ID (e.g., '01K3VHM0T0ZTMG9JPJ9GESB7XX').",
			},
			"sort_by": map[string]interface{}{
				"type":        "string",
				"description": "Sort order for results",
				"enum":        []string{"created_at_newest_first", "created_at_oldest_first"},
				"default":     "created_at_newest_first",
			},
		},
	}
}

func (t *ListPostmortemsTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	opts := &client.ListPostmortemsOptions{
		PageSize:   t.ValidateOptionalInt(args, "page_size", 25),
		After:      t.ValidateOptionalString(args, "after"),
		IncidentID: t.ValidateOptionalString(args, "incident_id"),
		SortBy:     t.ValidateOptionalString(args, "sort_by"),
	}

	resp, err := t.GetClient().ListPostmortems(ctx, opts)
	if err != nil {
		return "", err
	}

	message := "No postmortem documents found"
	if len(resp.PostmortemDocuments) > 0 {
		message = fmt.Sprintf("Found %d postmortem document(s)", len(resp.PostmortemDocuments))
	}

	response := t.CreatePaginationResponse(
		resp.PostmortemDocuments,
		resp.PaginationMeta,
		len(resp.PostmortemDocuments),
	)
	response["message"] = message

	return t.FormatResponse(response)
}

// GetPostmortemTool retrieves a specific postmortem document
type GetPostmortemTool struct {
	*BaseTool
}

func NewGetPostmortemTool(c *client.Client) *GetPostmortemTool {
	return &GetPostmortemTool{BaseTool: NewBaseTool(c)}
}

func (t *GetPostmortemTool) Name() string {
	return "get_postmortem"
}

func (t *GetPostmortemTool) Description() string {
	return "Get details of a specific postmortem document by ID. Returns metadata including title, status, editors, and URLs.\n\n" +
		"This returns metadata only. To get the full markdown content, use get_postmortem_content.\n\n" +
		"RESPONSE INCLUDES:\n" +
		"- id: Unique identifier\n" +
		"- incident_id: Associated incident\n" +
		"- title: Document title\n" +
		"- status: in_progress, in_review, or completed\n" +
		"- document_url: Link to view in dashboard\n" +
		"- exported_urls: External export locations (e.g., Notion, Confluence)\n" +
		"- editors: Users who have edited the document"
}

func (t *GetPostmortemTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "The postmortem document ID (e.g., '01GDZEW57FDA1K4S63MGMQ5DS9')",
			},
		},
		"required":             []interface{}{"id"},
		"additionalProperties": false,
	}
}

func (t *GetPostmortemTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	id, err := t.ValidateRequiredString(args, "id")
	if err != nil {
		return "", err
	}

	postmortem, err := t.GetClient().GetPostmortem(ctx, id)
	if err != nil {
		return "", err
	}

	return t.FormatResponse(postmortem)
}

// GetPostmortemContentTool retrieves the markdown content of a postmortem document
type GetPostmortemContentTool struct {
	*BaseTool
}

func NewGetPostmortemContentTool(c *client.Client) *GetPostmortemContentTool {
	return &GetPostmortemContentTool{BaseTool: NewBaseTool(c)}
}

func (t *GetPostmortemContentTool) Name() string {
	return "get_postmortem_content"
}

func (t *GetPostmortemContentTool) Description() string {
	return "Get the full markdown content of a postmortem document.\n\n" +
		"This returns the complete document content as markdown, which typically includes:\n" +
		"- Summary of what happened\n" +
		"- Timeline of events\n" +
		"- Root cause analysis\n" +
		"- Impact assessment\n" +
		"- Action items and follow-ups\n" +
		"- Lessons learned\n\n" +
		"Use get_postmortem or list_postmortems first to get the document ID.\n\n" +
		"EXAMPLE WORKFLOW:\n" +
		"1. list_postmortems({\"incident_id\": \"01ABC...\"}) - Find the postmortem for an incident\n" +
		"2. get_postmortem_content({\"id\": \"01XYZ...\"}) - Get the full content"
}

func (t *GetPostmortemContentTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "The postmortem document ID (e.g., '01GDZEW57FDA1K4S63MGMQ5DS9')",
			},
		},
		"required":             []interface{}{"id"},
		"additionalProperties": false,
	}
}

func (t *GetPostmortemContentTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	id, err := t.ValidateRequiredString(args, "id")
	if err != nil {
		return "", err
	}

	content, err := t.GetClient().GetPostmortemContent(ctx, id)
	if err != nil {
		return "", err
	}

	response := map[string]interface{}{
		"postmortem_id": id,
		"markdown":      content.Markdown,
	}

	return t.FormatResponse(response)
}
