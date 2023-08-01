package output

import (
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

	var topic string
	if v, ok := config["topic"]; !ok {
		glog.Fatal("kafka output must have topic setting")
	} else {
		topic = v.(string)
	}

	var err error
	p.producer, err = healer.NewProducer(topic, pc)
	if err != nil {
		glog.Fatalf("could not create kafka producer: %v", err)
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
