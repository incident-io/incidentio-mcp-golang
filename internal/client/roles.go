package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// IncidentRole represents an incident role
type IncidentRole struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Shortform    string `json:"shortform"`
	Description  string `json:"description"`
	Instructions string `json:"instructions"`
	RoleType     string `json:"role_type"`
	Required     bool   `json:"required"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// User represents a user in incident.io (expanded from existing definition)
type UserDetailed struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	SlackUserID string    `json:"slack_user_id,omitempty"`
	Role        string    `json:"role"`
	BaseRole    *BaseRole `json:"base_role,omitempty"`
	CustomRoles []Role    `json:"custom_roles,omitempty"`
}

// BaseRole represents a base role
type BaseRole struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Slug        string `json:"slug"`
}

// Role represents a custom role
type Role struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Slug        string `json:"slug"`
}

// ListIncidentRolesOptions represents options for listing incident roles
type ListIncidentRolesOptions struct {
	PageSize int
	After    string
}

// ListIncidentRolesResponse represents the response from listing incident roles
type ListIncidentRolesResponse struct {
	IncidentRoles []IncidentRole `json:"incident_roles"`
	ListResponse
}

// ListUsersOptions represents options for listing users
type ListUsersOptions struct {
	PageSize int
	After    string
	Email    string // Filter by email
}

// ListUsersResponse represents the response from listing users
type ListUsersResponse struct {
	Users []UserDetailed `json:"users"`
	ListResponse
}

// ListIncidentRoles retrieves a list of incident roles
func (c *Client) ListIncidentRoles(ctx context.Context, opts *ListIncidentRolesOptions) (*ListIncidentRolesResponse, error) {
	// Set default page size
	pageSize := 25
	if opts != nil && opts.PageSize > 0 {
		pageSize = opts.PageSize
	}

	params := url.Values{}
	params.Set("page_size", strconv.Itoa(pageSize)) // Always set (may be required)

	if opts != nil && opts.After != "" {
		params.Set("after", opts.After)
	}

	respBody, err := c.doRequest(ctx, "GET", "/incident_roles", params, nil)
	if err != nil {
		return nil, err
	}

	var response ListIncidentRolesResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// ListUsers retrieves a single page of users
// Pagination is controlled by the caller using PageSize and After parameters
func (c *Client) ListUsers(ctx context.Context, opts *ListUsersOptions) (*ListUsersResponse, error) {
	pageSize := 10 // Conservative default to avoid exceeding MCP client limits
	after := ""

	if opts != nil {
		if opts.PageSize > 0 {
			pageSize = opts.PageSize
		}
		if opts.After != "" {
			after = opts.After
		}
	}

	params := url.Values{}
	params.Set("page_size", strconv.Itoa(pageSize))

	if after != "" {
		params.Set("after", after)
	}

	if opts != nil && opts.Email != "" {
		params.Set("email", opts.Email)
	}

	respBody, err := c.doRequest(ctx, "GET", "/users", params, nil)
	if err != nil {
		return nil, err
	}

	var response ListUsersResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}
