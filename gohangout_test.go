package main

import (
	"fmt"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/childe/gohangout/filter"
	"github.com/childe/gohangout/input"
	"github.com/childe/gohangout/internal/config"
	"github.com/childe/gohangout/output"
	"github.com/smartystreets/goconvey/convey"
	yaml "gopkg.in/yaml.v2"
)

type MockInput struct{}

func (m *MockInput) ReadOneEvent() map[string]interface{} {
	return nil
}
func (m *MockInput) Shutdown() {}

func TestBuildPluginLink(t *testing.T) {
	mockey.PatchConvey("test buildPluginLink", t, func() {
		c := `inputs:
    - Stdin:
        codec: json
    - Kafka:
        topic:
            test-none: 1
        codec: json
        consumer_settings:
          bootstrap.servers: "127.0.0.1:9094"
          group.id: gohangout-test-none
          retry.backoff.ms: 1000
          sasl:
            mechanism: PLAIN
            user: admin
            password: admin-secret

filters:
  - Date:
      src: '@timestamp'
      target: pt
      formats:
        - '2006-01-02T15:04:05.000-07:00'

outputs:
  - Stdout: {}`

		config := make(map[string]interface{})
		yaml.Unmarshal([]byte(c), &config)

		mockey.Mock(input.GetInput).Return(&MockInput{}).Build()

		boxes, err := buildPluginLink(config)
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(boxes), convey.ShouldEqual, 2)
	})
}

func TestEndToEndProcessNilExist(t *testing.T) {
	mockey.PatchConvey("end-to-end process testing, read 10 messages and then nil and exist", t, func() {
		_c := `
inputs:
    - Stdin:
        codec: json
filters:
  - Date:
      src: '@timestamp'
      target: pt
      formats:
        - '2006-01-02T15:04:05.000-07:00'

outputs:
  - Stdout: {}`

		c := make(map[string]interface{})
		yaml.Unmarshal([]byte(_c), &c) //

		mockey.Mock(config.ParseConfig).Return(c, nil).Build()

		emit := mockey.Mock((*output.StdoutOutput).Emit).Return().Build()

		msgCount := 10
		mockey.Mock((*input.StdinInput).ReadOneEvent).To(func(*input.StdinInput) map[string]interface{} {
			if msgCount > 0 {
				msgCount--
				return map[string]interface{}{"msg": fmt.Sprintf("msg-%d", msgCount)}
			} else {
				return nil
			}
		}).Build()

		options.exitWhenNil = true
		_main()

		convey.So(emit.Times(), convey.ShouldEqual, 10)
	})
}

func TestEndToEndProcess(t *testing.T) {
	mockey.PatchConvey("end-to-end process testing, read 10 messages and then cancel", t, func() {
		_c := `
inputs:
    - Stdin:
        codec: json
filters:
  - Date:
      src: '@timestamp'
      target: pt
      formats:
        - '2006-01-02T15:04:05.000-07:00'

outputs:
  - Stdout: {}`

		c := make(map[string]interface{})
		yaml.Unmarshal([]byte(_c), &c) //

		mockey.Mock(config.ParseConfig).Return(c, nil).Build()

		emit := mockey.Mock((*output.StdoutOutput).Emit).Return().Build()

		dateFilterOrigin := (*filter.DateFilter).Filter
		dateFilter := mockey.Mock((*filter.DateFilter).Filter).To(func(f *filter.DateFilter, event map[string]interface{}) (map[string]interface{}, bool) {
			return dateFilterOrigin(f, event)
		}).Origin(&dateFilterOrigin).Build()

		readOneEvent := mockey.Mock((*input.StdinInput).ReadOneEvent).To(func(*input.StdinInput) map[string]interface{} {
			return map[string]interface{}{"msg": "mock message"}
		}).Build()

		ch := make(chan struct{})
		go func() {
			_main()
		}()

		for {
			if emit.Times() == 10 {
				close(ch)
				break
			}
		}

		<-ch
		cancel()

		convey.So(readOneEvent.Times(), convey.ShouldBeGreaterThanOrEqualTo, 10)
		convey.So(emit.Times(), convey.ShouldBeGreaterThanOrEqualTo, 10)
		convey.So(dateFilter.Times(), convey.ShouldBeGreaterThanOrEqualTo, 10)

		convey.So(readOneEvent.Times(), convey.ShouldBeGreaterThanOrEqualTo, dateFilter.Times())
		convey.So(dateFilter.Times(), convey.ShouldBeGreaterThanOrEqualTo, emit.Times())
	})
}
