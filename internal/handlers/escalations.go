package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/incident-io/incidentio-mcp-golang/internal/client"
)

// --- Escalation paths (Escalations V2) ---

type ListEscalationPathsTool struct {
	apiClient *client.Client
}

func NewListEscalationPathsTool(c *client.Client) *ListEscalationPathsTool {
	return &ListEscalationPathsTool{apiClient: c}
}

func (t *ListEscalationPathsTool) Name() string { return "list_escalation_paths" }

func (t *ListEscalationPathsTool) Description() string {
	return "List escalation paths in your account (On-call). Pagination: use pagination_meta.after on the next call. " +
		"See https://docs.incident.io/api-reference/escalations-v2/listpaths"
}

func (t *ListEscalationPathsTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"page_size": map[string]interface{}{
				"type":        "integer",
				"description": "Page size (1–25; API maximum 25)",
				"minimum":     1,
				"maximum":     25,
			},
			"after": map[string]interface{}{
				"type":        "string",
				"description": "Pagination cursor from pagination_meta.after",
			},
		},
		"additionalProperties": false,
	}
}

func (t *ListEscalationPathsTool) Execute(args map[string]interface{}) (string, error) {
	params := &client.ListEscalationPathsParams{}
	if pageSize, ok := args["page_size"].(float64); ok {
		params.PageSize = int(pageSize)
	}
	if after, ok := args["after"].(string); ok {
		params.After = after
	}
	result, err := t.apiClient.ListEscalationPaths(params)
	if err != nil {
		return "", fmt.Errorf("failed to list escalation paths: %w", err)
	}
	out, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}
	return string(out), nil
}

type GetEscalationPathTool struct {
	apiClient *client.Client
}

func NewGetEscalationPathTool(c *client.Client) *GetEscalationPathTool {
	return &GetEscalationPathTool{apiClient: c}
}

func (t *GetEscalationPathTool) Name() string { return "get_escalation_path" }

func (t *GetEscalationPathTool) Description() string {
	return "Get a single escalation path by ID. See https://docs.incident.io/api-reference/escalations-v2/showpath"
}

func (t *GetEscalationPathTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "Escalation path ID",
				"minLength":   1,
			},
		},
		"required":             []string{"id"},
		"additionalProperties": false,
	}
}

func (t *GetEscalationPathTool) Execute(args map[string]interface{}) (string, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("id is required")
	}
	raw, err := t.apiClient.GetEscalationPath(id)
	if err != nil {
		return "", fmt.Errorf("failed to get escalation path: %w", err)
	}
	var pretty interface{}
	if err := json.Unmarshal(raw, &pretty); err != nil {
		return "", fmt.Errorf("failed to decode escalation path: %w", err)
	}
	out, err := json.MarshalIndent(map[string]interface{}{"escalation_path": pretty}, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}
	return string(out), nil
}

type CreateEscalationPathTool struct {
	apiClient *client.Client
}

func NewCreateEscalationPathTool(c *client.Client) *CreateEscalationPathTool {
	return &CreateEscalationPathTool{apiClient: c}
}

func (t *CreateEscalationPathTool) Name() string { return "create_escalation_path" }

func (t *CreateEscalationPathTool) Description() string {
	return "Create an escalation path. `payload` must match the API (name, path, team_ids, repeat_config, working_hours, …). " +
		"Prefer building paths in the incident.io UI for complex graphs. " +
		"See https://docs.incident.io/api-reference/escalations-v2/createpath"
}

func (t *CreateEscalationPathTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"payload": map[string]interface{}{
				"type":        "object",
				"description": "Request body as defined by Escalations V2 CreatePath (name, path, team_ids, etc.)",
			},
		},
		"required":             []string{"payload"},
		"additionalProperties": false,
	}
}

func (t *CreateEscalationPathTool) Execute(args map[string]interface{}) (string, error) {
	payload, ok := args["payload"].(map[string]interface{})
	if !ok || payload == nil {
		return "", fmt.Errorf("payload must be an object")
	}
	raw, err := t.apiClient.CreateEscalationPath(payload)
	if err != nil {
		return "", fmt.Errorf("failed to create escalation path: %w", err)
	}
	var pretty interface{}
	if err := json.Unmarshal(raw, &pretty); err != nil {
		return "", fmt.Errorf("failed to decode escalation path: %w", err)
	}
	out, err := json.MarshalIndent(map[string]interface{}{"escalation_path": pretty}, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}
	return string(out), nil
}

