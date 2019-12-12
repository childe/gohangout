package output

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/childe/gohangout/value_render"
	"github.com/golang/glog"
)

const ()

type InAction struct {
	measurement string
	event       map[string]interface{}
	tags        []string
	fields      []string
	timestamp   string
}

func (action *InAction) Encode() []byte {
	bulk_buf := []byte(action.measurement)

	//tag set
	tag_set := make([]string, 0)
	for _, tag := range action.tags {
		if v, ok := action.event[tag]; ok {
			tag_set = append(tag_set, fmt.Sprintf("%s=%v", tag, v))
		}
	}
	if len(tag_set) > 0 {
		bulk_buf = append(bulk_buf, ',')
		bulk_buf = append(bulk_buf, strings.Join(tag_set, ",")...)
	}

	//field set
	field_set := make([]string, 0)
	for _, field := range action.fields {
		if v, ok := action.event[field]; ok {
			field_set = append(field_set, fmt.Sprintf("%s=%v", field, v))
		}
	}
	if len(field_set) <= 0 {
		glog.V(20).Infof("field set is nil. fields: %v. event: %v", action.fields, action.event)
		return nil
	} else {
		bulk_buf = append(bulk_buf, ' ')
		bulk_buf = append(bulk_buf, strings.Join(field_set, ",")...)
	}

	//timestamp
	t := action.event[action.timestamp]
	if t != nil && reflect.TypeOf(t).String() == "time.Time" {
		bulk_buf = append(bulk_buf, fmt.Sprintf(" %d", t.(time.Time).UnixNano())...)
	} else {
		glog.V(20).Infof("%s is not time.Time", action.timestamp)
	}

	return bulk_buf
}

type InfluxdbBulkRequest struct {
	events   []Event
	bulk_buf []byte
}

func (br *InfluxdbBulkRequest) add(event Event) {
	br.bulk_buf = append(br.bulk_buf, event.Encode()...)
	br.bulk_buf = append(br.bulk_buf, '\n')
	br.events = append(br.events, event)
}

func (br *InfluxdbBulkRequest) bufSizeByte() int {
	return len(br.bulk_buf)
}
func (br *InfluxdbBulkRequest) eventCount() int {
	return len(br.events)
}
func (br *InfluxdbBulkRequest) readBuf() []byte {
	return br.bulk_buf
}

type InfluxdbOutput struct {
	config map[interface{}]interface{}

	db          string
	measurement value_render.ValueRender
	tags        []string
	fields      []string
	timestamp   string

	bulkProcessor BulkProcessor
}

func influxdbGetRetryEvents(resp *http.Response, respBody []byte, bulkRequest *BulkRequest) ([]int, []int, BulkRequest) {
	return nil, nil, nil
}

func (l *MethodLibrary) NewInfluxdbOutput(config map[interface{}]interface{}) *InfluxdbOutput {
	rst := &InfluxdbOutput{
		config: config,
	}

	if v, ok := config["db"]; ok {
		rst.db = v.(string)
	} else {
		glog.Fatal("db must be set in elasticsearch output")
	}

	if v, ok := config["measurement"]; ok {
		rst.measurement = value_render.GetValueRender(v.(string))
	} else {
		glog.Fatal("measurement must be set in elasticsearch output")
	}

	if v, ok := config["tags"]; ok {
		for _, t := range v.([]interface{}) {
			rst.tags = append(rst.tags, t.(string))
		}
	}
	if v, ok := config["fields"]; ok {
		for _, f := range v.([]interface{}) {
			rst.fields = append(rst.fields, f.(string))
		}
	}
	if v, ok := config["timestamp"]; ok {
		rst.timestamp = v.(string)
	} else {
		rst.timestamp = "@timestamp"
	}

	var (
		bulk_size, bulk_actions, flush_interval, concurrent int
		compress                                            bool
	)
	if v, ok := config["bulk_size"]; ok {
		bulk_size = v.(int) * 1024 * 1024
	} else {
		bulk_size = DEFAULT_BULK_SIZE
	}

	if v, ok := config["bulk_actions"]; ok {
		bulk_actions = v.(int)
	} else {
		bulk_actions = DEFAULT_BULK_ACTIONS
	}
	if v, ok := config["flush_interval"]; ok {
		flush_interval = v.(int)
	} else {
		flush_interval = DEFAULT_FLUSH_INTERVAL
	}
	if v, ok := config["concurrent"]; ok {
		concurrent = v.(int)
	} else {
		concurrent = DEFAULT_CONCURRENT
	}
	if concurrent <= 0 {
		glog.Fatal("concurrent must > 0")
	}
	if v, ok := config["compress"]; ok {
		compress = v.(bool)
	} else {
		compress = true
	}

	var hosts []string
	if v, ok := config["hosts"]; ok {
		for _, h := range v.([]interface{}) {
			hosts = append(hosts, h.(string)+"/write?db="+rst.db)
		}
	} else {
		glog.Fatal("hosts must be set in elasticsearch output")
	}

	headers := make(map[string]string)
	if v, ok := config["headers"]; ok {
		for keyI, valueI := range v.(map[interface{}]interface{}) {
			headers[keyI.(string)] = valueI.(string)
		}
	}
	var requestMethod string = "POST"

	var retryResponseCode map[int]bool
	if v, ok := config["retry_response_code"]; ok {
		for _, cI := range v.([]interface{}) {
			retryResponseCode[cI.(int)] = true
		}
	}

	byte_size_applied_in_advance := bulk_size + 1024*1024
	if byte_size_applied_in_advance > MAX_BYTE_SIZE_APPLIED_IN_ADVANCE {
		byte_size_applied_in_advance = MAX_BYTE_SIZE_APPLIED_IN_ADVANCE
	}
	var f = func() BulkRequest {
		return &InfluxdbBulkRequest{
			bulk_buf: make([]byte, 0, byte_size_applied_in_advance),
		}
	}

	rst.bulkProcessor = NewHTTPBulkProcessor(headers, hosts, requestMethod, retryResponseCode, bulk_size, bulk_actions, flush_interval, concurrent, compress, f, influxdbGetRetryEvents)
	return rst
}

func (p *InfluxdbOutput) Emit(event map[string]interface{}) {
	var (
		measurement string = p.measurement.Render(event).(string)
	)
	p.bulkProcessor.add(&InAction{measurement, event, p.tags, p.fields, p.timestamp})
}
func (outputPlugin *InfluxdbOutput) Shutdown() {
	outputPlugin.bulkProcessor.awaitclose(30 * time.Second)
}
