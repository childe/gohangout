package main

import (
	"github.com/childe/gohangout/codec"
	"github.com/childe/healer"
	"github.com/golang/glog"
)

type KafkaInput struct {
	config   map[interface{}]interface{}
	messages chan *healer.FullMessage

	decoder codec.Decoder
}

func NewKafkaInput(config map[interface{}]interface{}) *KafkaInput {
	var (
		codertype string = "plain"
		topics    map[interface{}]interface{}
	)

	consumer_settings := make(map[string]interface{})
	if v, ok := config["consumer_settings"]; !ok {
		glog.Fatal("kafka input must have consumer_settings")
	} else {
		for x, y := range v.(map[interface{}]interface{}) {
			consumer_settings[x.(string)] = y
		}
	}
	if v, ok := config["topic"]; !ok {
		glog.Fatal("kafka input must have topics")
	} else {
		topics = v.(map[interface{}]interface{})
	}
	if v, ok := config["codec"]; ok {
		codertype = v.(string)
	}

	kafkaInput := &KafkaInput{
		config:   config,
		messages: make(chan *healer.FullMessage, 100),

		decoder: codec.NewDecoder(codertype),
	}
	for topic, threadCount := range topics {

		for i := 0; i < threadCount.(int); i++ {

			consumer_settings["topic"] = topic
			c, err := healer.NewGroupConsumer(consumer_settings)
			if err != nil {
				glog.Fatalf("could not init GroupConsumer:%s", err)
			}

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

	if message.Error != nil {
		return nil
	}
	s := string(message.Message.Value)
	return inputPlugin.decoder.Decode(s)
}
