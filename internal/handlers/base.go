package handlers

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incident-io/incidentio-mcp-golang/internal/client"
)

// BaseTool provides common functionality for all MCP tools
type BaseTool struct {
	apiClient *client.Client
}

// NewBaseTool creates a new base tool with the given client
func NewBaseTool(c *client.Client) *BaseTool {
	return &BaseTool{apiClient: c}
}

// GetClient returns the API client
func (b *BaseTool) GetClient() *client.Client {
	return b.apiClient
}

// FormatResponse formats a response as compact JSON to minimize context window usage
func (b *BaseTool) FormatResponse(response interface{}) (string, error) {
	result, err := json.Marshal(response) // Compact JSON (no indentation)
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}
	return string(result), nil
}

// ValidateRequiredString validates that a required string parameter exists
func (b *BaseTool) ValidateRequiredString(args map[string]interface{}, paramName string) (string, error) {
	value, ok := args[paramName].(string)
	if !ok || value == "" {
		argDetails := make(map[string]interface{})
		for key, val := range args {
			argDetails[key] = val
		}
		return "", fmt.Errorf("%s parameter is required and must be a non-empty string. Received parameters: %+v", paramName, argDetails)
	}
	return value, nil
}

// ValidateOptionalString validates an optional string parameter
func (b *BaseTool) ValidateOptionalString(args map[string]interface{}, paramName string) string {
	if value, ok := args[paramName].(string); ok && value != "" {
		return value
	}
	return ""
}

// ValidateOptionalInt validates an optional integer parameter
func (b *BaseTool) ValidateOptionalInt(args map[string]interface{}, paramName string, defaultValue int) int {
	if value, ok := args[paramName].(float64); ok {
		return int(value)
	}
	return defaultValue
}

// ValidateOptionalStringArray validates an optional string array parameter
func (b *BaseTool) ValidateOptionalStringArray(args map[string]interface{}, paramName string) []string {
	var result []string
	if values, ok := args[paramName].([]interface{}); ok {
		for _, v := range values {
			if str, ok := v.(string); ok {
				result = append(result, str)
			}
		}
	}
	return result
}

// CreatePaginationResponse creates a standardized pagination response
func (b *BaseTool) CreatePaginationResponse(data interface{}, paginationMeta interface{}, count int) map[string]interface{} {
	response := map[string]interface{}{
		"data":            data,
		"pagination_meta": paginationMeta,
		"count":           count,
	}

	// Add pagination hints if there are more results
	if paginationMeta != nil {
		// Use reflection to check if pagination has an "after" field
		paginationValue := reflect.ValueOf(paginationMeta)
		if paginationValue.Kind() == reflect.Ptr {
			paginationValue = paginationValue.Elem()
		}

		if paginationValue.Kind() == reflect.Struct {
			afterField := paginationValue.FieldByName("After")
			totalRecordCountField := paginationValue.FieldByName("TotalRecordCount")

			// Use total_record_count to determine if there are more results
			// The "after" cursor is only needed for the next API call, not for determining if more results exist
			hasMore := false
			if totalRecordCountField.IsValid() {
				// We have total count, so we can determine if there are more records
				totalRecords := int(totalRecordCountField.Int())
				recordsFetched := count // Use the count parameter passed to the function
				hasMore = recordsFetched < totalRecords

				// Add progress information
				response["pagination_progress"] = map[string]interface{}{
					"records_fetched":  recordsFetched,
					"total_records":    totalRecords,
					"remaining":        totalRecords - recordsFetched,
					"progress_percent": fmt.Sprintf("%.1f%%", float64(recordsFetched)/float64(totalRecords)*100),
				}
			} else if afterField.IsValid() && afterField.String() != "" {
				// Fallback: if no total count available, assume there are more results if after cursor is present
				hasMore = true
			}

			if hasMore {
				response["has_more_results"] = true
				response["next_page_hint"] = fmt.Sprintf("More results available. Use after='%s' to fetch the next page.", afterField.String())
			} else {
				response["has_more_results"] = false
				response["pagination_status"] = "COMPLETE - All results fetched"
			}
		}
	}

	return response
}

// CreateSimpleResponse creates a simple response with data and count
func (b *BaseTool) CreateSimpleResponse(data interface{}, message string) map[string]interface{} {
	response := map[string]interface{}{
		"data": data,
	}

	// Get count from data if it's a slice
	dataValue := reflect.ValueOf(data)
	if dataValue.Kind() == reflect.Slice {
		response["count"] = dataValue.Len()
	}

	if message != "" {
		response["message"] = message
	}

	return response
}

// StandardInputSchema creates a standard input schema with common patterns
func (b *BaseTool) StandardInputSchema(properties map[string]interface{}, required []string) map[string]interface{} {
	schema := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}

	if len(required) > 0 {
		requiredInterface := make([]interface{}, len(required))
		for i, req := range required {
			requiredInterface[i] = req
		}
		schema["required"] = requiredInterface
	}

	return schema
}

// StandardPaginationProperties returns standard pagination input properties
func (b *BaseTool) StandardPaginationProperties() map[string]interface{} {
	return map[string]interface{}{
		"page_size": map[string]interface{}{
			"type":        "integer",
			"description": "Number of results per page (1-100). Default varies by endpoint.",
			"minimum":     1,
			"maximum":     100,
		},
		"after": map[string]interface{}{
			"type":        "string",
			"description": "Pagination cursor from previous response's 'pagination_meta.after' field. Use this to fetch the next page of results.",
		},
	}
}
