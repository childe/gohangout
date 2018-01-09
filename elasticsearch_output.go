package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/childe/gohangout/value_render"
	"github.com/golang/glog"
)

const (
	DEFAULT_BULK_SIZE      = 15
	DEFAULT_BULK_ACTIONS   = 5000
	DEFAULT_INDEX_TYPE     = "logs"
	DEFAULT_FLUSH_INTERVAL = 30
)

type HostSelector interface {
	selectOneHost() string
	reduceWeight(string)
	addWeight(string)
}

type RRHostSelector struct {
	hosts      []string
	initWeight int
	weight     []int
	index      int
	hostsCount int
}

func NewRRHostSelector(hosts []string, weight int) *RRHostSelector {
	hostsCount := len(hosts)
	rst := &RRHostSelector{
		hosts:      hosts,
		index:      0,
		hostsCount: hostsCount,
		initWeight: weight,
	}
	rst.weight = make([]int, hostsCount)
	for i := 0; i < hostsCount; i++ {
		rst.weight[i] = weight
	}

	return rst
}

func (s *RRHostSelector) selectOneHost() string {
	for i := 0; i < s.hostsCount; i++ {
		s.index = (s.index + 1) % s.hostsCount
		if s.weight[s.index] > 0 {
			return s.hosts[s.index]
		}
	}
	s.resetWeight(1)
	return ""
}

func (s *RRHostSelector) resetWeight(weight int) {
	for i := range s.weight {
		s.weight[i] = weight
	}
}

func (s *RRHostSelector) reduceWeight(host string) {
	for i, h := range s.hosts {
		if host == h {
			s.weight[i] = s.weight[i] - 1
			if s.weight[i] < 0 {
				s.weight[i] = 0
			}
			return
		}
	}
}

func (s *RRHostSelector) addWeight(host string) {
	for i, h := range s.hosts {
		if host == h {
			s.weight[i] = s.weight[i] + 1
			if s.weight[i] > s.initWeight {
				s.weight[i] = s.initWeight
			}
			return
		}
	}
}

type Action struct {
	op         string
	index      string
	index_type string
	id         string
	event      map[string]interface{}
}

type BulkRequest struct {
	actions   []*Action
	bulk_buf  [][]byte
	bulk_size int
}

func (br *BulkRequest) add(action *Action) {
	var meta []byte
	if action.id != "" {
		meta = []byte(fmt.Sprintf(`{"%s":{"_index":"%s","_type":"%s","_id":"%s"}}`, action.op, action.index, action.index_type, action.id))
	} else {
		meta = []byte(fmt.Sprintf(`{"%s":{"_index":"%s","_type":"%s"}}`, action.op, action.index, action.index_type))
	}
	buf, err := json.Marshal(action.event)
	if err != nil {
		glog.Errorf("could marshal event(%v):%s", action.event, err)
		return
	}

	br.bulk_buf = append(br.bulk_buf, meta)
	br.bulk_buf = append(br.bulk_buf, buf)

	br.bulk_size += len(buf) + len(meta)

	br.actions = append(br.actions, action)
}

func (br *BulkRequest) byteSize() int {
	return br.bulk_size
}
func (br *BulkRequest) actionCount() int {
	return len(br.actions)
}

type BulkProcessor interface {
	add(*Action)
	bulk()
	//flush()
}

type HTTPBulkProcessor struct {
	bulk_size      int
	bulk_actions   int
	flush_interval int
	execution_id   int
	client         *http.Client
	hostSelector   HostSelector
	bulkRequest    *BulkRequest
	mux            sync.Mutex
}

func (p *HTTPBulkProcessor) add(action *Action) {
	p.bulkRequest.add(action)

	if p.bulkRequest.byteSize() >= p.bulk_size || p.bulkRequest.actionCount() >= p.bulk_actions {
		p.bulk()
	}
}

//filter status if filterErrorType is nil
// else filter error type
func (p *HTTPBulkProcessor) abstraceBulkResponseItemsByStatus(bulkResponse map[string]interface{}) ([]int, []int) {
	glog.V(20).Infof("%v", bulkResponse)

	errors := bulkResponse["errors"].(bool)
	//glog.Infof("errors:%v", errors)

	retry := make([]int, 0)
	noRetry := make([]int, 0)

	if errors == false {
		return retry, noRetry
	}

	hasLog := false
	for i, item := range bulkResponse["items"].([]interface{}) {
		index := item.(map[string]interface{})["index"].(map[string]interface{})

		if errorValue, ok := index["error"]; ok {
			//errorType := errorValue.(map[string]interface{})["type"].(string)
			if !hasLog {
				glog.Infof("error :%v", errorValue)
				hasLog = true
			}

			status := index["status"].(float64)
			if status == 429 || status >= 500 {
				retry = append(retry, i)
			} else {
				noRetry = append(noRetry, i)
			}
		}
	}
	return retry, noRetry
}

