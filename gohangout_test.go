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

func (m *MockInput) ReadOneEvent() map[string]any {
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

		config := make(map[string]any)
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

		c := make(map[string]any)
		yaml.Unmarshal([]byte(_c), &c) //

		mockey.Mock(config.ParseConfig).Return(c, nil).Build()

		emit := mockey.Mock((*output.StdoutOutput).Emit).Return().Build()

		msgCount := 10
		mockey.Mock((*input.StdinInput).ReadOneEvent).To(func(*input.StdinInput) map[string]any {
			if msgCount > 0 {
				msgCount--
				return map[string]any{"msg": fmt.Sprintf("msg-%d", msgCount)}
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

		c := make(map[string]any)
		yaml.Unmarshal([]byte(_c), &c) //

		mockey.Mock(config.ParseConfig).Return(c, nil).Build()

		emit := mockey.Mock((*output.StdoutOutput).Emit).Return().Build()

		dateFilterOrigin := (*filter.DateFilter).Filter
		dateFilter := mockey.Mock((*filter.DateFilter).Filter).To(func(f *filter.DateFilter, event map[string]any) (map[string]any, bool) {
			return dateFilterOrigin(f, event)
		}).Origin(&dateFilterOrigin).Build()

		readOneEvent := mockey.Mock((*input.StdinInput).ReadOneEvent).To(func(*input.StdinInput) map[string]any {
			return map[string]any{"msg": "mock message"}
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

func TestEndToEndAllFilters(t *testing.T) {
	mockey.PatchConvey("end-to-end process testing, test all filtes", t, func() {
		_c := `
inputs:
  - Stdin:
      codec: json
filters:
  - Grok:
      failTag: grokfail
      match:
      - '^(?P<clientip>(?:(\b(?:[0-9A-Za-z][0-9A-Za-z-]{0,62})(?:\.(?:[0-9A-Za-z][0-9A-Za-z-]{0,62}))*(\.?|\b))|((?:(((([0-9A-Fa-f]{1,4}:){7}([0-9A-Fa-f]{1,4}|:))|(([0-9A-Fa-f]{1,4}:){6}(:[0-9A-Fa-f]{1,4}|((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){5}(((:[0-9A-Fa-f]{1,4}){1,2})|:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){4}(((:[0-9A-Fa-f]{1,4}){1,3})|((:[0-9A-Fa-f]{1,4})?:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){3}(((:[0-9A-Fa-f]{1,4}){1,4})|((:[0-9A-Fa-f]{1,4}){0,2}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){2}(((:[0-9A-Fa-f]{1,4}){1,5})|((:[0-9A-Fa-f]{1,4}){0,3}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){1}(((:[0-9A-Fa-f]{1,4}){1,6})|((:[0-9A-Fa-f]{1,4}){0,4}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(:(((:[0-9A-Fa-f]{1,4}){1,7})|((:[0-9A-Fa-f]{1,4}){0,5}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:)))(%.+)?)|((?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)))))) (?P<user>([a-zA-Z0-9._-]+)) \[(?P<ts>((?:(?:0[1-9])|(?:[12][0-9])|(?:3[01])|[1-9]))/(\b(?:Jan(?:uary)?|Feb(?:ruary)?|Mar(?:ch)?|Apr(?:il)?|May|Jun(?:e)?|Jul(?:y)?|Aug(?:ust)?|Sep(?:tember)?|Oct(?:ober)?|Nov(?:ember)?|Dec(?:ember)?)\b)/((\d\d){1,2}):(([^0-9]?)((?:2[0123]|[01]?[0-9])):((?:[0-5][0-9]))(?::((?:(?:[0-5][0-9]|60)(?:[:.,][0-9]+)?)))([^0-9]?)) ((?:[+-]?(?:[0-9]+))))\] "(?:(?P<verb>\b\w+\b) (?P<request>\S+)(?: HTTP/(?P<httpversion>(?:(([+-]?(?:[0-9]+(?:\.[0-9]+)?)|\.[0-9]+)))))?|(?P<rawrequest>.*?))" (?P<status>(?:(([+-]?(?:[0-9]+(?:\.[0-9]+)?)|\.[0-9]+)))) (?:(?P<bytes>(?:(([+-]?(?:[0-9]+(?:\.[0-9]+)?)|\.[0-9]+))))|-) "(?P<referrer>(.*))" (?P<ua>(.*))" "(?P<x_forwarded_for>(.*))" (?P<body>(.*))$'
  - Filters:
      if:
        - '!EQ($.tags,"grokfail")'
      filters:
      - Date:
          formats:
          - '02/Jan/2006:15:04:05 -0700'
          src: 'ts'
          target: timestamp
          remove_fields: [ts]
      - Convert:
          fields:
            status:
              remove_if_fail: false
              setto_if_fail: 0
              to: int
  - Drop:
      if:
      - 'EQ($.tags,"grokfail")'
outputs:
  - Stdout: {}
`

		c := make(map[string]any)
		yaml.Unmarshal([]byte(_c), &c) //

		mockey.Mock(config.ParseConfig).Return(c, nil).Build()

		rst := make([]map[string]any, 0)
		emit := mockey.Mock((*output.StdoutOutput).Emit).To(func(o *output.StdoutOutput, e map[string]any) {
			rst = append(rst, e)
		}).Build()

		eventCount := 0
		readOneEvent := mockey.Mock((*input.StdinInput).ReadOneEvent).To(func(*input.StdinInput) map[string]any {
			switch eventCount {
			case 0:
				eventCount++
				return map[string]any{"message": `10.0.0.100 - [09/Sep/2024:18:40:55 +0800] "GET /assets/loading.gif HTTP/1.1" 200 838555 "http://kafka-admin.corp.com/" "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36 Edg/128.0.0.0" "10.10.10.100" "-"`}
			case 1:
				eventCount++
				return map[string]any{"message": "mock message"}
			default:
				return nil
			}
		}).Build()

		options.exitWhenNil = true
		_main()

		convey.So(readOneEvent.Times(), convey.ShouldEqual, 3)
		convey.So(emit.Times(), convey.ShouldEqual, 1)
		convey.So(len(rst), convey.ShouldEqual, 1)
		event := rst[0]
		convey.So(event["x_forwarded_for"], convey.ShouldEqual, "10.10.10.100")
		convey.So(event["body"], convey.ShouldEqual, `"-"`)
		convey.So(event["status"], convey.ShouldEqual, 200)
		convey.So(event["timestamp"], convey.ShouldNotBeNil)
		convey.So(event["ts"], convey.ShouldBeNil)
	})
}
