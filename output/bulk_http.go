package output

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/golang/glog"
)

const (
	DEFAULT_BULK_SIZE      = 15 * 1024 * 1024
	DEFAULT_BULK_ACTIONS   = 5000
	DEFAULT_FLUSH_INTERVAL = 30
	DEFAULT_CONCURRENT     = 1

	MAX_BYTE_SIZE_APPLIED_IN_ADVANCE = 1024 * 1024 * 50
)

var (
	REMOVE_HTTP_AUTH_REGEXP = regexp.MustCompile(`^(?i)(http(s?)://)[^:]+:[^@]+@`)
)

type Event interface {
	Encode() []byte
}

type BulkRequest interface {
	add(Event)
	bufSizeByte() int
	eventCount() int
	readBuf() []byte
}
type NewBulkRequestFunc func() BulkRequest

type BulkProcessor interface {
	add(Event)
	bulk()
	awaitclose(time.Duration)
}

type GetRetryEventsFunc func(*http.Response, []byte, *BulkRequest) ([]int, []int, BulkRequest)

type HTTPBulkProcessor struct {
	stop              bool
	headers           map[string]string
	requestMethod     string
	retryResponseCode map[int]bool
	bulk_size         int
	bulk_actions      int
	flush_interval    int
	concurrent        int
	compress          bool
	execution_id      int
	client            *http.Client
	mux               sync.Mutex
	executionIDmux    sync.Mutex
	wg                sync.WaitGroup

	bulkChan chan *BulkRequest

	hostSelector       HostSelector
	bulkRequest        BulkRequest
	newBulkRequestFunc NewBulkRequestFunc
	getRetryEventsFunc GetRetryEventsFunc
}

func NewHTTPBulkProcessor(headers map[string]string, hosts []string, requestMethod string, retryResponseCode map[int]bool, bulk_size, bulk_actions, flush_interval, concurrent int, compress bool, newBulkRequestFunc NewBulkRequestFunc, getRetryEventsFunc GetRetryEventsFunc) *HTTPBulkProcessor {
	hostsI := make([]interface{}, len(hosts))
	for i, h := range hosts {
		hostsI[i] = h
	}
	bulkProcessor := &HTTPBulkProcessor{
		headers:            headers,
		requestMethod:      requestMethod,
		retryResponseCode:  retryResponseCode,
		bulk_size:          bulk_size,
		bulk_actions:       bulk_actions,
		flush_interval:     flush_interval,
		client:             &http.Client{},
		hostSelector:       NewRRHostSelector(hostsI, 3),
		concurrent:         concurrent,
		compress:           compress,
		bulkRequest:        newBulkRequestFunc(),
		newBulkRequestFunc: newBulkRequestFunc,
		getRetryEventsFunc: getRetryEventsFunc,
		bulkChan:           make(chan *BulkRequest, concurrent),
	}
	for i := 0; i < concurrent; i++ {
		go func() {
			for !bulkProcessor.stop {
				bulkRequest := <-bulkProcessor.bulkChan
				if (*bulkRequest).eventCount() <= 0 {
					continue
				}
				bulkProcessor.wg.Add(1)
				bulkProcessor.innerBulk(bulkRequest)
				bulkProcessor.wg.Done()
			}
		}()
	}

	ticker := time.NewTicker(time.Second * time.Duration(flush_interval))
	go func() {
		for range ticker.C {
			bulkProcessor.bulk()
		}
	}()

	return bulkProcessor
}

func (p *HTTPBulkProcessor) add(event Event) {
	p.mux.Lock()
	p.bulkRequest.add(event)
	p.mux.Unlock()

	if p.bulkRequest.bufSizeByte() >= p.bulk_size || p.bulkRequest.eventCount() >= p.bulk_actions {
		p.bulk()
	}
}

func (p *HTTPBulkProcessor) bulk() {
	p.mux.Lock()
	bulkRequest := p.bulkRequest
	p.bulkRequest = p.newBulkRequestFunc()
	p.mux.Unlock()

	p.bulkChan <- &bulkRequest
}

