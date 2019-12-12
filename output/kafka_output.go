package output

import (
	"github.com/childe/gohangout/codec"
	"github.com/childe/gohangout/value_render"
	"github.com/childe/healer"
	"github.com/golang/glog"
)

type KafkaOutput struct {
	config map[interface{}]interface{}

	encoder codec.Encoder

	producer *healer.Producer
	key      value_render.ValueRender
}

func (l *MethodLibrary) NewKafkaOutput(config map[interface{}]interface{}) *KafkaOutput {
	p := &KafkaOutput{
		config: config,
	}

	if v, ok := config["codec"]; ok {
		p.encoder = codec.NewEncoder(v.(string))
	} else {
		p.encoder = codec.NewEncoder("json")
	}

	producerConfig := healer.DefaultProducerConfig()

	var topic string
	if v, ok := config["topic"]; !ok {
		glog.Fatal("kafka output must have topic setting")
	} else {
		topic = v.(string)
	}

	if v, ok := config["bootstrap.servers"]; !ok {
		glog.Fatal("kafka output must have bootstrap.servers setting")
	} else {
		producerConfig.BootstrapServers = v.(string)
	}

	if v, ok := config["compression.type"]; ok {
		producerConfig.CompressionType = v.(string)
	}
	if v, ok := config["message.max.count"]; ok {
		producerConfig.MessageMaxCount = v.(int)
	}
	if v, ok := config["flush.interval.ms"]; ok {
		producerConfig.FlushIntervalMS = v.(int)
	}
	if v, ok := config["metadata.max.age.ms"]; ok {
		producerConfig.MetadataMaxAgeMS = v.(int)
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
