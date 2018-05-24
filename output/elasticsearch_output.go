package output

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"

	"github.com/childe/gohangout/value_render"
	"github.com/golang/glog"
)

const (
	DEFAULT_BULK_SIZE      = 15
	DEFAULT_BULK_ACTIONS   = 5000
	DEFAULT_INDEX_TYPE     = "logs"
	DEFAULT_FLUSH_INTERVAL = 30
	DEFAULT_CONCURRENT     = 1
	META_FORMAT_WITH_ID    = `{"%s":{"_index":"%s","_type":"%s","_id":"%s","routing":"%s"}}` + "\n"
	META_FORMAT_WITHOUT_ID = `{"%s":{"_index":"%s","_type":"%s","routing":"%s"}}` + "\n"
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
	// reset weight and return "" if all hosts are down
	var hasAtLeastOneUp bool = false
	for i := 0; i < s.hostsCount; i++ {
		if s.weight[i] > 0 {
			hasAtLeastOneUp = true
		}
	}
	if !hasAtLeastOneUp {
		s.resetWeight(s.initWeight)
		return ""
	}

	s.index = (s.index + 1) % s.hostsCount
	return s.hosts[s.index]
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
	routing    string
	event      map[string]interface{}
}

type BulkRequest struct {
	actions  []*Action
	bulk_buf []byte
}

func (br *BulkRequest) add(action *Action) {
	var meta []byte
	if action.id != "" {
		meta = []byte(fmt.Sprintf(META_FORMAT_WITH_ID, action.op, action.index, action.index_type, action.id, action.routing))
	} else {
		meta = []byte(fmt.Sprintf(META_FORMAT_WITHOUT_ID, action.op, action.index, action.index_type, action.routing))
	}
	buf, err := json.Marshal(action.event)
	if err != nil {
		glog.Errorf("could marshal event(%v):%s", action.event, err)
		return
	}

	br.bulk_buf = append(br.bulk_buf, meta...)
	br.bulk_buf = append(br.bulk_buf, buf...)
	br.bulk_buf = append(br.bulk_buf, '\n')

	br.actions = append(br.actions, action)
}

func (br *BulkRequest) bufSizeByte() int {
	return len(br.bulk_buf)
}
func (br *BulkRequest) actionCount() int {
	return len(br.actions)
}

type BulkProcessor interface {
	add(*Action)
	bulk(*BulkRequest, int)
	awaityclose(time.Duration)
}

type HTTPBulkProcessor struct {
	bulk_size      int
	bulk_actions   int
	flush_interval int
	concurrent     int
	compress       bool
	execution_id   int
	client         *http.Client
	hostSelector   HostSelector
	bulkRequest    *BulkRequest
	mux            sync.Mutex
	wg             sync.WaitGroup

	semaphore *semaphore.Weighted
}

func NewHTTPBulkProcessor(hosts []string, bulk_size, bulk_actions, flush_interval, concurrent int, compress bool) *HTTPBulkProcessor {
	bulkProcessor := &HTTPBulkProcessor{
		bulk_size:      bulk_size,
		bulk_actions:   bulk_actions,
		flush_interval: flush_interval,
		client:         &http.Client{},
		bulkRequest:    &BulkRequest{},
		hostSelector:   NewRRHostSelector(hosts, 3),
		concurrent:     concurrent,
		compress:       compress,
	}
	bulkProcessor.semaphore = semaphore.NewWeighted(int64(concurrent))

	ticker := time.NewTicker(time.Second * time.Duration(flush_interval))
	go func() {
		for range ticker.C {
			bulkProcessor.semaphore.Acquire(context.TODO(), 1)
			bulkProcessor.mux.Lock()
			if bulkProcessor.bulkRequest.actionCount() == 0 {
				bulkProcessor.mux.Unlock()
				bulkProcessor.semaphore.Release(1)
				continue
			}
			bulkRequest := bulkProcessor.bulkRequest
			bulkProcessor.bulkRequest = &BulkRequest{}
			bulkProcessor.execution_id++
			execution_id := bulkProcessor.execution_id
			bulkProcessor.mux.Unlock()
			bulkProcessor.bulk(bulkRequest, execution_id)
		}
	}()

	return bulkProcessor
}

