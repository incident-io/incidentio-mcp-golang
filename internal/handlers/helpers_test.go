package handlers

import (
	"fmt"
	"os"
	"testing"
)

func TestFormatJSONResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
		hasError bool
	}{
		{
			name:     "simple map",
			input:    map[string]interface{}{"key": "value"},
			expected: `{"key":"value"}`,
			hasError: false,
		},
		{
			name:     "nested map",
			input:    map[string]interface{}{"data": map[string]interface{}{"id": "123", "name": "test"}},
			expected: `{"data":{"id":"123","name":"test"}}`,
			hasError: false,
		},
		{
			name:     "slice",
			input:    []string{"item1", "item2"},
			expected: `["item1","item2"]`,
			hasError: false,
		},
		{
			name:     "nil input",
			input:    nil,
			expected: "null",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatJSONResponse(tt.input)

			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestGetStringArg(t *testing.T) {
	tests := []struct {
		name     string
		args     map[string]interface{}
		key      string
		expected string
	}{
		{
			name:     "valid string",
			args:     map[string]interface{}{"key": "value"},
			key:      "key",
			expected: "value",
		},
		{
			name:     "missing key",
			args:     map[string]interface{}{"other": "value"},
			key:      "key",
			expected: "",
		},
		{
			name:     "empty string",
			args:     map[string]interface{}{"key": ""},
			key:      "key",
			expected: "",
		},
		{
			name:     "wrong type",
			args:     map[string]interface{}{"key": 123},
			key:      "key",
			expected: "",
		},
		{
			name:     "nil value",
			args:     map[string]interface{}{"key": nil},
			key:      "key",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetStringArg(tt.args, tt.key)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGetIntArg(t *testing.T) {
	tests := []struct {
		name         string
		args         map[string]interface{}
		key          string
		defaultValue int
		expected     int
	}{
		{
			name:         "valid int",
			args:         map[string]interface{}{"key": 42.0},
			key:          "key",
			defaultValue: 10,
			expected:     42,
		},
		{
			name:         "missing key",
			args:         map[string]interface{}{"other": 42.0},
			key:          "key",
			defaultValue: 10,
			expected:     10,
		},
		{
			name:         "wrong type",
			args:         map[string]interface{}{"key": "not a number"},
			key:          "key",
			defaultValue: 10,
			expected:     10,
		},
		{
			name:         "zero value",
			args:         map[string]interface{}{"key": 0.0},
			key:          "key",
			defaultValue: 10,
			expected:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetIntArg(tt.args, tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestGetStringArrayArg(t *testing.T) {
	tests := []struct {
		name     string
		args     map[string]interface{}
		key      string
		expected []string
	}{
		{
			name:     "valid string array",
			args:     map[string]interface{}{"key": []interface{}{"item1", "item2"}},
			key:      "key",
			expected: []string{"item1", "item2"},
		},
		{
			name:     "empty array",
			args:     map[string]interface{}{"key": []interface{}{}},
			key:      "key",
			expected: []string{},
		},
		{
			name:     "missing key",
			args:     map[string]interface{}{"other": []interface{}{"item1"}},
			key:      "key",
			expected: []string{},
		},
		{
			name:     "wrong type",
			args:     map[string]interface{}{"key": "not an array"},
			key:      "key",
			expected: []string{},
		},
		{
			name:     "mixed types in array",
			args:     map[string]interface{}{"key": []interface{}{"item1", 123, "item2"}},
			key:      "key",
			expected: []string{"item1", "item2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetStringArrayArg(tt.args, tt.key)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(result))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("Expected %q at index %d, got %q", tt.expected[i], i, v)
				}
			}
		})
	}
}

func TestCreateSimpleResponse(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		message  string
		expected map[string]interface{}
	}{
		{
			name:    "with message",
			data:    []string{"item1", "item2"},
			message: "Success",
			expected: map[string]interface{}{
				"data":    []string{"item1", "item2"},
				"count":   2,
				"message": "Success",
			},
		},
		{
			name:    "without message",
			data:    []string{"item1"},
			message: "",
			expected: map[string]interface{}{
				"data":  []string{"item1"},
				"count": 1,
			},
		},
		{
			name:    "non-slice data",
			data:    map[string]string{"key": "value"},
			message: "Test",
			expected: map[string]interface{}{
				"data":    map[string]string{"key": "value"},
				"message": "Test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateSimpleResponse(tt.data, tt.message)

			// Check data (handle slices specially)
			if expectedSlice, ok := tt.expected["data"].([]string); ok {
				if resultSlice, ok := result["data"].([]string); ok {
					if len(expectedSlice) != len(resultSlice) {
						t.Errorf("Expected slice length %d, got %d", len(expectedSlice), len(resultSlice))
					} else {
						for i, v := range expectedSlice {
							if v != resultSlice[i] {
								t.Errorf("Expected slice[%d] %q, got %q", i, v, resultSlice[i])
							}
						}
					}
				} else {
					t.Errorf("Expected slice data, got %T", result["data"])
				}
			} else if expectedMap, ok := tt.expected["data"].(map[string]string); ok {
				if resultMap, ok := result["data"].(map[string]string); ok {
					if len(expectedMap) != len(resultMap) {
						t.Errorf("Expected map length %d, got %d", len(expectedMap), len(resultMap))
					} else {
						for k, v := range expectedMap {
							if resultMap[k] != v {
								t.Errorf("Expected map[%q] %q, got %q", k, v, resultMap[k])
							}
						}
					}
				} else {
					t.Errorf("Expected map data, got %T", result["data"])
				}
			} else {
				if result["data"] != tt.expected["data"] {
					t.Errorf("Expected data %v, got %v", tt.expected["data"], result["data"])
				}
			}

			// Check count (if expected)
			if expectedCount, ok := tt.expected["count"]; ok {
				if result["count"] != expectedCount {
					t.Errorf("Expected count %v, got %v", expectedCount, result["count"])
				}
			}

			// Check message (if expected)
			if expectedMessage, ok := tt.expected["message"]; ok {
				if result["message"] != expectedMessage {
					t.Errorf("Expected message %v, got %v", expectedMessage, result["message"])
				}
			}
		})
	}
}

func TestGetMaxResponseSize(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected int
	}{
		{
			name:     "default when no env var",
			envValue: "",
			expected: DefaultMaxResponseSize,
		},
		{
			name:     "valid env var",
			envValue: "100000",
			expected: 100000,
		},
		{
			name:     "invalid env var (non-numeric)",
			envValue: "invalid",
			expected: DefaultMaxResponseSize,
		},
		{
			name:     "invalid env var (negative)",
			envValue: "-1000",
			expected: DefaultMaxResponseSize,
		},
		{
			name:     "invalid env var (zero)",
			envValue: "0",
			expected: DefaultMaxResponseSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envValue != "" {
				if err := os.Setenv("MCP_MAX_RESPONSE_SIZE", tt.envValue); err != nil {
					t.Fatalf("Failed to set env var: %v", err)
				}
				defer func() {
					if err := os.Unsetenv("MCP_MAX_RESPONSE_SIZE"); err != nil {
						t.Errorf("Failed to unset env var: %v", err)
					}
				}()
			} else {
				if err := os.Unsetenv("MCP_MAX_RESPONSE_SIZE"); err != nil {
					t.Fatalf("Failed to unset env var: %v", err)
				}
			}

			result := getMaxResponseSize()
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestTruncateResponse(t *testing.T) {
	// Create a large JSON response for testing
	largeJSON := `{"data":[`
	for i := 0; i < 100; i++ {
		if i > 0 {
			largeJSON += ","
		}
		largeJSON += fmt.Sprintf(`{"id":"%d","name":"item%d","description":"This is a long description for item %d"}`, i, i, i)
	}
	largeJSON += `],"count":100}`

	tests := []struct {
		name         string
		jsonBytes    []byte
		maxSize      int
		shouldError  bool
		checkWarning bool
	}{
		{
			name:         "truncate large response",
			jsonBytes:    []byte(largeJSON),
			maxSize:      2000,
			shouldError:  false,
			checkWarning: true,
		},
		{
			name:         "truncate with comma",
			jsonBytes:    []byte(largeJSON),
			maxSize:      3000,
			shouldError:  false,
			checkWarning: true,
		},
		{
			name:         "very small max size",
			jsonBytes:    []byte(largeJSON),
			maxSize:      1500,
			shouldError:  false,
			checkWarning: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := truncateResponse(tt.jsonBytes, tt.maxSize, map[string]interface{}{})

			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
				return
			}

			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.checkWarning {
				if !contains(result, "_warning") {
					t.Error("Expected warning in truncated response")
				}
				if !contains(result, "_reason") {
					t.Error("Expected reason in truncated response")
				}
				if !contains(result, "_original_size") {
					t.Error("Expected original_size in truncated response")
				}
			}

			// Verify result is smaller than original
			if len(result) > len(tt.jsonBytes) {
				t.Errorf("Truncated result (%d bytes) is larger than original (%d bytes)", len(result), len(tt.jsonBytes))
			}
		})
	}
}

func TestFindLastComma(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "string with comma",
			input:    `{"key":"value","another":"test"}`,
			expected: 14, // Position of comma after "value"
		},
		{
			name:     "string without comma",
			input:    `{"key":"value"}`,
			expected: -1,
		},
		{
			name:     "empty string",
			input:    "",
			expected: -1,
		},
		{
			name:     "multiple commas",
			input:    `{"a":"b","c":"d","e":"f"}`,
			expected: 16, // Position of last comma
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findLastComma(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestFormatJSONResponse_WithTruncation(t *testing.T) {
	// Create a large response that will trigger truncation
	largeData := make([]map[string]string, 1000)
	for i := 0; i < 1000; i++ {
		largeData[i] = map[string]string{
			"id":   fmt.Sprintf("id-%d", i),
			"name": fmt.Sprintf("name-%d", i),
			"desc": "This is a long description that adds to the response size",
		}
	}

	response := map[string]interface{}{
		"data":  largeData,
		"count": len(largeData),
	}

	// Set a small max size to trigger truncation
	if err := os.Setenv("MCP_MAX_RESPONSE_SIZE", "5000"); err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("MCP_MAX_RESPONSE_SIZE"); err != nil {
			t.Errorf("Failed to unset env var: %v", err)
		}
	}()

	result, err := FormatJSONResponse(response)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify truncation occurred
	if !contains(result, "_warning") {
		t.Error("Expected warning in truncated response")
	}

	// Verify size is within limit
	if len(result) > 5000 {
		t.Errorf("Response size (%d) exceeds limit (5000)", len(result))
	}
}

func TestCreateSimpleResponse_DifferentSliceTypes(t *testing.T) {
	tests := []struct {
		name          string
		data          interface{}
		message       string
		expectedCount int
		hasCount      bool
	}{
		{
			name:          "interface slice",
			data:          []interface{}{"a", "b", "c"},
			message:       "Test",
			expectedCount: 3,
			hasCount:      true,
		},
		{
			name:          "int slice",
			data:          []int{1, 2, 3, 4},
			message:       "",
			expectedCount: 4,
			hasCount:      true,
		},
		{
			name:          "float64 slice",
			data:          []float64{1.1, 2.2, 3.3},
			message:       "Numbers",
			expectedCount: 3,
			hasCount:      true,
		},
		{
			name:     "non-slice data",
			data:     map[string]string{"key": "value"},
			message:  "Map",
			hasCount: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateSimpleResponse(tt.data, tt.message)

			// Check data
			if result["data"] == nil {
				t.Error("Expected data field")
			}

			// Check count
			if tt.hasCount {
				count, ok := result["count"].(int)
				if !ok {
					t.Error("Expected count field to be int")
				} else if count != tt.expectedCount {
					t.Errorf("Expected count %d, got %d", tt.expectedCount, count)
				}
			} else {
				if _, exists := result["count"]; exists {
					t.Error("Did not expect count field for non-slice data")
				}
			}

			// Check message
			if tt.message != "" {
				if result["message"] != tt.message {
					t.Errorf("Expected message %q, got %q", tt.message, result["message"])
				}
			} else {
				if _, exists := result["message"]; exists {
					t.Error("Did not expect message field when message is empty")
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