func (p *HTTPBulkProcessor) innerBulk(bulkRequest *BulkRequest) {
	p.executionIDmux.Lock()
	p.execution_id++
	execution_id := p.execution_id
	p.executionIDmux.Unlock()

	_startTime := float64(time.Now().UnixNano()/1000000) / 1000
	eventCount := (*bulkRequest).eventCount()
	glog.Infof("bulk %d docs with execution_id %d", eventCount, p.execution_id)
	for {
		nexthost := p.hostSelector.Next()
		if nexthost == nil {
			glog.Info("no available host, wait for 30s")
			time.Sleep(30 * time.Second)
			continue
		}
		host := nexthost.(string)

		glog.Infof("try to bulk with host (%s)", REMOVE_HTTP_AUTH_REGEXP.ReplaceAllString(host, "${1}"))

		url := host
		success, shouldRetry, noRetry, newBulkRequest := p.tryOneBulk(url, bulkRequest)
		if success {
			_finishTime := float64(time.Now().UnixNano()/1000000) / 1000
			timeTaken := _finishTime - _startTime
			glog.Infof("bulk done with execution_id %d %.3f %d %.3f", execution_id, timeTaken, eventCount, float64(eventCount)/timeTaken)
			p.hostSelector.AddWeight()
		} else {
			glog.Errorf("bulk failed with %s", url)
			p.hostSelector.ReduceWeight()
			continue
		}

		if len(shouldRetry) > 0 || len(noRetry) > 0 {
			glog.Infof("%d should retry; %d need not retry", len(shouldRetry), len(noRetry))
		}

		if len(shouldRetry) > 0 {
			p.innerBulk(&newBulkRequest)
		}

		return // only success will go to here
	}
}

func (p *HTTPBulkProcessor) tryOneBulk(url string, br *BulkRequest) (bool, []int, []int, BulkRequest) {
	glog.V(5).Infof("request size:%d", (*br).bufSizeByte())
	glog.V(20).Infof("%s", (*br).readBuf())

	var (
		shouldRetry    = make([]int, 0)
		noRetry        = make([]int, 0)
		err            error
		req            *http.Request
		newBulkRequest BulkRequest
	)

	if p.compress {
		var buf bytes.Buffer
		g := gzip.NewWriter(&buf)
		if _, err = g.Write((*br).readBuf()); err != nil {
			glog.Errorf("gzip bulk buf error: %s", err)
			return false, shouldRetry, noRetry, nil
		}
		if err = g.Close(); err != nil {
			glog.Errorf("gzip bulk buf error: %s", err)
			return false, shouldRetry, noRetry, nil
		}
		req, err = http.NewRequest(p.requestMethod, url, &buf)
		if err != nil {
			glog.Errorf("create request error: %s", err)
			return false, shouldRetry, noRetry, nil
		} else {
			req.Header.Set("Content-Encoding", "gzip")
		}
	} else {
		req, err = http.NewRequest(p.requestMethod, url, bytes.NewBuffer((*br).readBuf()))
	}

	if err != nil {
		glog.Errorf("create request error: %s", err)
		return false, shouldRetry, noRetry, nil
	}

	for k, v := range p.headers {
		req.Header.Set(k, v)
	}

	resp, err := p.client.Do(req)

	if err != nil {
		glog.Infof("request with %s error: %s", url, err)
		return false, shouldRetry, noRetry, nil
	}

	defer resp.Body.Close()

	if p.retryResponseCode[resp.StatusCode] {
		return false, shouldRetry, noRetry, nil
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf(`read bulk response error: %s. will NOT retry`, err)
		return true, shouldRetry, noRetry, nil
	}
	glog.V(5).Infof("get response[%d]", len(respBody))
	glog.V(20).Infof("%s", respBody)

	shouldRetry, noRetry, newBulkRequest = p.getRetryEventsFunc(resp, respBody, br)

	return true, shouldRetry, noRetry, newBulkRequest
}

func (p *HTTPBulkProcessor) awaitclose(timeout time.Duration) {
	p.stop = true
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

	defer func() {
		go func() {
			p.wg.Wait()
			c <- true
		}()
	}()

AllBulkReqInChan:
	for {
		select {
		case bulkRequest := <-p.bulkChan:
			if (*bulkRequest).eventCount() <= 0 {
				continue
			}
			p.wg.Add(1)
			go func() {
				glog.Infof("bulk %d docs from bulkChan in awaitclose", (*bulkRequest).eventCount())
				p.innerBulk(bulkRequest)
				p.wg.Done()
			}()
		default:
			break AllBulkReqInChan
		}
	}

	p.mux.Lock()
	defer p.mux.Unlock()
	if p.bulkRequest.eventCount() == 0 {
		return
	}
	bulkRequest := p.bulkRequest
	p.bulkRequest = p.newBulkRequestFunc()

	p.wg.Add(1)
	go func() {
		glog.Infof("bulk last %d docs in awaitclose", bulkRequest.eventCount())
		p.innerBulk(&bulkRequest)
		p.wg.Done()
	}()
}
