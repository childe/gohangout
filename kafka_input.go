package main

import (
	"time"

	"github.com/childe/healer"
	"github.com/golang/glog"
)

type KafkaInput struct {
	config map[interface{}]interface{}
	//consumers []*healer.GroupConsumer
	messages chan *healer.FullMessage

	deserializer Deserializer
}

func NewKafkaInput(config map[interface{}]interface{}) *KafkaInput {
	var (
		brokers string
		groupID string
		topics  map[interface{}]interface{}
	)

	if v, ok := config["consumer_settings"]; !ok {
		glog.Fatal("kafka input must have consumer_settings")
	} else {
		value := v.(map[interface{}]interface{})
		brokers = value["bootstrap.servers"].(string)
		groupID = value["group.id"].(string)
	}
	if v, ok := config["topic"]; !ok {
		glog.Fatal("kafka input must have topics")
	} else {
		topics = v.(map[interface{}]interface{})
	}

	kafkaInput := &KafkaInput{
		config: config,
		//consumers: make([]*healer.GroupConsumer, 0),
		messages: make(chan *healer.FullMessage, 100),

		deserializer: NewHermesDeserializer(),
	}
	for topic, threadCount := range topics {

		for i := 0; i < threadCount.(int); i++ {

			config := make(map[string]interface{})
			config["brokers"] = brokers
			config["topic"] = topic
			config["groupID"] = groupID
			c, err := healer.NewGroupConsumer(config)
			if err != nil {
				glog.Fatalf("could not init GroupConsumer:%s", err)
			}

			//kafkaInput.consumers = append(kafkaInput.consumers, c)

			_, err = c.Consume(true, kafkaInput.messages)
			if err != nil {
				glog.Fatalf("could not get messages channel:%s", err)
			}
		}
	}

	return kafkaInput
}

func (inputPlugin *KafkaInput) readOneEvent() map[string]interface{} {
	message := <-inputPlugin.messages
	rst := make(map[string]interface{})
	//rst["message"] = string(message.Message.Value)
	rst["message"] = inputPlugin.deserializer.deserialize("", message.Message.Value)
	rst["@timestamp"] = time.Now()
	return rst
}
