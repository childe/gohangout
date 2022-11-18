package input

import (
	"time"

	"github.com/childe/gohangout/codec"
	"github.com/childe/gohangout/topology"
	"github.com/childe/healer"
	"github.com/golang/glog"
	jsoniter "github.com/json-iterator/go"
)

type KafkaInput struct {
	config         map[interface{}]interface{}
	decorateEvents bool

	messages chan *healer.FullMessage

	decoder codec.Decoder

	groupConsumers []*healer.GroupConsumer
	consumers      []*healer.Consumer
}

func init() {
	Register("Kafka", newKafkaInput)
}

func newKafkaInput(config map[interface{}]interface{}) topology.Input {
	var (
		codertype      string = "plain"
		decorateEvents        = false
		topics         map[interface{}]interface{}
		assign         map[string][]int
	)

	consumer_settings := make(map[string]interface{})
	if v, ok := config["consumer_settings"]; !ok {
		glog.Fatal("kafka input must have consumer_settings")
	} else {
		// official json marshal: unsupported type: map[interface {}]interface {}
		json := jsoniter.ConfigCompatibleWithStandardLibrary
		if b, err := json.Marshal(v); err != nil {
			glog.Fatalf("marshal consumer settings error: %v", err)
		} else {
			json.Unmarshal(b, &consumer_settings)
		}
	}
	if v, ok := config["topic"]; ok {
		topics = v.(map[interface{}]interface{})
	} else {
		topics = nil
	}
	if v, ok := config["assign"]; ok {
		assign = make(map[string][]int)
		for topicName, partitions := range v.(map[interface{}]interface{}) {
			assign[topicName.(string)] = make([]int, len(partitions.([]interface{})))
			for i, p := range partitions.([]interface{}) {
				assign[topicName.(string)][i] = p.(int)
			}
		}
	} else {
		assign = nil
	}

	if topics == nil && assign == nil {
		glog.Fatal("either topic or assign should be set")
	}
	if topics != nil && assign != nil {
		glog.Fatal("topic and assign can not be both set")
	}

	if codecV, ok := config["codec"]; ok {
		codertype = codecV.(string)
	}

	if decorateEventsV, ok := config["decorate_events"]; ok {
		decorateEvents = decorateEventsV.(bool)
	}

	messagesLength := 10
	if v, ok := config["messages_queue_length"]; ok {
		messagesLength = v.(int)
	}

	kafkaInput := &KafkaInput{
		config:         config,
		decorateEvents: decorateEvents,
		messages:       make(chan *healer.FullMessage, messagesLength),

		decoder: codec.NewDecoder(codertype),
	}

	consumerConfig, err := healer.GetConsumerConfig(consumer_settings)
	if err != nil {
		glog.Fatalf("error in consumer settings: %s", err)
	}

	// GroupConsumer
	if topics != nil {
		for topic, threadCount := range topics {
			for i := 0; i < threadCount.(int); i++ {
				c, err := healer.NewGroupConsumer(topic.(string), consumerConfig)
				if err != nil {
					glog.Fatalf("could not init GroupConsumer: %s", err)
				}
				kafkaInput.groupConsumers = append(kafkaInput.groupConsumers, c)

				go func() {
					_, err = c.Consume(kafkaInput.messages)
					if err != nil {
						glog.Fatalf("try to consumer error: %s", err)
					}
				}()
			}
		}
	} else {
		c, err := healer.NewConsumer(consumerConfig)
		if err != nil {
			glog.Fatalf("could not init SimpleConsumer: %s", err)
		}
		kafkaInput.consumers = append(kafkaInput.consumers, c)

		c.Assign(assign)

		go func() {
			_, err = c.Consume(kafkaInput.messages)
			if err != nil {
				glog.Fatalf("try to consume error: %s", err)
			}
		}()
	}

	return kafkaInput
}

// ReadOneEvent implement method in topology.Input.
// gohangout call this method to get one event and pass it to filter or output
func (p *KafkaInput) ReadOneEvent() map[string]interface{} {
	message, more := <-p.messages
	if !more {
		return nil
	}

	if message.Error != nil {
		glog.Error("kafka message carries error: ", message.Error)
		return nil
	}
	event := p.decoder.Decode(message.Message.Value)
	if p.decorateEvents {
		kafkaMeta := make(map[string]interface{})
		kafkaMeta["topic"] = message.TopicName
		kafkaMeta["partition"] = message.PartitionID
		kafkaMeta["offset"] = message.Message.Offset
		event["@metadata"] = map[string]interface{}{"kafka": kafkaMeta}
	}
	return event
}

// Shutdown implement method in topology.Input. It closes all consumers
func (p *KafkaInput) Shutdown() {
	if len(p.groupConsumers) > 0 {
		for _, c := range p.groupConsumers {
			c.AwaitClose(30 * time.Second)
		}
	}
	if len(p.consumers) > 0 {
		for _, c := range p.consumers {
			c.AwaitClose(30 * time.Second)
		}
	}
}
