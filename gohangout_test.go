package main

import (
	"testing"

	"github.com/bytedance/mockey"
	"github.com/childe/gohangout/input"
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

func TestEndToEndProcess(t *testing.T) {
	mockey.PatchConvey("end-to-end process testing", t, func() {
		c := `
inputs:
    - Mock:
        codec: json
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
