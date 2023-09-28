// 使用 https://github.com/Shopify/sarama kafka 库
package input

import (
	"context"
	sysjson "encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/childe/gohangout/codec"
	"github.com/childe/gohangout/topology"
	"k8s.io/klog"
)

var (
	running               bool
	defaultConsumerConfig = ConsumerConfig{
		NetConfig: NetConfig{
			ConnectTimeoutMS:    30000,
			TimeoutMS:           30000,
			TimeoutMSForEachAPI: make([]int, 0),
			KeepAliveMS:         7200000,
		},
		ClientID:             "etl-ingest",
		GroupID:              "",
		SessionTimeoutMS:     30000,
		RetryBackOffMS:       100,
		MetadataMaxAgeMS:     300000,
		FetchMaxWaitMS:       500,
		FetchMaxBytes:        10 * 1024 * 1024,
		FetchMinBytes:        1,
		FromBeginning:        false,
		AutoCommit:           true,
		AutoCommitIntervalMS: 5000,
		OffsetsStorage:       1,
		Version:              sarama.V1_0_0_0.String(),
	}
)

type ConsumerConfig struct {
	NetConfig
	SaslConfig           *SaslConfig `json:"sasl_config" mapstructure:"sasl_config"`
	BootstrapServers     string      `json:"bootstrap.servers" mapstructure:"bootstrap.servers"`
	ClientID             string      `json:"client.id" mapstructure:"client.id"`
	GroupID              string      `json:"group.id" mapstructure:"group.id"`
	RetryBackOffMS       int         `json:"retry.backoff.ms,string" mapstructure:"retry.backoff.ms"`
	MetadataMaxAgeMS     int         `json:"metadata.max.age.ms,string" mapstructure:"metadata.max.age.ms"`
	SessionTimeoutMS     int32       `json:"session.timeout.ms,string" mapstructure:"session.timeout.ms"`
	FetchMaxWaitMS       int32       `json:"fetch.max.wait.ms,string" mapstructure:"fetch.max.wait.ms"`
	FetchMaxBytes        int32       `json:"fetch.max.bytes,string" mapstructure:"fetch.max.bytes"`
	FetchMinBytes        int32       `json:"fetch.min.bytes,string" mapstructure:"fetch.min.bytes"`
	FromBeginning        bool        `json:"from.beginning,string" mapstructure:"from.beginning"`
	AutoCommit           bool        `json:"auto.commit,string" mapstructure:"auto.commit"`
	AutoCommitIntervalMS int         `json:"auto.commit.interval.ms,string" mapstructure:"auto.commit.interval.ms"`
	OffsetsStorage       int         `json:"offsets.storage,string" mapstructure:"offsets.storage"`
	Version              string      `json:"version"`

	MetadataRefreshIntervalMS int `json:"metadata.refresh.interval.ms,string" mapstructure:"metadata.refresh.interval.ms"`
}

type NetConfig struct {
	ConnectTimeoutMS    int   `json:"connect.timeout.ms,string" mapstructure:"connect.timeout.ms"`
	TimeoutMS           int   `json:"timeout.ms,string" mapstructure:"timeout.ms"`
	TimeoutMSForEachAPI []int `json:"timeout.ms.for.eachapi" mapstructure:"timeout.ms.for.eachapi"`
	KeepAliveMS         int   `json:"keepalive.ms,string" mapstructure:"keepalive.ms"`
}

type SaslConfig struct {
	SaslMechanism string `json:"sasl.mechanism" mapstructure:"sasl.mechanism"`
	SaslUser      string `json:"sasl.user" mapstructure:"sasl.user"`
	SaslPassword  string `json:"sasl.password" mapstructure:"sasl.password"`
}

type KafkaSaramaInput struct {
	config         map[interface{}]interface{}
	decorateEvents bool

	messages chan ConsumerMessageAndSession

	decoder codec.Decoder

	client   sarama.ConsumerGroup
	shutdown bool
}

