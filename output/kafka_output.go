package output

import (
	"encoding/json"

	"github.com/childe/gohangout/codec"
	"github.com/childe/gohangout/topology"
	"github.com/childe/gohangout/value_render"
	"github.com/childe/healer"
	"github.com/golang/glog"
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
		glog.Fatal("kafka output must have producer_settings")
	}
	newPc := make(map[string]interface{})
	for k, v := range pc.(map[interface{}]interface{}) {
		newPc[k.(string)] = v
	}
	producer_settings := make(map[string]interface{})
	if b, err := json.Marshal(newPc); err != nil {
		glog.Fatalf("could not init kafka producer config: %v", err)
	} else {
		json.Unmarshal(b, &producer_settings)
	}

	glog.Info(producer_settings)

	producerConfig, err := healer.GetProducerConfig(producer_settings)
	if err != nil {
		glog.Fatalf("could not init kafka producer config: %v", err)
	}

	var topic string
	if v, ok := config["topic"]; !ok {
		glog.Fatal("kafka output must have topic setting")
	} else {
		topic = v.(string)
	}

	p.producer = healer.NewProducer(topic, producerConfig)
	if p.producer == nil {
		glog.Fatal("could not create kafka producer")
	}

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
		glog.Errorf("marshal %v error: %s", event, err)
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