type UpdateEscalationPathTool struct {
	apiClient *client.Client
}

func NewUpdateEscalationPathTool(c *client.Client) *UpdateEscalationPathTool {
	return &UpdateEscalationPathTool{apiClient: c}
}

func (t *UpdateEscalationPathTool) Name() string { return "update_escalation_path" }

func (t *UpdateEscalationPathTool) Description() string {
	return "Replace an escalation path (HTTP PUT). `payload` is the full path document. " +
		"See https://docs.incident.io/api-reference/escalations-v2/updatepath"
}

func (t *UpdateEscalationPathTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "Escalation path ID",
				"minLength":   1,
			},
			"payload": map[string]interface{}{
				"type":        "object",
				"description": "Full Escalations V2 path payload (name, path, team_ids, repeat_config, working_hours, …)",
			},
		},
		"required":             []string{"id", "payload"},
		"additionalProperties": false,
	}
}

func (t *UpdateEscalationPathTool) Execute(args map[string]interface{}) (string, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("id is required")
	}
	payload, ok := args["payload"].(map[string]interface{})
	if !ok || payload == nil {
		return "", fmt.Errorf("payload must be an object")
	}
	raw, err := t.apiClient.UpdateEscalationPath(id, payload)
	if err != nil {
		return "", fmt.Errorf("failed to update escalation path: %w", err)
	}
	var pretty interface{}
	if err := json.Unmarshal(raw, &pretty); err != nil {
		return "", fmt.Errorf("failed to decode escalation path: %w", err)
	}
	out, err := json.MarshalIndent(map[string]interface{}{"escalation_path": pretty}, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}
	return string(out), nil
}

type DestroyEscalationPathTool struct {
	apiClient *client.Client
}

func NewDestroyEscalationPathTool(c *client.Client) *DestroyEscalationPathTool {
	return &DestroyEscalationPathTool{apiClient: c}
}

func (t *DestroyEscalationPathTool) Name() string { return "destroy_escalation_path" }

func (t *DestroyEscalationPathTool) Description() string {
	return "Archive (delete) an escalation path. Returns success with no resource body (HTTP 204). " +
		"See https://docs.incident.io/api-reference/escalations-v2/destroypath"
}

func (t *DestroyEscalationPathTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "Escalation path ID to archive",
				"minLength":   1,
			},
		},
		"required":             []string{"id"},
		"additionalProperties": false,
	}
}

func (t *DestroyEscalationPathTool) Execute(args map[string]interface{}) (string, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("id is required")
	}
	if err := t.apiClient.DestroyEscalationPath(id); err != nil {
		return "", fmt.Errorf("failed to destroy escalation path: %w", err)
	}
	out, err := json.MarshalIndent(map[string]interface{}{
		"archived": true,
		"id":       id,
	}, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}
	return string(out), nil
}

// --- Escalations (instances) ---

type ListEscalationsTool struct {
	apiClient *client.Client
}

func NewListEscalationsTool(c *client.Client) *ListEscalationsTool {
	return &ListEscalationsTool{apiClient: c}
}

func (t *ListEscalationsTool) Name() string { return "list_escalations" }

func (t *ListEscalationsTool) Description() string {
	return "List escalations with optional filters (status, escalation path, alert, dates, idempotency key). " +
		"Status values: pending, triggered, acked, resolved, expired, cancelled. " +
		"See https://docs.incident.io/api-reference/escalations-v2/list"
}