type ConsumerMessageAndSession struct {
	session sarama.ConsumerGroupSession
	message *sarama.ConsumerMessage
}

func init() {
	Register("Kafka_Sarama", newKafkaInput)
}

func newKafkaSaramaInput(config map[interface{}]interface{}) (topology.Input, error) {
	var (
		codertype      = "plain"
		decorateEvents = false
		topics         []string
	)
	consumer_settings := make(map[string]interface{})
	if v, ok := config["consumer_settings"]; !ok {
		return nil, fmt.Errorf("kafka input must have consumer_settings")
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
	if v, ok := config["topic"].([]interface{}); ok {
		for _, topic := range v {
			topics = append(topics, topic.(string))
		}
	} else {
		topics = nil
	}

	if topics == nil {
		return nil, fmt.Errorf("topic should be set")
	}

	if codecV, ok := config["codec"]; ok {
		codertype = codecV.(string)
	}

	if decorateEventsV, ok := config["decorate_events"]; ok {
		decorateEvents = decorateEventsV.(bool)
	}

	brokers, groupID, consumerConfig, err := getConsumerConfig(consumer_settings)
	if err != nil {
		return nil, fmt.Errorf("error in consumer settings: %s", err)
	}

	kafkaSaramaInput := &KafkaSaramaInput{
		config:         config,
		decorateEvents: decorateEvents,
		messages:       make(chan ConsumerMessageAndSession, 10),
		decoder:        codec.NewDecoder(codertype),
		shutdown:       false,
	}
	consumer := Consumer{
		ready:    make(chan bool),
		messages: kafkaSaramaInput.messages,
	}
	sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)

	go func() {
		for !kafkaSaramaInput.shutdown {
			klog.Infof("start connect/reconnect kafka and consume from cluster: %v, topic: %v, groupName: %v",
				brokers, topics, groupID)

			// GroupConsumer
			ctx := context.Background()
			client, err := sarama.NewConsumerGroup(brokers, groupID, consumerConfig)
			if err != nil {
				klog.Errorf("could not init GroupConsumer: %s", err)
				if client != nil {
					client.Close()
				}

				time.Sleep(10 * time.Second)
				continue
			}
			kafkaSaramaInput.client = client

			go func() {
				for err := range client.Errors() {
					klog.Errorf("consumer error: %v", err)
				}
			}()

			for {
				// `Consume` should be called inside an infinite loop, when a
				// server-side rebalance happens, the consumer session will need to be
				// recreated to get the new claims
				if err := client.Consume(ctx, topics, &consumer); err != nil {
					klog.Errorf("Error from consumer: %v, will close and reconnect", err)
					client.Close()
					break
				}
				// check if context was cancelled, signaling that the consumer should stop
				if kafkaSaramaInput.shutdown {
					return
				}
				consumer.ready = make(chan bool)
				klog.Info("Sarama consumer up and running!...")
			}
			time.Sleep(10 * time.Second) // 运行过程中断开链接时，10秒后重新连接
		}
	}()

	// <-consumer.ready // Await till the consumer has been set up

	return kafkaSaramaInput, nil
}

func (p *KafkaSaramaInput) ReadOneEvent() map[string]interface{} {
	messageAndSession, ok := <-p.messages
	if ok {
		message := messageAndSession.message
		event := p.decoder.Decode(message.Value)
		if p.decorateEvents {
			kafkaMeta := make(map[string]interface{})
			kafkaMeta["topic"] = message.Topic
			kafkaMeta["partition"] = message.Partition
			kafkaMeta["offset"] = message.Offset
			kafkaMeta["session"] = messageAndSession.session
			event["__metadata"] = map[string]interface{}{"kafka": kafkaMeta}
		}
		return event
	}
	return nil
}

func (p *KafkaSaramaInput) Pause() {
	if p.client != nil {
		p.client.PauseAll()
	}
}

