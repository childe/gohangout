package input

import (
	"reflect"
	"time"

	"github.com/childe/gohangout/codec"
	"github.com/childe/healer"
	"github.com/golang/glog"
)

type KafkaInput struct {
	config         map[interface{}]interface{}
	decorateEvents bool

	messages chan *healer.FullMessage

	decoder codec.Decoder

	groupConsumers []*healer.GroupConsumer
	consumers      []*healer.Consumer
}

func (l *MethodLibrary) NewKafkaInput(config map[interface{}]interface{}) *KafkaInput {
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
		for x, y := range v.(map[interface{}]interface{}) {
			if reflect.TypeOf(y).Kind() == reflect.Map {
				yy := make(map[string]interface{})
				for kk, vv := range y.(map[interface{}]interface{}) {
					yy[kk.(string)] = vv
				}
				consumer_settings[x.(string)] = yy
			} else {
				consumer_settings[x.(string)] = y
			}
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

	kafkaInput := &KafkaInput{
		config:         config,
		decorateEvents: decorateEvents,
		messages:       make(chan *healer.FullMessage, 10),

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
