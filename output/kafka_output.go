package output

import (
	"encoding/json"

	"github.com/childe/gohangout/codec"
	"github.com/childe/gohangout/topology"
	"github.com/childe/gohangout/value_render"
	"github.com/childe/healer"
	"k8s.io/klog/v2"
)

func init() {
	Register("Kafka", newKafkaOutput)
}

type KafkaOutput struct {
	config map[interface{}]interface{}

	encoder codec.Encoder

	producer *healer.Producer
	key      value_render.ValueRender
}

func newKafkaOutput(config map[interface{}]interface{}) topology.Output {
	p := &KafkaOutput{
		config: config,
	}

	if v, ok := config["codec"]; ok {
		p.encoder = codec.NewEncoder(v.(string))
	} else {
		p.encoder = codec.NewEncoder("json")
	}

	pc, ok := config["producer_settings"]
	if !ok {
		klog.Fatal("kafka output must have producer_settings")
	}
	newPc := make(map[string]interface{})
	for k, v := range pc.(map[interface{}]interface{}) {
		newPc[k.(string)] = v
	}
	producer_settings := make(map[string]interface{})
	if b, err := json.Marshal(newPc); err != nil {
		klog.Fatalf("could not init kafka producer config: %v", err)
	} else {
		json.Unmarshal(b, &producer_settings)
	}

	klog.Info(producer_settings)

	var topic string
	if v, ok := config["topic"]; !ok {
		klog.Fatal("kafka output must have topic setting")
	} else {
		topic = v.(string)
	}

	producer, err := healer.NewProducer(topic, producer_settings)
	if err != nil {
		klog.Fatalf("could not create kafka producer: %v", err)
	}
	p.producer = producer

	if v, ok := config["key"]; ok {
		p.key = value_render.GetValueRender(v.(string))
	} else {
		p.key = nil
	}

	return p
}

func (p *KafkaOutput) Emit(event map[string]interface{}) {
	buf, err := p.encoder.Encode(event)
	if err != nil {
		klog.Errorf("marshal %v error: %s", event, err)
		return
	}
	if p.key == nil {
		p.producer.AddMessage(nil, buf)
	} else {
		key := []byte(p.key.Render(event).(string))
		p.producer.AddMessage(key, buf)
	}
}

func (p *KafkaOutput) Shutdown() {
	p.producer.Close()
}
