package filter

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestConvertToJSONCompatible(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected any
		panics   bool
	}{
		{
			name: "map[any]any to map[string]any",
			input: map[any]any{
				"key1": "value1",
				"key2": 123,
				"key3": map[any]any{
					"nested": "value",
				},
			},
			expected: map[string]any{
				"key1": "value1",
				"key2": 123,
				"key3": map[string]any{
					"nested": "value",
				},
			},
			panics: false,
		},
		{
			name: "slice conversion",
			input: []any{
				"string",
				123,
				map[any]any{"key": "value"},
			},
			expected: []any{
				"string",
				123,
				map[string]any{"key": "value"},
			},
			panics: false,
		},
		{
			name: "non-string key should panic",
			input: map[any]any{
				123: "value",
			},
			panics: true,
		},
		{
			name:     "primitive value unchanged",
			input:    "string",
			expected: "string",
			panics:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.panics {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("ConvertToJSONCompatible() should have panicked")
					}
				}()
				ConvertToJSONCompatible(tt.input)
			} else {
				result := ConvertToJSONCompatible(tt.input)
				if !equalAny(result, tt.expected) {
					t.Errorf("ConvertToJSONCompatible() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestSafeDecodeConfig(t *testing.T) {
	type TestConfig struct {
		Field1 string `json:"field1"`
		Field2 int    `json:"field2"`
		Field3 bool   `json:"field3"`
	}

	tests := []struct {
		name       string
		filterType string
		config     map[any]any
		expected   TestConfig
		panics     bool
	}{
		{
			name:       "valid config",
			filterType: "Test",
			config: map[any]any{
				"field1": "value1",
				"field2": json.Number("123"),
				"field3": true,
			},
			expected: TestConfig{
				Field1: "value1",
				Field2: 123,
				Field3: true,
			},
			panics: false,
		},
		{
			name:       "partial config with defaults",
			filterType: "Test",
			config: map[any]any{
				"field1": "value1",
			},
			expected: TestConfig{
				Field1: "value1",
				Field2: 0,
				Field3: false,
			},
			panics: false,
		},
		{
			name:       "invalid JSON structure should panic",
			filterType: "Test",
			config: map[any]any{
				123: "invalid key",
			},
			panics: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.panics {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("SafeDecodeConfig() should have panicked")
					}
				}()
				var result TestConfig
				SafeDecodeConfig(tt.filterType, tt.config, &result)
			} else {
				var result TestConfig
				SafeDecodeConfig(tt.filterType, tt.config, &result)
				if result != tt.expected {
					t.Errorf("SafeDecodeConfig() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestValidateRequiredFields(t *testing.T) {
	tests := []struct {
		name       string
		filterType string
		fields     map[string]any
		panics     bool
	}{
		{
			name:       "all fields present",
			filterType: "Test",
			fields: map[string]any{
				"field1": "value1",
				"field2": 123,
			},
			panics: false,
		},
		{
			name:       "nil field should panic",
			filterType: "Test",
			fields: map[string]any{
				"field1": nil,
			},
			panics: true,
		},
		{
			name:       "empty string should panic",
			filterType: "Test",
			fields: map[string]any{
				"field1": "",
			},
			panics: true,
		},
		{
			name:       "zero value should pass",
			filterType: "Test",
			fields: map[string]any{
				"field1": 0,
			},
			panics: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.panics {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("ValidateRequiredFields() should have panicked")
					} else {
						// Verify panic message contains filter type and field name
						panicMsg := r.(string)
						if !strings.Contains(panicMsg, tt.filterType) {
							t.Errorf("Panic message should contain filter type: %s", panicMsg)
						}
					}
				}()
				ValidateRequiredFields(tt.filterType, tt.fields)
			} else {
				ValidateRequiredFields(tt.filterType, tt.fields)
			}
		})
	}
}

func TestConvertJSONNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected any
	}{
		{
			name:     "json.Number to int",
			input:    json.Number("123"),
			expected: 123,
		},
		{
			name:     "json.Number to float",
			input:    json.Number("123.45"),
			expected: 123.45,
		},
		{
			name:     "invalid json.Number returns as string",
			input:    json.Number("invalid"),
			expected: "invalid",
		},
		{
			name:     "non-json.Number returns original",
			input:    "string",
			expected: "string",
		},
		{
			name:     "nil returns nil",
			input:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertJSONNumber(tt.input)
			if !equalAny(result, tt.expected) {
				t.Errorf("convertJSONNumber() = %v (%T), want %v (%T)", result, result, tt.expected, tt.expected)
			}
		})
	}
}

// Helper function to compare any values deeply
func equalAny(a, b any) bool {
	switch va := a.(type) {
	case map[string]any:
		vb, ok := b.(map[string]any)
		if !ok || len(va) != len(vb) {
			return false
		}
		for k, v := range va {
			if !equalAny(v, vb[k]) {
				return false
			}
		}
		return true
	case []any:
		vb, ok := b.([]any)
		if !ok || len(va) != len(vb) {
			return false
		}
		for i, v := range va {
			if !equalAny(v, vb[i]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}