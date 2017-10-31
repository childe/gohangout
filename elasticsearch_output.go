package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
)

const (
	DEFAULT_BULK_SIZE      = 15
	DEFAULT_BULK_ACTIONS   = 5000
	DEFAULT_INDEX_TYPE     = "logs"
	DEFAULT_FLUSH_INTERVAL = 30
)

type BulkProcessor interface {
	add(string, string, string, string, map[string]interface{})
	bulk()
}

type HTTPBulkProcessor struct {
	hosts []string

	docs     []map[string]interface{}
	bulk_buf [][]byte

	bulk_size      int
	bulk_actions   int
	flush_interval int

	current_bulk_size int
}

func (p *HTTPBulkProcessor) add(index, index_type, id, op string, event map[string]interface{}) {
	var meta []byte
	if id != "" {
		meta = []byte(fmt.Sprintf(`{"%s":{"_index":"%s","_type":"%s","_id":"%s"}}`, op, index, index_type, id))
	} else {
		meta = []byte(fmt.Sprintf(`{"%s":{"_index":"%s","_type":"%s"}}`, op, index, index_type))
	}
	buf, err := json.Marshal(event)
	if err != nil {
		glog.Errorf("could marshal event(%v):%s", event, err)
		return
	}

	p.bulk_buf = append(p.bulk_buf, meta)
	p.bulk_buf = append(p.bulk_buf, buf)

	p.current_bulk_size += len(buf) + len(meta)

	p.docs = append(p.docs, event)

	if p.current_bulk_size >= p.bulk_size || len(p.bulk_buf)/2 >= p.bulk_actions {
		p.bulk()
	}
}

func (p *HTTPBulkProcessor) bulk() {
	var url = p.hosts[0] + "/_bulk"
	glog.Info(url)
	bulk_buf := make([]byte, 0)
	for _, buf := range p.bulk_buf {
		bulk_buf = append(bulk_buf, buf...)
		bulk_buf = append(bulk_buf, byte('\n'))
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bulk_buf))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		glog.Errorf("bulk error:%s", err)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("read bulk response error:%s", err)
	}
	glog.Infof("%s", respBody)
}

func NewHTTPBulkProcessor(hosts []string, bulk_size, bulk_actions, flush_interval int) *HTTPBulkProcessor {
	return &HTTPBulkProcessor{
		hosts:          hosts,
		bulk_size:      bulk_size,
		bulk_actions:   bulk_actions,
		flush_interval: flush_interval,
	}
}

type ElasticsearchOutput struct {
	config map[interface{}]interface{}

	index      ValueRender
	index_type ValueRender
	id         ValueRender

	bulkProcessor BulkProcessor
}

func NewElasticsearchOutput(config map[interface{}]interface{}) *ElasticsearchOutput {
	rst := &ElasticsearchOutput{
		config: config,
	}

	if indexValue, ok := config["index"]; ok {
		rst.index = getValueRender(indexValue.(string))
	} else {
		glog.Fatal("index must be set in elasticsearch output")
	}

	if indextypeValue, ok := config["index_type"]; ok {
		rst.index_type = getValueRender(indextypeValue.(string))
	} else {
		rst.index_type = getValueRender(DEFAULT_INDEX_TYPE)
	}

	if idValue, ok := config["id"]; ok {
		rst.id = getValueRender(idValue.(string))
	} else {
		rst.id = nil
	}

	var bulk_size, bulk_actions, flush_interval int
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

	var hosts []string
	if v, ok := config["hosts"]; ok {
		for _, h := range v.([]interface{}) {
			hosts = append(hosts, h.(string))
		}
	} else {
		glog.Fatal("hosts must be set in elasticsearch output")
	}

	rst.bulkProcessor = NewHTTPBulkProcessor(hosts, bulk_size, bulk_actions, flush_interval)
	return rst
}

func (p *ElasticsearchOutput) emit(event map[string]interface{}) {
	var (
		index      string = p.index.render(event).(string)
		index_type string = p.index_type.render(event).(string)
		op         string = "index"
		id         string
	)
	if p.id == nil {
		id = ""
	} else {
		id = p.id.render(event).(string)
	}
	p.bulkProcessor.add(index, index_type, id, op, event)
}