func (p *HTTPBulkProcessor) add(action *Action) {
	p.bulkRequest.add(action)

	// TODO bulkRequest passed to bulk may be empty, but execution_id has ++
	if p.bulkRequest.bufSizeByte() >= p.bulk_size || p.bulkRequest.actionCount() >= p.bulk_actions {
		p.semaphore.Acquire(context.TODO(), 1)
		p.mux.Lock()
		bulkRequest := p.bulkRequest
		p.bulkRequest = &BulkRequest{}
		p.execution_id++
		execution_id := p.execution_id
		p.mux.Unlock()
		go p.bulk(bulkRequest, execution_id)
	}
}

// TODO: timeout implement
func (p *HTTPBulkProcessor) awaityclose(timeout time.Duration) {
	c := make(chan bool)
	defer func() {
		select {
		case <-c:
			glog.Info("all bulk job done. return")
			return
		case <-time.After(timeout):
			glog.Info("await timeout. return")
			return
		}
	}()

	p.mux.Lock()
	if len(p.bulkRequest.actions) == 0 {
		return
	}
	bulkRequest := p.bulkRequest
	p.bulkRequest = &BulkRequest{}
	p.execution_id++
	execution_id := p.execution_id
	p.mux.Unlock()

	go func() {
		p.innerBulk(bulkRequest, execution_id)
		p.wg.Wait()
		c <- true
	}()

}

func (p *HTTPBulkProcessor) bulk(bulkRequest *BulkRequest, execution_id int) {
	defer p.wg.Done()
	defer p.semaphore.Release(1)
	p.wg.Add(1)
	if bulkRequest.actionCount() == 0 {
		return
	}
	p.innerBulk(bulkRequest, execution_id)
}

func (p *HTTPBulkProcessor) innerBulk(bulkRequest *BulkRequest, execution_id int) {
	glog.Infof("bulk %d docs with execution_id %d", bulkRequest.actionCount(), execution_id)
	for {
		host := p.hostSelector.selectOneHost()
		if host == "" {
			glog.Info("no available host, wait for 30s")
			time.Sleep(30 * time.Second)
			continue
		}

		glog.Infof("try to bulk with host (%s)", host)

		url := host + "/_bulk"
		success, shouldRetry, noRetry := p.tryOneBulk(url, bulkRequest)
		if success {
			glog.Infof("bulk done with execution_id %d", execution_id)
			p.hostSelector.addWeight(host)
		} else {
			glog.Errorf("bulk failed using %s", url)
			p.hostSelector.reduceWeight(host)
			continue
		}

		if len(shouldRetry) > 0 || len(noRetry) > 0 {
			glog.Infof("%d should retry; %d need not retry", len(shouldRetry), len(noRetry))
		}

		if len(noRetry) > 0 {
			b, err := json.Marshal(bulkRequest.actions[noRetry[0]])
			if err != nil {
				glog.Infof("one failed doc that need no retry: %v", bulkRequest.actions[noRetry[0]])
			} else {
				glog.Infof("one failed doc that need no retry: %s", b)
			}
		}

		if len(shouldRetry) > 0 {
			newBulkRequest := &BulkRequest{}
			for _, i := range shouldRetry {
				newBulkRequest.add(bulkRequest.actions[i])
			}
			p.mux.Lock()
			p.execution_id++
			execution_id := p.execution_id
			p.mux.Unlock()
			p.innerBulk(newBulkRequest, execution_id)
		}

		return // only success will go to here
	}
}

