package output

import (
	"fmt"

	"github.com/childe/gohangout/codec"
	"github.com/childe/gohangout/topology"
	"github.com/childe/gohangout/value_render"
	"github.com/childe/healer"
	"k8s.io/klog/v2"
)

// KafkaConfig defines the configuration structure for Kafka output
type KafkaConfig struct {
	Codec            string         `json:"codec"`
	Topic            string         `json:"topic"`
	Key              string         `json:"key"`
	ProducerSettings map[string]any `json:"producer_settings"`
}

func init() {
	Register("Kafka", newKafkaOutput)
}

type KafkaOutput struct {
	config map[any]any

	encoder codec.Encoder

	producer *healer.Producer
	key      value_render.ValueRender
}

func newKafkaOutput(config map[any]any) topology.Output {
	// Parse configuration using SafeDecodeConfig helper
	var kafkaConfig KafkaConfig
	kafkaConfig.Codec = "json" // set default value

	SafeDecodeConfig("Kafka", config, &kafkaConfig)

	// Validate required fields
	ValidateRequiredFields("Kafka", map[string]any{
		"topic":             kafkaConfig.Topic,
		"producer_settings": kafkaConfig.ProducerSettings,
	})

	p := &KafkaOutput{
		config:  config,
		encoder: codec.NewEncoder(kafkaConfig.Codec),
	}

	klog.Info(kafkaConfig.ProducerSettings)

	producer, err := healer.NewProducer(kafkaConfig.Topic, kafkaConfig.ProducerSettings)
	if err != nil {
		panic(fmt.Sprintf("could not create kafka producer: %v", err))
	}
	p.producer = producer

	if kafkaConfig.Key != "" {
		p.key = value_render.GetValueRender(kafkaConfig.Key)
	} else {
		p.key = nil
	}

	return p
}

func (p *KafkaOutput) Emit(event map[string]any) {
	buf, err := p.encoder.Encode(event)
	if err != nil {
		klog.Errorf("marshal %v error: %s", event, err)
		return
	}
	if p.key == nil {
		p.producer.AddMessage(nil, buf)
	} else {
		key, _ := p.key.Render(event)
		p.producer.AddMessage([]byte(key.(string)), buf)
	}
}

func (p *KafkaOutput) Shutdown() {
	p.producer.Close()
}