func (p *HTTPBulkProcessor) tryOneBulk(url string, br *BulkRequest) bool {

	bulk_buf := make([]byte, 0)
	for _, buf := range br.bulk_buf {
		bulk_buf = append(bulk_buf, buf...)
		bulk_buf = append(bulk_buf, byte('\n'))
	}

	glog.V(20).Infof("%s", bulk_buf)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bulk_buf))
	req.Header.Set("Content-Type", "application/x-ndjson")

	resp, err := p.client.Do(req)

	if err != nil {
		glog.Infof("could not bulk with %s:%s", url, err)
		return false
	}

	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf(`read bulk response error:"%s". will NOT retry`, err)
		return true
	}
	glog.V(20).Infof("%s", respBody)

	var responseI interface{}
	err = json.Unmarshal(respBody, &responseI)

	if err != nil {
		glog.Errorf(`could not unmarshal bulk response:"%s". will NOT retry`, err)
		return true
	}

	bulkResponse := responseI.(map[string]interface{})
	shouldRetry, noRetry := p.abstraceBulkResponseItemsByStatus(bulkResponse)
	if len(shouldRetry) > 0 || len(noRetry) > 0 {
		glog.Infof("%d should retry; %d need not retry", len(shouldRetry), len(noRetry))
	}
	for _, i := range shouldRetry {
		p.add(br.actions[i])
	}
	return true
}

func (p *HTTPBulkProcessor) bulk() {
	p.mux.Lock()

	if p.bulkRequest.actionCount() == 0 {
		p.mux.Unlock()
		return
	}

	p.execution_id += 1
	glog.Infof("bulk %d docs with execution_id %d", p.bulkRequest.actionCount(), p.execution_id)

	bulkRequest := p.bulkRequest
	p.bulkRequest = &BulkRequest{}

	p.mux.Unlock()

	for {
		host := p.hostSelector.selectOneHost()
		if host == "" {
			//glog.Info("no available host, wait for next ticker")
			glog.Info("no available host, wait for 30s")
			time.Sleep(30 * time.Second)
			continue
		}

		glog.Infof("try to bulk with host (%s)", host)

		url := host + "/_bulk"
		success := p.tryOneBulk(url, bulkRequest)

		if success {
			glog.Infof("bulk done with execution_id %d", p.execution_id)
			p.hostSelector.addWeight(host)
			return
		}
		p.hostSelector.reduceWeight(host)
	}
}

func NewHTTPBulkProcessor(hosts []string, bulk_size, bulk_actions, flush_interval int) *HTTPBulkProcessor {
	bulkProcessor := &HTTPBulkProcessor{
		bulk_size:      bulk_size,
		bulk_actions:   bulk_actions,
		flush_interval: flush_interval,
		client:         &http.Client{},
		bulkRequest:    &BulkRequest{},
		hostSelector:   NewRRHostSelector(hosts, 3),
	}
	ticker := time.NewTicker(time.Second * time.Duration(flush_interval))
	go func() {
		for range ticker.C {
			bulkProcessor.bulk()
		}
	}()

	return bulkProcessor
}

type ElasticsearchOutput struct {
	config map[interface{}]interface{}

	index      value_render.ValueRender
	index_type value_render.ValueRender
	id         value_render.ValueRender

	bulkProcessor BulkProcessor
}

func NewElasticsearchOutput(config map[interface{}]interface{}) *ElasticsearchOutput {
	rst := &ElasticsearchOutput{
		config: config,
	}

	if indexValue, ok := config["index"]; ok {
		rst.index = value_render.GetValueRender(indexValue.(string))
	} else {
		glog.Fatal("index must be set in elasticsearch output")
	}

	if indextypeValue, ok := config["index_type"]; ok {
		rst.index_type = value_render.GetValueRender(indextypeValue.(string))
	} else {
		rst.index_type = value_render.GetValueRender(DEFAULT_INDEX_TYPE)
	}

	if idValue, ok := config["id"]; ok {
		rst.id = value_render.GetValueRender(idValue.(string))
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
		index      string = p.index.Render(event).(string)
		index_type string = p.index_type.Render(event).(string)
		op         string = "index"
		id         string
	)
	if p.id == nil {
		id = ""
	} else {
		id = p.id.Render(event).(string)
	}
	p.bulkProcessor.add(&Action{op, index, index_type, id, event})
}
