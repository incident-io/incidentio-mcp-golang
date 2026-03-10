package handlers

import (
	"encoding/json"
	"fmt"
)

const (
	maxSearchResults = 50
	maxSearchPages   = 50
)

// FormatJSONResponse is a simple helper to eliminate JSON marshaling duplication
func FormatJSONResponse(response interface{}) (string, error) {
	result, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}
	return string(result), nil
}

// GetStringArg is a simple helper to eliminate string parameter validation duplication
func GetStringArg(args map[string]interface{}, key string) string {
	if value, ok := args[key].(string); ok && value != "" {
		return value
	}
	return ""
}

// GetIntArg is a simple helper to eliminate int parameter validation duplication
func GetIntArg(args map[string]interface{}, key string, defaultValue int) int {
	if value, ok := args[key].(float64); ok {
		return int(value)
	}
	return defaultValue
}

// GetStringArrayArg is a simple helper to eliminate string array parameter validation duplication
func GetStringArrayArg(args map[string]interface{}, key string) []string {
	var result []string
	if values, ok := args[key].([]interface{}); ok {
		for _, v := range values {
			if str, ok := v.(string); ok {
				result = append(result, str)
			}
		}
	}
	return result
}

// CreateSimpleResponse is a simple helper to eliminate response creation duplication
func CreateSimpleResponse(data interface{}, message string) map[string]interface{} {
	response := map[string]interface{}{
		"data": data,
	}

	// Add count if data is a slice (handle different slice types)
	switch v := data.(type) {
	case []interface{}:
		response["count"] = len(v)
	case []string:
		response["count"] = len(v)
	case []int:
		response["count"] = len(v)
	case []float64:
		response["count"] = len(v)
	}

	if message != "" {
		response["message"] = message
	}

	return response
}
