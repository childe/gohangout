package output

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
				"topic": "test-topic",
				"producer_settings": map[any]any{
					"bootstrap.servers": "localhost:9092",
					"batch.size":        json.Number("16384"),
				},
			},
			expected: map[string]any{
				"codec": "json",
				"topic": "test-topic",
				"producer_settings": map[string]any{
					"bootstrap.servers": "localhost:9092",
					"batch.size":        json.Number("16384"),
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
			input:    "json",
			expected: "json",
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
		name       string
		outputType string
		config     map[any]any
		expected   StdoutConfig
		panics     bool
	}{
		{
			name:       "valid stdout config",
			outputType: "Stdout",
			config: map[any]any{
				"codec": "json",
			},
			expected: StdoutConfig{
				Codec: "json",
			},
			panics: false,
		},
		{
			name:       "empty config with defaults",
			outputType: "Stdout",
			config:     map[any]any{},
			expected: StdoutConfig{
				Codec: "",
			},
			panics: false,
		},
		{
			name:       "invalid JSON structure should panic",
			outputType: "Stdout",
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
				var result StdoutConfig
				SafeDecodeConfig(tt.outputType, tt.config, &result)
			} else {
				var result StdoutConfig
				SafeDecodeConfig(tt.outputType, tt.config, &result)
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
		outputType string
		fields     map[string]any
		panics     bool
	}{
		{
			name:       "all fields present",
			outputType: "Kafka",
			fields: map[string]any{
				"topic": "test-topic",
				"producer_settings": map[string]any{
					"bootstrap.servers": "localhost:9092",
				},
			},
			panics: false,
		},
		{
			name:       "nil field should panic",
			outputType: "Kafka",
			fields: map[string]any{
				"topic": nil,
			},
			panics: true,
		},
		{
			name:       "empty string should panic",
			outputType: "Kafka",
			fields: map[string]any{
				"topic": "",
			},
			panics: true,
		},
		{
			name:       "empty map should pass",
			outputType: "Kafka",
			fields: map[string]any{
				"producer_settings": map[string]any{},
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
						// Verify panic message contains output type and field name
						panicMsg := r.(string)
						if !strings.Contains(panicMsg, tt.outputType) {
							t.Errorf("Panic message should contain output type: %s", panicMsg)
						}
					}
				}()
				ValidateRequiredFields(tt.outputType, tt.fields)
			} else {
				ValidateRequiredFields(tt.outputType, tt.fields)
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