func (t *ListEscalationsTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"page_size": map[string]interface{}{
				"type":        "integer",
				"description": "Page size (1–50; API maximum 50)",
				"minimum":     1,
				"maximum":     50,
			},
			"after": map[string]interface{}{
				"type":        "string",
				"description": "Pagination cursor from pagination_meta.after",
			},
			"escalation_path_one_of": map[string]interface{}{
				"type":        "array",
				"description": "Escalation path IDs (escalation_path[one_of])",
				"items":       map[string]interface{}{"type": "string"},
			},
			"escalation_path_not_in": map[string]interface{}{
				"type":        "array",
				"description": "Escalation path IDs to exclude (escalation_path[not_in])",
				"items":       map[string]interface{}{"type": "string"},
			},
			"status_one_of": map[string]interface{}{
				"type":        "array",
				"description": "Statuses to include (status[one_of])",
				"items":       map[string]interface{}{"type": "string"},
			},
			"status_not_in": map[string]interface{}{
				"type":        "array",
				"description": "Statuses to exclude (status[not_in])",
				"items":       map[string]interface{}{"type": "string"},
			},
			"alert_one_of": map[string]interface{}{
				"type":        "array",
				"description": "Alert IDs (alert[one_of])",
				"items":       map[string]interface{}{"type": "string"},
			},
			"alert_not_in": map[string]interface{}{
				"type":        "array",
				"description": "Alert IDs to exclude (alert[not_in])",
				"items":       map[string]interface{}{"type": "string"},
			},
			"created_at_gte": map[string]interface{}{
				"type":        "string",
				"description": "created_at[gte] (e.g. 2025-01-01)",
			},
			"created_at_lte": map[string]interface{}{
				"type":        "string",
				"description": "created_at[lte]",
			},
			"created_at_date_range": map[string]interface{}{
				"type":        "string",
				"description": "created_at[date_range] as start~end (e.g. 2025-01-01~2025-01-31)",
			},
			"updated_at_gte": map[string]interface{}{
				"type":        "string",
				"description": "updated_at[gte]",
			},
			"updated_at_lte": map[string]interface{}{
				"type":        "string",
				"description": "updated_at[lte]",
			},
			"updated_at_date_range": map[string]interface{}{
				"type":        "string",
				"description": "updated_at[date_range] as start~end",
			},
			"idempotency_key_is": map[string]interface{}{
				"type":        "string",
				"description": "idempotency_key[is] exact match",
			},
			"idempotency_key_starts_with": map[string]interface{}{
				"type":        "string",
				"description": "idempotency_key[starts_with] prefix",
			},
		},
		"additionalProperties": false,
	}
}

func stringSliceArg(args map[string]interface{}, key string) []string {
	raw, ok := args[key].([]interface{})
	if !ok {
		return nil
	}
	var out []string
	for _, v := range raw {
		if s, ok := v.(string); ok && s != "" {
			out = append(out, s)
		}
	}
	return out
}

func (t *ListEscalationsTool) Execute(args map[string]interface{}) (string, error) {
	params := &client.ListEscalationsParams{}
	if pageSize, ok := args["page_size"].(float64); ok {
		params.PageSize = int(pageSize)
	}
	if after, ok := args["after"].(string); ok {
		params.After = after
	}
	params.EscalationPathOneOf = stringSliceArg(args, "escalation_path_one_of")
	params.EscalationPathNotIn = stringSliceArg(args, "escalation_path_not_in")
	params.StatusOneOf = stringSliceArg(args, "status_one_of")
	params.StatusNotIn = stringSliceArg(args, "status_not_in")
	params.AlertOneOf = stringSliceArg(args, "alert_one_of")
	params.AlertNotIn = stringSliceArg(args, "alert_not_in")
	if s, ok := args["created_at_gte"].(string); ok {
		params.CreatedAtGte = s
	}
	if s, ok := args["created_at_lte"].(string); ok {
		params.CreatedAtLte = s
	}
	if s, ok := args["created_at_date_range"].(string); ok {
		params.CreatedAtDateRange = s
	}
	if s, ok := args["updated_at_gte"].(string); ok {
		params.UpdatedAtGte = s
	}
	if s, ok := args["updated_at_lte"].(string); ok {
		params.UpdatedAtLte = s
	}
	if s, ok := args["updated_at_date_range"].(string); ok {
		params.UpdatedAtDateRange = s
	}
	if s, ok := args["idempotency_key_is"].(string); ok {
		params.IdempotencyKeyIs = s
	}
	if s, ok := args["idempotency_key_starts_with"].(string); ok {
		params.IdempotencyKeyStartsWith = s
	}

	result, err := t.apiClient.ListEscalations(params)
	if err != nil {
		return "", fmt.Errorf("failed to list escalations: %w", err)
	}
	out, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}
	return string(out), nil
}

