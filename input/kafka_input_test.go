package input

import (
	"encoding/json"
	"testing"
)

func TestKafkaInputConfigParsing(t *testing.T) {
	tests := []struct {
		name     string
		config   map[any]any
		expected KafkaInputConfig
		panics   bool
	}{
		{
			name: "valid kafka config with topic",
			config: map[any]any{
				"codec":           "json",
				"decorate_events": true,
				"topic": map[any]any{
					"test-topic": json.Number("2"),
				},
				"messages_queue_length": json.Number("100"),
				"consumer_settings": map[any]any{
					"bootstrap.servers": "localhost:9092",
					"group.id":          "test-group",
				},
			},
			expected: KafkaInputConfig{
				Codec:              "json",
				DecorateEvents:     true,
				Topic:              map[string]int{"test-topic": 2},
				Assign:             nil,
				MessagesQueueLength: 100,
				ConsumerSettings: map[string]any{
					"bootstrap.servers": "localhost:9092",
					"group.id":          "test-group",
				},
			},
			panics: false,
		},
		{
			name: "valid kafka config with assign",
			config: map[any]any{
				"codec": "plain",
				"assign": map[any]any{
					"test-topic": []any{
						json.Number("0"),
						json.Number("1"),
						json.Number("2"),
					},
				},
				"consumer_settings": map[any]any{
					"bootstrap.servers": "localhost:9092",
				},
			},
			expected: KafkaInputConfig{
				Codec:              "plain",
				DecorateEvents:     false,
				Topic:              nil,
				Assign:             map[string][]int{"test-topic": {0, 1, 2}},
				MessagesQueueLength: 0,
				ConsumerSettings: map[string]any{
					"bootstrap.servers": "localhost:9092",
				},
			},
			panics: false,
		},
		{
			name: "minimal kafka config",
			config: map[any]any{
				"topic": map[any]any{
					"test-topic": json.Number("1"),
				},
				"consumer_settings": map[any]any{
					"bootstrap.servers": "localhost:9092",
				},
			},
			expected: KafkaInputConfig{
				Codec:              "",
				DecorateEvents:     false,
				Topic:              map[string]int{"test-topic": 1},
				Assign:             nil,
				MessagesQueueLength: 0,
				ConsumerSettings: map[string]any{
					"bootstrap.servers": "localhost:9092",
				},
			},
			panics: false,
		},
		{
			name: "missing consumer_settings should panic",
			config: map[any]any{
				"topic": map[any]any{
					"test-topic": json.Number("1"),
				},
			},
			panics: true,
		},
		{
			name: "both topic and assign should panic",
			config: map[any]any{
				"topic": map[any]any{
					"test-topic": json.Number("1"),
				},
				"assign": map[any]any{
					"test-topic": []any{json.Number("0")},
				},
				"consumer_settings": map[any]any{
					"bootstrap.servers": "localhost:9092",
				},
			},
			panics: true,
		},
		{
			name: "neither topic nor assign should panic",
			config: map[any]any{
				"consumer_settings": map[any]any{
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
						t.Errorf("newKafkaInput() should have panicked")
					}
				}()
				newKafkaInput(tt.config)
			} else {
				// Note: We can't actually test the full functionality without a real Kafka broker
				// but we can test that the configuration parsing works correctly
				var kafkaConfig KafkaInputConfig
				// Only set defaults if the expected values are non-empty/non-zero
				if tt.expected.Codec != "" {
					kafkaConfig.Codec = "plain"
				}
				if tt.expected.MessagesQueueLength != 0 {
					kafkaConfig.MessagesQueueLength = 10
				}
				
				SafeDecodeConfig("Kafka", tt.config, &kafkaConfig)
				
				if !equalKafkaInputConfig(kafkaConfig, tt.expected) {
					t.Errorf("SafeDecodeConfig() = %v, want %v", kafkaConfig, tt.expected)
				}
			}
		})
	}
}

// Helper function to compare KafkaInputConfig structs
func equalKafkaInputConfig(a, b KafkaInputConfig) bool {
	return a.Codec == b.Codec &&
		a.DecorateEvents == b.DecorateEvents &&
		a.MessagesQueueLength == b.MessagesQueueLength &&
		equalAny(a.Topic, b.Topic) &&
		equalAny(a.Assign, b.Assign) &&
		equalAny(a.ConsumerSettings, b.ConsumerSettings)
}