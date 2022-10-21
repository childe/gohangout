package input

import (
	"context"
	"crypto/tls"
	sysjson "encoding/json"
	"github.com/childe/gohangout/topology"
	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
	"time"

	"github.com/childe/gohangout/codec"
	"github.com/golang/glog"
	kafka_go "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
)

// GoKafkaInput 使用Kafka-go的input插件
type GoKafkaInput struct {
	config         map[interface{}]interface{}
	decorateEvents bool
	messages       chan *kafka_go.Message
	decoder        codec.Decoder
	reader         *kafka_go.Reader
	readConfig     *kafka_go.ReaderConfig
}

func init() {
	Register("KafkaGo", newGoKafkaInput)
}

/*
New 插件模式的初始化
*/
func newGoKafkaInput(config map[interface{}]interface{}) topology.Input {

	var (
		codertype = "plain"
		topics    map[interface{}]interface{}
		assign    map[string][]int
	)

	consumer_settings := make(map[string]interface{})

	if v, ok := config["consumer_settings"]; !ok {
		glog.Fatal("kafka input must have consumer_settings")
	} else {
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

	if topics == nil && assign == nil {
		glog.Fatal("either topic or assign should be set")
	}
	if topics != nil && assign != nil {
		glog.Fatal("topic and assign can not be both set")
	}

	if codecV, ok := config["codec"]; ok {
		codertype = codecV.(string)
	}

	messagesLength := 10
	if v, ok := config["messages_queue_length"]; ok {
		messagesLength = v.(int)
	}

	p := &GoKafkaInput{
		messages:       make(chan *kafka_go.Message, messagesLength),
		decorateEvents: false,
		reader:         nil,
	}

	if decorateEventsV, ok := config["decorate_events"]; ok {
		p.decorateEvents = decorateEventsV.(bool)
	}

	if v, ok := config["decorateEvents"]; ok {
		p.decorateEvents = v.(bool)
	}

	if v, ok := config["code"]; ok {
		codertype = v.(string)
	}


	p.decoder = codec.NewDecoder(codertype)
	// 起携程，将所有收到的消息，存放到现在这个队列里面
	var err error

	if topics != nil {
		//threadCount
		for topic, _ := range topics {
			var readConfig *kafka_go.ReaderConfig

			readConfig, err = p.getConsumerConfig(consumer_settings)
			readConfig.Topic = topic.(string)

			if err == nil {
				p.reader = kafka_go.NewReader(*readConfig)
			} else {
				glog.Fatal("consumer_settings wrong")
			}

			go func() {
				for {
					m, err := p.reader.ReadMessage(context.Background())
					if err != nil {
						glog.Error("ReadMessage Error: ", err)
						break
					}
					//TODO 这里是不是要做一些异常检查
					p.messages <- &m
				}
			}()

		}
	}

	return p
}

// ReadOneEvent 单次事件的处理函数
func (p *GoKafkaInput) ReadOneEvent() map[string]interface{} {
	message, ok := <-p.messages
	if ok {
		event := p.decoder.Decode(message.Value)
		if p.decorateEvents {
			kafkaMeta := make(map[string]interface{})
			kafkaMeta["topic"] = message.Topic
			kafkaMeta["length"] = len(message.Value)
			kafkaMeta["partition"] = message.Partition
			kafkaMeta["offset"] = message.Offset
			event["@metadata"] = map[string]interface{}{"kafka": kafkaMeta}
		}
		return event
	}
	return nil
}

// Shutdown 关闭需要做的事情
func (p *GoKafkaInput) Shutdown() {
	if err := p.reader.Close(); err != nil {
		glog.Fatal("failed to close reader:", err)
	}
}

/*
HTTPKafka 增加一个状态获取的接口
*/
type HTTPKafka struct {
	kafka *GoKafkaInput
}


/*
*
格式转换
*/
func (p *GoKafkaInput) getConsumerConfig(config map[string]interface{}) (*kafka_go.ReaderConfig, error) {
	// 处理kafka参数

	c := &kafka_go.ReaderConfig{
		Brokers: make([]string, 1),
	}

	cs, err := sysjson.Marshal(config)
	if err != nil {
		glog.Fatal(err)
	}
	csStr := string(cs)

	bs := gjson.Get(csStr, "bootstrap\\.servers").Array()

	for _, i := range bs {
		c.Brokers = append(c.Brokers, i.String())
	}

	c.GroupID = gjson.Get(csStr, "group\\.id").String()
	c.MinBytes = int(gjson.Get(csStr, "MinBytes").Int())
	c.MaxBytes = int(gjson.Get(csStr, "MaxBytes").Int())

	c.Topic = ""

	if c.MinBytes == 0 {
		c.MinBytes = 10e3
	}
	if c.MaxBytes == 0 {
		c.MaxBytes = 10e6
	}

	HeartbeatInterval := gjson.Get(csStr, "HeartbeatInterval").Index
	if HeartbeatInterval != 0 {
		c.HeartbeatInterval = time.Duration(HeartbeatInterval) * time.Second
	}

	CommitInterval := gjson.Get(csStr, "CommitInterval").Index
	if CommitInterval != 0 {
		c.CommitInterval = time.Duration(CommitInterval) * time.Second
	}
	MaxWait := gjson.Get(csStr, "MaxWait").Index
	if MaxWait != 0 {
		c.MaxWait = time.Duration(MaxWait) * time.Second
	}
	SessionTimeout := gjson.Get(csStr, "SessionTimeout").Index
	if SessionTimeout != 0 {
		c.SessionTimeout = time.Duration(SessionTimeout) * time.Second
	}

	RebalanceTimeout := gjson.Get(csStr, "RebalanceTimeout").Index
	if RebalanceTimeout != 0 {
		c.RebalanceTimeout = time.Duration(RebalanceTimeout) * time.Second
	}

	var dialer *kafka_go.Dialer

	Timeout := gjson.Get(csStr, "Timeout").Index
	if Timeout != 0 {
		dialer = &kafka_go.Dialer{
			Timeout:   time.Duration(Timeout) * time.Second,
			DualStack: true,
		}
	} else {
		dialer = &kafka_go.Dialer{
			Timeout:   10 * time.Second,
			DualStack: true,
		}
	}

	if v, ok := config["SASL"]; ok {
		vh := v.(map[string]interface{})
		v1, ok1 := vh["Type"]
		v2, ok2 := vh["Username"]
		v3, ok3 := vh["Password"]
		if ok1 && ok2 && ok3 {
			var (
				mechanism sasl.Mechanism
				err       error
			)
			switch v1.(string) {
			case "Plain":
				mechanism = plain.Mechanism{
					Username: v2.(string),
					Password: v3.(string),
				}
			case "SCRAM":
				mechanism, err = scram.Mechanism(
					scram.SHA512,
					v2.(string),
					v3.(string),
				)
				if err != nil {
					glog.Fatal("ERROR FOR SCRAM: ", err)
				}
			default:
				glog.Fatalf("ERROR SASL type: %s", v1.(string))
			}
			dialer.SASLMechanism = mechanism
		} else {
			glog.Fatal("NO CONFIG FOR SASL")
		}
	}

	//TODO 后面再看是否需要完成了，理论上内网的KAFKA不会开启TLS，毕竟这个耗性能
	if v, ok := config["TLS"]; ok {
		vh := v.(map[string]interface{})
		var tls *tls.Config
		if _, ok1 := vh["PrivateKey"]; ok1 {
			glog.Info("TODO")
		}
		dialer.TLS = tls
	}

	c.Dialer = dialer

	return c, nil
}
