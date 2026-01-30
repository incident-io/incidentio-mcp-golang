package client

import "time"

// Incident represents an incident in incident.io
type Incident struct {
	ID                      string             `json:"id"`
	Reference               string             `json:"reference"`
	Name                    string             `json:"name"`
	Summary                 string             `json:"summary,omitempty"`
	Permalink               string             `json:"permalink"`
	IncidentStatus          IncidentStatus     `json:"incident_status"`
	Severity                Severity           `json:"severity"`
	IncidentType            IncidentType       `json:"incident_type"`
	Mode                    string             `json:"mode"`
	Visibility              string             `json:"visibility"`
	CreatedAt               time.Time          `json:"created_at"`
	UpdatedAt               time.Time          `json:"updated_at"`
	SlackTeamID             string             `json:"slack_team_id,omitempty"`
	SlackChannelID          string             `json:"slack_channel_id,omitempty"`
	SlackChannelName        string             `json:"slack_channel_name,omitempty"`
	IncidentRoleAssignments []RoleAssignment   `json:"incident_role_assignments"`
	CustomFieldEntries      []CustomFieldEntry `json:"custom_field_entries"`
	HasDebrief              bool               `json:"has_debrief"`
}

// IncidentStatus represents the status of an incident
type IncidentStatus struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Rank        int       `json:"rank"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Severity represents the severity of an incident
type Severity struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Rank        int       `json:"rank"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// IncidentType represents the type of an incident
type IncidentType struct {
	ID                   string    `json:"id"`
	Name                 string    `json:"name"`
	Description          string    `json:"description"`
	IsDefault            bool      `json:"is_default"`
	PrivateIncidentsOnly bool      `json:"private_incidents_only"`
	CreateInTriage       string    `json:"create_in_triage"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// RoleAssignment represents a role assignment in an incident
type RoleAssignment struct {
	Role struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		Shortform    string `json:"shortform"`
		Description  string `json:"description"`
		Instructions string `json:"instructions"`
		RoleType     string `json:"role_type"`
		Required     bool   `json:"required"`
		CreatedAt    string `json:"created_at"`
		UpdatedAt    string `json:"updated_at"`
	} `json:"role"`
	Assignee *User `json:"assignee,omitempty"`
}

// User represents a user in incident.io
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// CustomFieldEntry represents a custom field entry
type CustomFieldEntry struct {
	CustomField struct {
		ID          string        `json:"id"`
		Name        string        `json:"name"`
		Description string        `json:"description"`
		FieldType   string        `json:"field_type"`
		Options     []interface{} `json:"options"`
	} `json:"custom_field"`
	Values []interface{} `json:"values"`
}

// AlertAttribute represents an alert attribute
type AlertAttribute struct {
	ArrayValue []AlertAttributeValue `json:"array_value,omitempty"`
	Attribute  AlertAttributeDef     `json:"attribute"`
	Value      AlertAttributeValue   `json:"value"`
}

// AlertAttributeDef represents an alert attribute definition
type AlertAttributeDef struct {
	Array    bool   `json:"array"`
	ID       string `json:"id"`
	Name     string `json:"name"`
	Required bool   `json:"required"`
	Type     string `json:"type"`
}

// AlertAttributeValue represents an alert attribute value
type AlertAttributeValue struct {
	CatalogEntry *CatalogEntry `json:"catalog_entry,omitempty"`
	Label        string        `json:"label"`
	Literal      string        `json:"literal"`
}

// Alert represents an alert in incident.io
type Alert struct {
	ID               string           `json:"id"`
	AlertSourceID    string           `json:"alert_source_id"`
	Attributes       []AlertAttribute `json:"attributes"`
	CreatedAt        time.Time        `json:"created_at"`
	DeduplicationKey string           `json:"deduplication_key"`
	Description      string           `json:"description"`
	ResolvedAt       *time.Time       `json:"resolved_at,omitempty"`
	SourceURL        string           `json:"source_url"`
	Status           string           `json:"status"`
	Title            string           `json:"title"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

// IncidentAlert represents the connection between an incident and an alert
type IncidentAlert struct {
	Alert        Alert    `json:"alert"`
	AlertRouteID string   `json:"alert_route_id"`
	ID           string   `json:"id"`
	Incident     Incident `json:"incident"`
}

// Action represents an action in incident.io
type Action struct {
	ID          string     `json:"id"`
	IncidentID  string     `json:"incident_id"`
	Status      string     `json:"status"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Assignee    *User      `json:"assignee,omitempty"`
}

// WorkflowTrigger represents a workflow trigger in incident.io
type WorkflowTrigger struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}

// Workflow represents a workflow in incident.io
type Workflow struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Trigger   WorkflowTrigger `json:"trigger"`
	Enabled   bool            `json:"enabled"`
	Runs      []WorkflowRun   `json:"runs,omitempty"`
	State     string          `json:"state,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// WorkflowRun represents a workflow run
type WorkflowRun struct {
	ID         string    `json:"id"`
	WorkflowID string    `json:"workflow_id"`
	State      string    `json:"state"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// AlertRoute represents an alert route in incident.io
type AlertRoute struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Enabled      bool                   `json:"enabled"`
	Conditions   []AlertCondition       `json:"conditions"`
	Escalations  []EscalationBinding    `json:"escalations"`
	GroupingKeys []string               `json:"grouping_keys,omitempty"`
	Template     map[string]interface{} `json:"template,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// AlertCondition represents a condition for alert routing
type AlertCondition struct {
	Field     string `json:"field"`
	Operation string `json:"operation"`
	Value     string `json:"value"`
}

// EscalationBinding represents an escalation in alert routing
type EscalationBinding struct {
	ID    string `json:"id"`
	Level int    `json:"level"`
}

// AlertEvent represents an alert event
type AlertEvent struct {
	ID               string                 `json:"id"`
	AlertSourceID    string                 `json:"alert_source_id"`
	DeduplicationKey string                 `json:"deduplication_key,omitempty"`
	Status           string                 `json:"status"`
	Title            string                 `json:"title"`
	Description      string                 `json:"description,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// RetrospectiveIncidentOptionsRequest represents retrospective options for an incident
type RetrospectiveIncidentOptionsRequest struct {
	ExternalID            int64  `json:"external_id,omitempty"`
	PostmortemDocumentURL string `json:"postmortem_document_url,omitempty"`
	SlackChannelID        string `json:"slack_channel_id,omitempty"`
}

// CreateIncidentRequest represents a request to create an incident
type CreateIncidentRequest struct {
	IdempotencyKey               string                               `json:"idempotency_key"`
	Name                         string                               `json:"name"`
	Summary                      string                               `json:"summary,omitempty"`
	IncidentStatusID             string                               `json:"incident_status_id,omitempty"`
	SeverityID                   string                               `json:"severity_id,omitempty"`
	IncidentTypeID               string                               `json:"incident_type_id,omitempty"`
	Mode                         string                               `json:"mode,omitempty"`
	Visibility                   string                               `json:"visibility,omitempty"`
	CustomFieldEntries           []CustomFieldEntryRequest            `json:"custom_field_entries,omitempty"`
	IncidentRoleAssignments      []CreateRoleAssignmentRequest        `json:"incident_role_assignments,omitempty"`
	IncidentTimestampValues      []IncidentTimestampValueRequest      `json:"incident_timestamp_values,omitempty"`
	SlackChannelNameOverride     string                               `json:"slack_channel_name_override,omitempty"`
	SlackTeamID                  string                               `json:"slack_team_id,omitempty"`
	RetrospectiveIncidentOptions *RetrospectiveIncidentOptionsRequest `json:"retrospective_incident_options,omitempty"`
}

// CustomFieldEntryRequest represents a custom field entry in create/update requests
type CustomFieldEntryRequest struct {
	CustomFieldID string        `json:"custom_field_id"`
	Values        []interface{} `json:"values"`
}

// CreateRoleAssignmentRequest represents a role assignment in create request
type CreateRoleAssignmentRequest struct {
	IncidentRoleID string `json:"incident_role_id"`
	UserID         string `json:"user_id"`
}

// IncidentTimestampValueRequest represents a timestamp value update request
type IncidentTimestampValueRequest struct {
	IncidentTimestampID string `json:"incident_timestamp_id"`
	Value               string `json:"value"`
}

// UpdateIncidentRequest represents a request to update an incident
type UpdateIncidentRequest struct {
	Name                     string                          `json:"name,omitempty"`
	Summary                  string                          `json:"summary,omitempty"`
	IncidentStatusID         string                          `json:"incident_status_id,omitempty"`
	SeverityID               string                          `json:"severity_id,omitempty"`
	CallURL                  string                          `json:"call_url,omitempty"`
	SlackChannelNameOverride string                          `json:"slack_channel_name_override,omitempty"`
	CustomFieldEntries       []CustomFieldEntryRequest       `json:"custom_field_entries,omitempty"`
	IncidentRoleAssignments  []CreateRoleAssignmentRequest   `json:"incident_role_assignments,omitempty"`
	IncidentTimestampValues  []IncidentTimestampValueRequest `json:"incident_timestamp_values,omitempty"`
}

// IncidentUpdate represents a status update posted to an incident
type IncidentUpdate struct {
	ID         string    `json:"id"`
	IncidentID string    `json:"incident_id"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Author     *User     `json:"author,omitempty"`
}

// CreateIncidentUpdateRequest represents a request to create an incident update
type CreateIncidentUpdateRequest struct {
	IncidentID string `json:"incident_id"`
	Message    string `json:"message"`
}

// ListIncidentUpdatesOptions represents options for listing incident updates
type ListIncidentUpdatesOptions struct {
	IncidentID string
	PageSize   int
	After      string
}

// ListIncidentUpdatesResponse represents the response from listing incident updates
type ListIncidentUpdatesResponse struct {
	IncidentUpdates []IncidentUpdate `json:"incident_updates"`
	ListResponse
}

// ListResponse represents a paginated list response
type ListResponse struct {
	PaginationMeta struct {
		After            string `json:"after,omitempty"`
		PageSize         int    `json:"page_size"`
		TotalRecordCount int    `json:"total_record_count"`
	} `json:"pagination_meta"`
}

// CatalogType represents a catalog type in incident.io
type CatalogType struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	TypeName    string                 `json:"type_name"`
	Color       string                 `json:"color"`
	Icon        string                 `json:"icon"`
	Annotations map[string]interface{} `json:"annotations"`
	Attributes  []CatalogAttribute     `json:"attributes"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// CatalogAttribute represents an attribute of a catalog type
type CatalogAttribute struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// CatalogEntry represents a catalog entry in incident.io
type CatalogEntry struct {
	ID              string                                `json:"id"`
	Name            string                                `json:"name"`
	Aliases         []string                              `json:"aliases"`
	CatalogTypeID   string                                `json:"catalog_type_id"`
	AttributeValues map[string]CatalogEntryAttributeValue `json:"attribute_values"`
	ExternalID      string                                `json:"external_id"`
	Rank            int                                   `json:"rank"`
	CreatedAt       time.Time                             `json:"created_at"`
	UpdatedAt       time.Time                             `json:"updated_at"`
}

// CatalogEntryAttributeValue represents an attribute value in a catalog entry
type CatalogEntryAttributeValue struct {
	ArrayValue []CatalogEntryAttributeValueItem `json:"array_value,omitempty"`
	Value      *CatalogEntryAttributeValueItem  `json:"value,omitempty"`
}

// CatalogEntryAttributeValueItem represents a single attribute value item
type CatalogEntryAttributeValueItem struct {
	Literal string `json:"literal,omitempty"`
	ID      string `json:"id,omitempty"`
}

// ListCatalogTypesResponse represents the response from listing catalog types
type ListCatalogTypesResponse struct {
	CatalogTypes []CatalogType `json:"catalog_types"`
	ListResponse
}

// ListCatalogEntriesResponse represents the response from listing catalog entries
type ListCatalogEntriesResponse struct {
	CatalogEntries []CatalogEntry `json:"catalog_entries"`
	ListResponse
}

// UpdateCatalogEntryRequest represents a request to update a catalog entry
type UpdateCatalogEntryRequest struct {
	Name             string                                `json:"name,omitempty"`
	Aliases          []string                              `json:"aliases,omitempty"`
	AttributeValues  map[string]CatalogEntryAttributeValue `json:"attribute_values,omitempty"`
	ExternalID       string                                `json:"external_id,omitempty"`
	Rank             int                                   `json:"rank,omitempty"`
	UpdateAttributes []string                              `json:"update_attributes,omitempty"`
}

// CustomField represents a custom field in incident.io (V2)
type CustomField struct {
	ID                     string              `json:"id"`
	Name                   string              `json:"name"`
	Description            string              `json:"description"`
	FieldType              string              `json:"field_type"` // e.g., "single_select", "multi_select", "text", "link", "numeric"
	Required               string              `json:"required"`   // "never", "always", "before_closure"
	ShowBeforeClosure      bool                `json:"show_before_closure"`
	ShowBeforeCreation     bool                `json:"show_before_creation"`
	ShowBeforeUpdate       bool                `json:"show_before_update"`
	ShowInAnnouncementPost *bool               `json:"show_in_announcement_post,omitempty"`
	Options                []CustomFieldOption `json:"options,omitempty"`
	CatalogTypeID          string              `json:"catalog_type_id,omitempty"`
	CreatedAt              time.Time           `json:"created_at"`
	UpdatedAt              time.Time           `json:"updated_at"`
}

// CustomFieldOption represents an option for a custom field
type CustomFieldOption struct {
	ID        string    `json:"id"`
	Value     string    `json:"value"`
	SortKey   int       `json:"sort_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateCustomFieldRequest represents a request to create a custom field
type CreateCustomFieldRequest struct {
	Name                   string   `json:"name"`
	Description            string   `json:"description"`
	FieldType              string   `json:"field_type"`
	Required               string   `json:"required"` // "never", "always", "before_closure"
	ShowBeforeClosure      bool     `json:"show_before_closure"`
	ShowBeforeCreation     bool     `json:"show_before_creation"`
	ShowBeforeUpdate       bool     `json:"show_before_update"`
	ShowInAnnouncementPost *bool    `json:"show_in_announcement_post,omitempty"`
	CatalogTypeID          string   `json:"catalog_type_id,omitempty"` // For catalog fields
	Options                []string `json:"options,omitempty"`         // For select fields
}

// UpdateCustomFieldRequest represents a request to update a custom field
type UpdateCustomFieldRequest struct {
	Name                   string   `json:"name,omitempty"`
	Description            string   `json:"description,omitempty"`
	Required               string   `json:"required,omitempty"`
	ShowBeforeClosure      *bool    `json:"show_before_closure,omitempty"`
	ShowBeforeCreation     *bool    `json:"show_before_creation,omitempty"`
	ShowBeforeUpdate       *bool    `json:"show_before_update,omitempty"`
	ShowInAnnouncementPost *bool    `json:"show_in_announcement_post,omitempty"`
	Options                []string `json:"options,omitempty"`
}

// ListCustomFieldsResponse represents the response from listing custom fields
type ListCustomFieldsResponse struct {
	CustomFields []CustomField `json:"custom_fields"`
	ListResponse
}

// CreateCustomFieldOptionRequest represents a request to create a custom field option (V1)
type CreateCustomFieldOptionRequest struct {
	CustomFieldID string `json:"custom_field_id"`
	Value         string `json:"value"`
	SortKey       int    `json:"sort_key,omitempty"`
}

// UpdateCustomFieldOptionRequest represents a request to update a custom field option (V1)
type UpdateCustomFieldOptionRequest struct {
	Value   string `json:"value,omitempty"`
	SortKey int    `json:"sort_key,omitempty"`
}

// FollowUpPriority represents a follow-up priority
type FollowUpPriority struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Rank        int    `json:"rank"`
}

// FollowUpCreator represents the creator of a follow-up
type FollowUpCreator struct {
	User     *User     `json:"user,omitempty"`
	Alert    *Alert    `json:"alert,omitempty"`
	APIKey   *APIKey   `json:"api_key,omitempty"`
	Workflow *Workflow `json:"workflow,omitempty"`
}

// APIKey represents an API key
type APIKey struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ExternalIssueReference represents an external issue reference
type ExternalIssueReference struct {
	Provider       string `json:"provider"`
	IssueName      string `json:"issue_name"`
	IssuePermalink string `json:"issue_permalink"`
}

// FollowUp represents a follow-up in incident.io
type FollowUp struct {
	ID                     string                  `json:"id"`
	IncidentID             string                  `json:"incident_id"`
	Title                  string                  `json:"title"`
	Description            string                  `json:"description"`
	Status                 string                  `json:"status"`
	Assignee               *User                   `json:"assignee,omitempty"`
	Priority               *FollowUpPriority       `json:"priority,omitempty"`
	Creator                *FollowUpCreator        `json:"creator,omitempty"`
	ExternalIssueReference *ExternalIssueReference `json:"external_issue_reference,omitempty"`
	CreatedAt              time.Time               `json:"created_at"`
	UpdatedAt              time.Time               `json:"updated_at"`
	CompletedAt            *time.Time              `json:"completed_at,omitempty"`
}