func (p *KafkaSaramaInput) Shutdown() {
	p.shutdown = true
	if running && p.client != nil {
		// 避免重复close
		klog.Info("Shutting down consumer...")
		running = false
		if err := p.client.Close(); err != nil {
			klog.Errorf("Error closing client: %v", err)
		}
	}
}

// 未适配
// TLS 全部
// offsets.storage
// connect.timeout.ms
// timeout.ms.for.eachapi
func getConsumerConfig(config map[string]interface{}) (brokers []string, groupId string, cfg *sarama.Config, err error) {
	b, err := sysjson.Marshal(config)
	if err != nil {
		return
	}
	dc := defaultConsumerConfig
	err = sysjson.Unmarshal(b, &dc)
	if err != nil {
		return
	}

	brokers = strings.Split(dc.BootstrapServers, ",")
	groupId = dc.GroupID

	cfg = sarama.NewConfig()
	cfg.Version, err = sarama.ParseKafkaVersion(dc.Version)
	if err != nil {
		return
	}

	cfg.ClientID = dc.ClientID
	cfg.Metadata.Timeout = time.Duration(dc.MetadataMaxAgeMS) * time.Millisecond

	cfg.Consumer.Retry.Backoff = time.Duration(dc.RetryBackOffMS) * time.Millisecond
	cfg.Consumer.Group.Session.Timeout = time.Duration(dc.SessionTimeoutMS) * time.Millisecond
	cfg.Consumer.MaxWaitTime = time.Duration(dc.FetchMaxWaitMS) * time.Millisecond
	cfg.Consumer.Fetch.Max = dc.FetchMaxBytes
	cfg.Consumer.Fetch.Min = dc.FetchMinBytes

	if dc.FromBeginning {
		klog.Infoln("OffsetOldest")
		cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	} else {
		klog.Infoln("OffsetNewest")
		cfg.Consumer.Offsets.Initial = sarama.OffsetNewest
	}
	cfg.Consumer.Offsets.AutoCommit.Enable = dc.AutoCommit
	cfg.Consumer.Offsets.AutoCommit.Interval = time.Duration(dc.AutoCommitIntervalMS) * time.Millisecond

	cfg.Net.DialTimeout = time.Duration(dc.NetConfig.TimeoutMS) * time.Millisecond
	cfg.Net.KeepAlive = time.Duration(dc.NetConfig.KeepAliveMS) * time.Millisecond

	if dc.SaslConfig != nil && len(dc.SaslConfig.SaslUser) > 0 &&
		len(dc.SaslConfig.SaslPassword) > 0 && len(dc.SaslConfig.SaslMechanism) > 0 {
		cfg.Net.SASL.Enable = true
		cfg.Net.SASL.User = dc.SaslConfig.SaslUser
		cfg.Net.SASL.Password = dc.SaslConfig.SaslPassword
		cfg.Net.SASL.Mechanism = sarama.SASLMechanism(dc.SaslConfig.SaslMechanism)
	}

	return
}

// Consumer represents a Sarama consumer group consumer
type Consumer struct {
	ready    chan bool
	messages chan ConsumerMessageAndSession
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(session sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(consumer.ready)
	running = true
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	running = false
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
// Once the Messages() channel is closed, the Handler must finish its processing
// loop and exit.
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/IBM/sarama/blob/main/consumer_group.go#L27-L29
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				klog.Info("message channel was closed")
				return nil
			}
			klog.V(10).Infof("Message claimed: topic = %s partition = %d offset = %d", message.Topic, message.Partition, message.Offset)
			messageAndSession := ConsumerMessageAndSession{
				session: session,
				message: message,
			}
			consumer.messages <- messageAndSession
		// Should return when `session.Context()` is done.
		// If not, will raise `ErrRebalanceInProgress` or `read tcp <ip>:<port>: i/o timeout` when kafka rebalance. see:
		// https://github.com/IBM/sarama/issues/1192
		case <-session.Context().Done():
			return nil
		}
	}
}