type GetEscalationTool struct {
	apiClient *client.Client
}

func NewGetEscalationTool(c *client.Client) *GetEscalationTool {
	return &GetEscalationTool{apiClient: c}
}

func (t *GetEscalationTool) Name() string { return "get_escalation" }

func (t *GetEscalationTool) Description() string {
	return "Get one escalation by ID. See https://docs.incident.io/api-reference/escalations-v2/show"
}

func (t *GetEscalationTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "Escalation ID",
				"minLength":   1,
			},
		},
		"required":             []string{"id"},
		"additionalProperties": false,
	}
}

func (t *GetEscalationTool) Execute(args map[string]interface{}) (string, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("id is required")
	}
	raw, err := t.apiClient.GetEscalation(id)
	if err != nil {
		return "", fmt.Errorf("failed to get escalation: %w", err)
	}
	var pretty interface{}
	if err := json.Unmarshal(raw, &pretty); err != nil {
		return "", fmt.Errorf("failed to decode escalation: %w", err)
	}
	out, err := json.MarshalIndent(map[string]interface{}{"escalation": pretty}, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}
	return string(out), nil
}

type CreateEscalationTool struct {
	apiClient *client.Client
}

func NewCreateEscalationTool(c *client.Client) *CreateEscalationTool {
	return &CreateEscalationTool{apiClient: c}
}

func (t *CreateEscalationTool) Name() string { return "create_escalation" }

func (t *CreateEscalationTool) Description() string {
	return "Create an escalation: page via escalation_path_id OR user_ids (mutually exclusive). " +
		"Optional description, idempotency_key. Rate-limited (interactive use). " +
		"See https://docs.incident.io/api-reference/escalations-v2/create"
}

func (t *CreateEscalationTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"title": map[string]interface{}{
				"type":        "string",
				"description": "Escalation title",
				"minLength":   1,
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Additional details",
			},
			"escalation_path_id": map[string]interface{}{
				"type":        "string",
				"description": "Follow this escalation path (do not set with user_ids)",
			},
			"user_ids": map[string]interface{}{
				"type":        "array",
				"description": "User IDs to page directly (do not set with escalation_path_id)",
				"items":       map[string]interface{}{"type": "string"},
			},
			"idempotency_key": map[string]interface{}{
				"type":        "string",
				"description": "Dedupe key; same key returns existing escalation",
			},
		},
		"required":             []string{"title"},
		"additionalProperties": false,
	}
}

func (t *CreateEscalationTool) Execute(args map[string]interface{}) (string, error) {
	title, ok := args["title"].(string)
	if !ok || title == "" {
		return "", fmt.Errorf("title is required")
	}
	req := &client.CreateEscalationRequest{Title: title}
	if d, ok := args["description"].(string); ok {
		req.Description = d
	}
	if eid, ok := args["escalation_path_id"].(string); ok {
		req.EscalationPathID = eid
	}
	if raw, ok := args["user_ids"].([]interface{}); ok {
		for _, v := range raw {
			if s, ok := v.(string); ok && s != "" {
				req.UserIDs = append(req.UserIDs, s)
			}
		}
	}
	if k, ok := args["idempotency_key"].(string); ok {
		req.IdempotencyKey = k
	}

	hasPath := req.EscalationPathID != ""
	hasUsers := len(req.UserIDs) > 0
	if hasPath == hasUsers {
		return "", fmt.Errorf("set exactly one of escalation_path_id or user_ids (non-empty)")
	}

	raw, err := t.apiClient.CreateEscalation(req)
	if err != nil {
		return "", fmt.Errorf("failed to create escalation: %w", err)
	}
	var pretty interface{}
	if err := json.Unmarshal(raw, &pretty); err != nil {
		return "", fmt.Errorf("failed to decode escalation: %w", err)
	}
	out, err := json.MarshalIndent(map[string]interface{}{"escalation": pretty}, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}
	return string(out), nil
}
