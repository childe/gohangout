package input

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/childe/gohangout/codec"
	"github.com/childe/gohangout/topology"
	"github.com/childe/healer"
	"k8s.io/klog/v2"
)

// KafkaInputConfig defines the configuration structure for Kafka input
type KafkaInputConfig struct {
	Codec              string            `json:"codec"`
	DecorateEvents     bool              `json:"decorate_events"`
	Topic              map[string]int    `json:"topic"`
	Assign             map[string][]int  `json:"assign"`
	MessagesQueueLength int              `json:"messages_queue_length"`
	ConsumerSettings   map[string]any    `json:"consumer_settings"`
}

type KafkaInput struct {
	config         map[any]any
	decorateEvents bool

	messages chan *healer.FullMessage

	decoder codec.Decoder

	groupConsumers []*healer.GroupConsumer
	consumers      []*healer.Consumer
}

func init() {
	Register("Kafka", newKafkaInput)
}

// convertJSONNumber converts json.Number to appropriate types
func convertJSONNumber(val any) any {
	if num, ok := val.(json.Number); ok {
		if intVal, err := strconv.Atoi(string(num)); err == nil {
			return intVal
		}
		if floatVal, err := strconv.ParseFloat(string(num), 64); err == nil {
			return floatVal
		}
	}
	return val
}

func newKafkaInput(config map[any]any) topology.Input {
	// Parse configuration using SafeDecodeConfig helper
	var kafkaConfig KafkaInputConfig
	kafkaConfig.Codec = "plain" // set default value
	kafkaConfig.MessagesQueueLength = 10 // set default value

	SafeDecodeConfig("Kafka", config, &kafkaConfig)

	// Validate required fields
	ValidateRequiredFields("Kafka", map[string]any{
		"consumer_settings": kafkaConfig.ConsumerSettings,
	})

	if kafkaConfig.Topic == nil && kafkaConfig.Assign == nil {
		panic("kafka input: either topic or assign should be set")
	}
	if kafkaConfig.Topic != nil && kafkaConfig.Assign != nil {
		panic("kafka input: topic and assign can not be both set")
	}

	// Convert json.Number values in consumer_settings
	for k, v := range kafkaConfig.ConsumerSettings {
		kafkaConfig.ConsumerSettings[k] = convertJSONNumber(v)
	}

	kafkaInput := &KafkaInput{
		config:         config,
		decorateEvents: kafkaConfig.DecorateEvents,
		messages:       make(chan *healer.FullMessage, kafkaConfig.MessagesQueueLength),

		decoder: codec.NewDecoder(kafkaConfig.Codec),
	}

	// GroupConsumer
	if kafkaConfig.Topic != nil {
		for topic, threadCount := range kafkaConfig.Topic {
			for i := 0; i < threadCount; i++ {
				c, err := healer.NewGroupConsumer(topic, kafkaConfig.ConsumerSettings)
				if err != nil {
					panic(fmt.Sprintf("could not create kafka GroupConsumer: %s", err))
				}
				kafkaInput.groupConsumers = append(kafkaInput.groupConsumers, c)

				go func() {
					_, err = c.Consume(kafkaInput.messages)
					if err != nil {
						panic(fmt.Sprintf("try to consumer error: %s", err))
					}
				}()
			}
		}
	} else {
		c, err := healer.NewConsumer(kafkaConfig.ConsumerSettings)
		if err != nil {
			panic(fmt.Sprintf("could not create kafka Consumer: %s", err))
		}
		kafkaInput.consumers = append(kafkaInput.consumers, c)

		c.Assign(kafkaConfig.Assign)

		go func() {
			_, err = c.Consume(kafkaInput.messages)
			if err != nil {
				panic(fmt.Sprintf("try to consume error: %s", err))
			}
		}()
	}

	return kafkaInput
}

// ReadOneEvent implement method in topology.Input.
// gohangout call this method to get one event and pass it to filter or output
func (p *KafkaInput) ReadOneEvent() map[string]any {
	message, more := <-p.messages
	if !more {
		return nil
	}

	if message.Error != nil {
		klog.Error("kafka message carries error: ", message.Error)
		return nil
	}
	event := p.decoder.Decode(message.Message.Value)
	if p.decorateEvents {
		kafkaMeta := make(map[string]any)
		kafkaMeta["topic"] = message.TopicName
		kafkaMeta["partition"] = message.PartitionID
		kafkaMeta["offset"] = message.Message.Offset
		event["@metadata"] = map[string]any{"kafka": kafkaMeta}
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
