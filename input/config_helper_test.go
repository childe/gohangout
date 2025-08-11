package input

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
				"codec": "json",
				"topic": map[any]any{
					"test-topic": json.Number("2"),
				},
				"consumer_settings": map[any]any{
					"bootstrap.servers": "localhost:9092",
					"group.id":          "test-group",
				},
			},
			expected: map[string]any{
				"codec": "json",
				"topic": map[string]any{
					"test-topic": json.Number("2"),
				},
				"consumer_settings": map[string]any{
					"bootstrap.servers": "localhost:9092",
					"group.id":          "test-group",
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
			input:    "plain",
			expected: "plain",
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
	tests := []struct {
		name      string
		inputType string
		config    map[any]any
		expected  StdinConfig
		panics    bool
	}{
		{
			name:      "valid stdin config",
			inputType: "Stdin",
			config: map[any]any{
				"codec": "json",
			},
			expected: StdinConfig{
				Codec: "json",
			},
			panics: false,
		},
		{
			name:      "empty config with defaults",
			inputType: "Stdin",
			config:    map[any]any{},
			expected: StdinConfig{
				Codec: "",
			},
			panics: false,
		},
		{
			name:      "invalid JSON structure should panic",
			inputType: "Stdin",
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
				var result StdinConfig
				SafeDecodeConfig(tt.inputType, tt.config, &result)
			} else {
				var result StdinConfig
				SafeDecodeConfig(tt.inputType, tt.config, &result)
				if result != tt.expected {
					t.Errorf("SafeDecodeConfig() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestValidateRequiredFields(t *testing.T) {
	tests := []struct {
		name      string
		inputType string
		fields    map[string]any
		panics    bool
	}{
		{
			name:      "all fields present",
			inputType: "Kafka",
			fields: map[string]any{
				"consumer_settings": map[string]any{
					"bootstrap.servers": "localhost:9092",
					"group.id":          "test-group",
				},
			},
			panics: false,
		},
		{
			name:      "nil field should panic",
			inputType: "Kafka",
			fields: map[string]any{
				"consumer_settings": nil,
			},
			panics: true,
		},
		{
			name:      "empty string should panic",
			inputType: "Kafka",
			fields: map[string]any{
				"topic": "",
			},
			panics: true,
		},
		{
			name:      "empty map should pass",
			inputType: "Kafka",
			fields: map[string]any{
				"consumer_settings": map[string]any{},
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
						// Verify panic message contains input type and field name
						panicMsg := r.(string)
						if !strings.Contains(panicMsg, tt.inputType) {
							t.Errorf("Panic message should contain input type: %s", panicMsg)
						}
					}
				}()
				ValidateRequiredFields(tt.inputType, tt.fields)
			} else {
				ValidateRequiredFields(tt.inputType, tt.fields)
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
	case map[string]int:
		vb, ok := b.(map[string]int)
		if !ok || len(va) != len(vb) {
			return false
		}
		for k, v := range va {
			if v != vb[k] {
				return false
			}
		}
		return true
	case map[string][]int:
		vb, ok := b.(map[string][]int)
		if !ok || len(va) != len(vb) {
			return false
		}
		for k, v := range va {
			vbSlice, exists := vb[k]
			if !exists || len(v) != len(vbSlice) {
				return false
			}
			for i, val := range v {
				if val != vbSlice[i] {
					return false
				}
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