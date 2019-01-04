package input

import (
	"time"

	"github.com/childe/gohangout/codec"
	"github.com/childe/healer"
	"github.com/golang/glog"
)

type KafkaInput struct {
	BaseInput

	config        map[interface{}]interface{}
	fromBeginning bool

	messages chan *healer.FullMessage

	decoder codec.Decoder

	consumers []*healer.GroupConsumer

	simpleConsumers []*healer.SimpleConsumer // partitions config means we use simpleConsumers
}

func NewKafkaInput(config map[interface{}]interface{}) *KafkaInput {
	var (
		codertype string = "plain"
		topics    map[interface{}]interface{}
		assign    map[string][]int
	)

	consumer_settings := make(map[string]interface{})
	if v, ok := config["consumer_settings"]; !ok {
		glog.Fatal("kafka input must have consumer_settings")
	} else {
		for x, y := range v.(map[interface{}]interface{}) {
			consumer_settings[x.(string)] = y
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

	if v, ok := config["codec"]; ok {
		codertype = v.(string)
	}

	kafkaInput := &KafkaInput{
		BaseInput: BaseInput{},

		config:   config,
		messages: make(chan *healer.FullMessage, 100),

		decoder: codec.NewDecoder(codertype),
	}

	if v, ok := config["from_beginning"]; ok {
		kafkaInput.fromBeginning = v.(bool)
	} else {
		kafkaInput.fromBeginning = true
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
				kafkaInput.consumers = append(kafkaInput.consumers, c)

				_, err = c.Consume(kafkaInput.fromBeginning, kafkaInput.messages)
				if err != nil {
					glog.Fatalf("try to consumer error: %s", err)
				}
			}
		}
	} else {
		c, err := healer.NewConsumer(consumerConfig)
		if err != nil {
			glog.Fatalf("could not init SimpleConsumer: %s", err)
		}

		c.Assign(assign)

		kafkaInput.messages, err = c.Consume(kafkaInput.fromBeginning)
		if err != nil {
			glog.Fatalf("try to consume error: %s", err)
		}
	}

	return kafkaInput
}

func (inputPlugin *KafkaInput) readOneEvent() map[string]interface{} {
	message := <-inputPlugin.messages

	if message.Error != nil {
		return nil
	}
	return inputPlugin.decoder.Decode(message.Message.Value)
}

func (inputPlugin *KafkaInput) Shutdown() {
	for _, c := range inputPlugin.consumers {
		c.AwaitClose(30 * time.Second)
	}
}
