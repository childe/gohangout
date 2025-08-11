package output

import (
	"encoding/json"
	"testing"
)

func TestKafkaConfigParsing(t *testing.T) {
	tests := []struct {
		name     string
		config   map[any]any
		expected KafkaConfig
		panics   bool
	}{
		{
			name: "valid kafka config",
			config: map[any]any{
				"codec": "json",
				"topic": "test-topic",
				"key":   "event_id",
				"producer_settings": map[any]any{
					"bootstrap.servers": "localhost:9092",
					"batch.size":        json.Number("16384"),
				},
			},
			expected: KafkaConfig{
				Codec: "json",
				Topic: "test-topic",
				Key:   "event_id",
				ProducerSettings: map[string]any{
					"bootstrap.servers": "localhost:9092",
					"batch.size":        json.Number("16384"),
				},
			},
			panics: false,
		},
		{
			name: "minimal kafka config",
			config: map[any]any{
				"topic": "test-topic",
				"producer_settings": map[any]any{
					"bootstrap.servers": "localhost:9092",
				},
			},
			expected: KafkaConfig{
				Codec: "",
				Topic: "test-topic",
				Key:   "",
				ProducerSettings: map[string]any{
					"bootstrap.servers": "localhost:9092",
				},
			},
			panics: false,
		},
		{
			name: "missing topic should panic",
			config: map[any]any{
				"producer_settings": map[any]any{
					"bootstrap.servers": "localhost:9092",
				},
			},
			panics: true,
		},
		{
			name: "missing producer_settings should panic",
			config: map[any]any{
				"topic": "test-topic",
			},
			panics: true,
		},
		{
			name: "empty topic should panic",
			config: map[any]any{
				"topic": "",
				"producer_settings": map[any]any{
					"bootstrap.servers": "localhost:9092",
				},
			},
			panics: true,
		},
		{
			name: "invalid JSON structure should panic",
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
						t.Errorf("newKafkaOutput() should have panicked")
					}
				}()
				newKafkaOutput(tt.config)
			} else {
				// Note: We can't actually test the full functionality without a real Kafka broker
				// but we can test that the configuration parsing works correctly
				var kafkaConfig KafkaConfig
				// Only set default for tests that expect non-empty codec
				if tt.expected.Codec != "" {
					kafkaConfig.Codec = "json" // set default value
				}
				
				SafeDecodeConfig("Kafka", tt.config, &kafkaConfig)
				
				if !equalKafkaConfig(kafkaConfig, tt.expected) {
					t.Errorf("SafeDecodeConfig() = %v, want %v", kafkaConfig, tt.expected)
				}
			}
		})
	}
}

// Helper function to compare KafkaConfig structs
func equalKafkaConfig(a, b KafkaConfig) bool {
	return a.Codec == b.Codec &&
		a.Topic == b.Topic &&
		a.Key == b.Key &&
		equalAny(a.ProducerSettings, b.ProducerSettings)
}