// TODO custom status code?
func (p *HTTPBulkProcessor) abstraceBulkResponseItemsByStatus(bulkResponse map[string]interface{}) ([]int, []int) {
	glog.V(20).Infof("%v", bulkResponse)

	retry := make([]int, 0)
	noRetry := make([]int, 0)

	if bulkResponse["errors"] == nil {
		glog.Infof("could NOT get errors in response:%v", bulkResponse)
		return retry, noRetry
	}

	if bulkResponse["errors"].(bool) == false {
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

func (p *HTTPBulkProcessor) tryOneBulk(url string, br *BulkRequest) (bool, []int, []int) {
	glog.V(5).Infof("request size:%d", len(br.bulk_buf))
	glog.V(20).Infof("%s", br.bulk_buf)

	var (
		shouldRetry = make([]int, 0)
		noRetry     = make([]int, 0)
		err         error
		req         *http.Request
	)

	if p.compress {
		var buf bytes.Buffer
		g := gzip.NewWriter(&buf)
		if _, err = g.Write(br.bulk_buf); err != nil {
			glog.Errorf("gzip bulk buf error: %s", err)
			return false, shouldRetry, noRetry
		}
		if err = g.Close(); err != nil {
			glog.Errorf("gzip bulk buf error: %s", err)
			return false, shouldRetry, noRetry
		}
		req, err = http.NewRequest("POST", url, &buf)
		req.Header.Set("Content-Type", "application/x-ndjson")
		req.Header.Set("Content-Encoding", "gzip")
	} else {
		req, err = http.NewRequest("POST", url, bytes.NewBuffer(br.bulk_buf))
		req.Header.Set("Content-Type", "application/x-ndjson")
	}

	resp, err := p.client.Do(req)

	br.bulk_buf = nil

	if err != nil {
		glog.Infof("could not bulk with %s:%s", url, err)
		return false, shouldRetry, noRetry
	}
	switch resp.StatusCode {
	case 502:
		return false, shouldRetry, noRetry
	case 401:
		return false, shouldRetry, noRetry
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf(`read bulk response error:"%s". will NOT retry`, err)
		return true, shouldRetry, noRetry
	}
	glog.V(5).Infof("get response[%d]", len(respBody))
	glog.V(20).Infof("%s", respBody)

	err = resp.Body.Close()
	if err != nil {
		glog.Errorf("close response body error:%s", err)
		return true, shouldRetry, noRetry
	}

	var responseI interface{}
	err = json.Unmarshal(respBody, &responseI)

	if err != nil {
		glog.Errorf(`could not unmarshal bulk response:"%s". will NOT retry. %s`, err, string(respBody[:100]))
		return true, shouldRetry, noRetry
	}

	bulkResponse := responseI.(map[string]interface{})
	shouldRetry, noRetry = p.abstraceBulkResponseItemsByStatus(bulkResponse)
	return true, shouldRetry, noRetry
}

type ElasticsearchOutput struct {
	BaseOutput
	config map[interface{}]interface{}

	index      value_render.ValueRender
	index_type value_render.ValueRender
	id         value_render.ValueRender
	routing    value_render.ValueRender

	bulkProcessor BulkProcessor
}

func NewElasticsearchOutput(config map[interface{}]interface{}) *ElasticsearchOutput {
	rst := &ElasticsearchOutput{
		BaseOutput: NewBaseOutput(config),
		config:     config,
	}

	if v, ok := config["index"]; ok {
		rst.index = value_render.GetValueRender(v.(string))
	} else {
		glog.Fatal("index must be set in elasticsearch output")
	}

	if v, ok := config["index_type"]; ok {
		rst.index_type = value_render.GetValueRender(v.(string))
	} else {
		rst.index_type = value_render.GetValueRender(DEFAULT_INDEX_TYPE)
	}

	if v, ok := config["id"]; ok {
		rst.id = value_render.GetValueRender(v.(string))
	} else {
		rst.id = nil
	}

	if v, ok := config["routing"]; ok {
		rst.routing = value_render.GetValueRender(v.(string))
	} else {
		rst.routing = nil
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

	rst.bulkProcessor = NewHTTPBulkProcessor(hosts, bulk_size, bulk_actions, flush_interval, concurrent, compress)
	return rst
}

func (p *ElasticsearchOutput) Emit(event map[string]interface{}) {
	var (
		index      string = p.index.Render(event).(string)
		index_type string = p.index_type.Render(event).(string)
		op         string = "index"
		id         string
		routing    string
	)
	if p.id == nil {
		id = ""
	} else {
		t := p.id.Render(event)
		if t == nil {
			id = ""
			glog.V(20).Infof("could not render id:%s", event)
		} else {
			id = t.(string)
		}
	}

	if p.routing == nil {
		routing = ""
	} else {
		t := p.routing.Render(event)
		if t == nil {
			routing = ""
			glog.V(20).Infof("could not render routing:%s", event)
		} else {
			routing = t.(string)
		}
	}
	p.bulkProcessor.add(&Action{op, index, index_type, id, routing, event})
}
func (outputPlugin *ElasticsearchOutput) Shutdown() {
	outputPlugin.bulkProcessor.awaityclose(5 * time.Second)
}
