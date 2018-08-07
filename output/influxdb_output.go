package output

import (
	"net/http"
	"time"

	"github.com/childe/gohangout/value_render"
	"github.com/golang/glog"
)

const ()

type InAction struct {
	measurement string
	event       map[string]interface{}
}

func (action *InAction) Encode() []byte {
	bulk_buf := make([]byte, 0)
	bulk_buf = append(bulk_buf, '\n')
	return bulk_buf
}

type InfluxdbBulkRequest struct {
	events   []Event
	bulk_buf []byte
}

func (br *InfluxdbBulkRequest) add(event Event) {
	br.bulk_buf = append(br.bulk_buf, event.Encode()...)
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
	BaseOutput
	config map[interface{}]interface{}

	measurement value_render.ValueRender

	bulkProcessor BulkProcessor
}

func influxdbGetRetryEvents(resp *http.Response, respBody []byte) ([]int, []int) {
	return nil, nil
}
func NewInfluxdbOutput(config map[interface{}]interface{}) *InfluxdbOutput {
	rst := &InfluxdbOutput{
		BaseOutput: NewBaseOutput(config),
		config:     config,
	}

	if v, ok := config["measurement"]; ok {
		rst.measurement = value_render.GetValueRender(v.(string))
	} else {
		glog.Fatal("measurement must be set in elasticsearch output")
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
			hosts = append(hosts, h.(string))
		}
	} else {
		glog.Fatal("hosts must be set in elasticsearch output")
	}

	var headers = map[string]string{"Content-Type": "application/x-ndjson"}
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
		return &ESBulkRequest{
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
	p.bulkProcessor.add(&InAction{measurement, event})
}
func (outputPlugin *InfluxdbOutput) Shutdown() {
	outputPlugin.bulkProcessor.awaitclose(30 * time.Second)
}
