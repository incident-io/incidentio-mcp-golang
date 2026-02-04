package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

const (
	// DefaultMaxResponseSize is the default maximum response size in bytes (50KB)
	DefaultMaxResponseSize = 50 * 1024
)

// FormatJSONResponse formats response as compact JSON with size limits to minimize context window usage
func FormatJSONResponse(response interface{}) (string, error) {
	result, err := json.Marshal(response) // Compact JSON (no indentation)
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	// Check response size and truncate if needed
	maxSize := getMaxResponseSize()
	if len(result) > maxSize {
		return truncateResponse(result, maxSize, response)
	}

	return string(result), nil
}

// getMaxResponseSize returns the configured max response size
func getMaxResponseSize() int {
	if sizeStr := os.Getenv("MCP_MAX_RESPONSE_SIZE"); sizeStr != "" {
		if size, err := strconv.Atoi(sizeStr); err == nil && size > 0 {
			return size
		}
	}
	return DefaultMaxResponseSize
}

// truncateResponse truncates a response that exceeds the size limit
func truncateResponse(jsonBytes []byte, maxSize int, originalResponse interface{}) (string, error) {
	// Calculate how much we need to truncate
	truncateAt := maxSize - 500 // Reserve space for warning message
	if truncateAt < 1000 {
		truncateAt = 1000 // Minimum useful size
	}

	// Truncate and add warning
	truncated := string(jsonBytes[:truncateAt])

	// Try to end at a reasonable point (not mid-field)
	if lastComma := findLastComma(truncated); lastComma > 0 {
		truncated = truncated[:lastComma]
	}

	// Create warning message
	warning := map[string]interface{}{
		"_warning":        "Response truncated",
		"_reason":         fmt.Sprintf("Response size (%d bytes) exceeded limit (%d bytes)", len(jsonBytes), maxSize),
		"_original_size":  len(jsonBytes),
		"_truncated_size": len(truncated),
		"_suggestion":     "Use more specific filters or reduce page_size to get smaller responses",
	}

	warningJSON, _ := json.Marshal(warning)

	// Combine truncated response with warning
	// Remove trailing } or ] and add warning
	result := truncated
	if len(result) > 0 {
		lastChar := result[len(result)-1]
		if lastChar == '}' || lastChar == ']' {
			result = result[:len(result)-1]
		}
	}

	return result + "," + string(warningJSON)[1:], nil
}

// findLastComma finds the last comma in a string for clean truncation
func findLastComma(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == ',' {
			return i
		}
	}
	return -1
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
