package handlers

import "context"

// Handler represents an MCP tool handler
type Handler interface {
	Name() string
	Description() string
	InputSchema() map[string]interface{}
	Execute(ctx context.Context, args map[string]interface{}) (string, error)
}
