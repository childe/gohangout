package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/yalp/jsonpath"

	"github.com/childe/gohangout/codec"
	"github.com/childe/gohangout/condition_filter"
	"github.com/childe/gohangout/topology"
	"github.com/childe/gohangout/value_render"
	"github.com/golang/glog"
)

const (
	defaultIndexType = "logs"
	defaultEsVersion = 6
	defaultAction    = "index"
)

var (
	f                 func() codec.Encoder
	defaultNormalResp = []byte(`"errors":false,`)

	action string = defaultAction
)

type Action struct {
	op         string
	index      string
	index_type string
	id         string
	routing    string
	event      map[string]interface{}
	rawSource  []byte
	es_version int
}

func (action *Action) Encode() []byte {
	var (
		meta []byte = make([]byte, 0, 1000)
		buf  []byte
		err  error
	)
	meta = append(meta, `{"`+action.op+`":{"_index":`...)
	index, _ := f().Encode(action.index)
	meta = append(meta, index...)

	if action.es_version <= defaultEsVersion {
		meta = append(meta, `,"_type":`...)
		index_type, _ := f().Encode(action.index_type)
		meta = append(meta, index_type...)
	}

	if action.id != "" {
		meta = append(meta, `,"_id":`...)
		doc_id, _ := f().Encode(action.id)
		meta = append(meta, doc_id...)
	}

	meta = append(meta, `,"routing":`...)
	routing, _ := f().Encode(action.routing)
	meta = append(meta, routing...)

	meta = append(meta, "}}\n"...)

	if action.rawSource == nil {
		buf, err = f().Encode(action.event)
		if err != nil {
			glog.Errorf("could marshal event(%v):%s", action.event, err)
			return nil
		}
	} else {
		buf = action.rawSource
	}

	bulk_buf := make([]byte, 0, len(meta)+len(buf)+1)
	bulk_buf = append(bulk_buf, meta...)
	bulk_buf = append(bulk_buf, buf[:]...)
	bulk_buf = append(bulk_buf, '\n')
	return bulk_buf
}

type ESBulkRequest struct {
	events   []Event
	bulk_buf []byte
}

func (br *ESBulkRequest) add(event Event) {
	br.bulk_buf = append(br.bulk_buf, event.Encode()...)
	br.events = append(br.events, event)
}

func (br *ESBulkRequest) bufSizeByte() int {
	return len(br.bulk_buf)
}
func (br *ESBulkRequest) eventCount() int {
	return len(br.events)
}
func (br *ESBulkRequest) readBuf() []byte {
	return br.bulk_buf
}

type ElasticsearchOutput struct {
	config map[interface{}]interface{}

	action             string
	index              value_render.ValueRender
	index_type         value_render.ValueRender
	id                 value_render.ValueRender
	routing            value_render.ValueRender
	source_field       value_render.ValueRender
	bytes_source_field value_render.ValueRender
	es_version         int
	bulkProcessor      BulkProcessor

	scheme   string
	user     string
	password string
	hosts    []string
}

func esGetRetryEvents(resp *http.Response, respBody []byte, bulkRequest *BulkRequest) ([]int, []int, BulkRequest) {
	retry := make([]int, 0)
	noRetry := make([]int, 0)
	//make a string index to avoid json decode for speed up over 90%+ scences
	if bytes.Index(respBody, defaultNormalResp) != -1 {
		return retry, noRetry, nil
	}
	var responseI interface{}
	err := json.Unmarshal(respBody, &responseI)
	if err != nil {
		glog.Errorf(`could not unmarshal bulk response:"%s". will NOT retry. %s`, err, string(respBody[:100]))
		return retry, noRetry, nil
	}

	bulkResponse := responseI.(map[string]interface{})
	glog.V(20).Infof("%v", bulkResponse)

	if bulkResponse["errors"] == nil {
		glog.Infof("could NOT get errors in response:%s", string(respBody))
		return retry, noRetry, nil
	}

	if bulkResponse["errors"].(bool) == false {
		return retry, noRetry, nil
	}

	hasLog := false
	for i, item := range bulkResponse["items"].([]interface{}) {
		index := item.(map[string]interface{})[action].(map[string]interface{})

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
	newbulkRequest := buildRetryBulkRequest(retry, noRetry, bulkRequest)
	return retry, noRetry, newbulkRequest
}

func buildRetryBulkRequest(shouldRetry, noRetry []int, bulkRequest *BulkRequest) BulkRequest {
	esBulkRequest := (*bulkRequest).(*ESBulkRequest)
	if len(noRetry) > 0 {
		b, err := json.Marshal(esBulkRequest.events[noRetry[0]].(*Action).event)
		if err != nil {
			glog.Infof("one failed doc that need no retry: %+v", esBulkRequest.events[noRetry[0]].(*Action).event)
		} else {
			glog.Infof("one failed doc that need no retry: %s", b)
		}
	}

	if len(shouldRetry) > 0 {
		newBulkRequest := &ESBulkRequest{
			bulk_buf: make([]byte, 0),
		}
		for _, i := range shouldRetry {
			newBulkRequest.add(esBulkRequest.events[i])
		}
		return newBulkRequest
	}
	return nil
}

func init() {
	Register("Elasticsearch", newElasticsearchOutput)
}

func newElasticsearchOutput(config map[interface{}]interface{}) topology.Output {
	rst := &ElasticsearchOutput{
		config: config,
	}

	_codec := "simplejson"
	if v, ok := config["codec"]; ok {
		_codec = v.(string)
	}
	f = func() codec.Encoder { return codec.NewEncoder(_codec) }

	if v, ok := config["action"]; ok {
		rst.action = v.(string)
	} else {
		rst.action = defaultAction
	}
	action = rst.action

	if v, ok := config["index"]; ok {
		rst.index = value_render.GetValueRender(v.(string))
	} else {
		glog.Fatal("index must be set in elasticsearch output")
	}

	if v, ok := config["index_time_location"]; ok {
		if e, ok := rst.index.(*value_render.IndexRender); ok {
			e.SetTimeLocation(v.(string))
		} else {
			glog.Fatal("index_time_location is not supported in this index format")
		}
	}

	if v, ok := config["index_type"]; ok {
		rst.index_type = value_render.GetValueRender(v.(string))
	} else {
		rst.index_type = value_render.GetValueRender(defaultIndexType)
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

	if v, ok := config["source_field"]; ok {
		rst.source_field = value_render.GetValueRender2(v.(string))
	} else {
		rst.source_field = nil
	}

	if v, ok := config["bytes_source_field"]; ok {
		rst.bytes_source_field = value_render.GetValueRender2(v.(string))
	} else {
		rst.bytes_source_field = nil
	}

	if v, ok := config["es_version"]; ok {
		rst.es_version = v.(int)
	} else {
		rst.es_version = defaultEsVersion
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

	var headers = map[string]string{"Content-Type": "application/x-ndjson"}
	if v, ok := config["headers"]; ok {
		for keyI, valueI := range v.(map[interface{}]interface{}) {
			headers[keyI.(string)] = valueI.(string)
		}
	}
	var requestMethod string = "POST"

	var retryResponseCode map[int]bool = make(map[int]bool)
	if v, ok := config["retry_response_code"]; ok {
		for _, cI := range v.([]interface{}) {
			retryResponseCode[cI.(int)] = true
		}
	} else {
		retryResponseCode[401] = true
		retryResponseCode[502] = true
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

	var hosts []string = make([]string, 0)
	if v, ok := config["hosts"]; ok {
		for _, h := range v.([]interface{}) {
			scheme, user, password, host := getUserPasswordAndHost(h.(string))
			if host == "" {
				glog.Fatalf("invalid host: %q", host)
			}
			rst.scheme = scheme
			rst.user = user
			rst.password = password
			hosts = append(hosts, host)
		}
	} else {
		glog.Fatal("hosts must be set in elasticsearch output")
	}
	rst.hosts = hosts

	var err error
	if sniff, ok := config["sniff"]; ok {
		glog.Infof("sniff hosts in es cluster")
		sniff := sniff.(map[interface{}]interface{})
		hosts, err = sniffNodes(config)
		glog.Infof("new hosts after sniff: %v", hosts)
		if err != nil {
			glog.Fatalf("could not sniff hosts: %v", err)
		}
		if len(hosts) == 0 {
			glog.Fatal("no available hosts after sniff")
		}
		rst.hosts = hosts

		refreshInterval := sniff["refresh_interval"].(int)
		if refreshInterval > 0 {
			go func() {
				for range time.NewTicker(time.Second * time.Duration(refreshInterval)).C {
					hosts, err = sniffNodes(config)
					if err != nil {
						glog.Errorf("could not sniff hosts: %v", err)
					} else {
						if !reflect.DeepEqual(rst.hosts, hosts) {
							glog.Infof("new hosts after sniff: %v", hosts)
							rst.hosts = hosts
							rst.bulkProcessor.(*HTTPBulkProcessor).resetHosts(rst.assebleHosts())
						}
					}
				}
			}()
		}
	}
	rst.bulkProcessor = NewHTTPBulkProcessor(headers, rst.assebleHosts(), requestMethod, retryResponseCode, bulk_size, bulk_actions, flush_interval, concurrent, compress, f, esGetRetryEvents)
	return rst
}

func getUserPasswordAndHost(url string) (scheme, user, password, host string) {
	p := regexp.MustCompile(`^(?i)(?:(http(?:s?)))://(?:([^:]+):([^@]+)@)?(\S+)$`)
	r := p.FindStringSubmatch(url)
	if len(r) == 5 {
		scheme = r[1]
		user = r[2]
		password = r[3]
		host = strings.TrimRight(r[4], "/")
		return
	} else if len(r) == 3 {
		host = strings.TrimRight(r[2], "/")
		return
	} else {
		glog.Infof("%q is invalid host format", host)
		return
	}
}

func sniffNodes(config map[interface{}]interface{}) ([]string, error) {
	sniff := config["sniff"].(map[interface{}]interface{})
	var (
		match string
		ok    bool
	)
	v, _ := sniff["match"]
	if v != nil {
		match, ok = v.(string)
		if !ok {
			glog.Fatal("match in sniff settings must be string")
		}
	}
	for _, host := range config["hosts"].([]interface{}) {
		host := host.(string)
		if nodes, err := sniffNodesFromOneHost(host, match); err == nil {
			return nodes, err
		} else {
			glog.Errorf("sniff nodes error from %s: %v", REMOVE_HTTP_AUTH_REGEXP.ReplaceAllString(host, "${1}"), err)
		}
	}
	return nil, errors.New("sniff nodes error from all hosts")
}

func sniffNodesFromOneHost(host string, match string) ([]string, error) {
	url := strings.TrimRight(host, "/") + "/_nodes/_all/http"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("es sniff error %s:%d", url, resp.StatusCode)
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	v := make(map[string]interface{})
	err = json.Unmarshal(respBody, &v)
	if err != nil {
		return nil, err
	}

	glog.Infof("sniff resp: %v", v)
	return filterNodesIPList(v, match)
}

// filterNodesIPList gets ip lists from what is returned from _nodes/_all/info
// it uses `match` config to filter the nodes you what
func filterNodesIPList(v map[string]interface{}, match string) ([]string, error) {
	var nodes map[string]interface{}
	if nodesV, ok := v["nodes"]; ok {
		if nodesV, ok := nodesV.(map[string]interface{}); ok {
			nodes = nodesV
		} else {
			return nil, errors.New("es sniff error: `nodes` is not map")
		}
	} else {
		return nil, errors.New("es sniff error: `nodes` not exist")
	}

	var f condition_filter.Condition
	if match != "" {
		f = condition_filter.NewCondition(match)
	}
	IPList := make([]string, 0)
	for _, node := range nodes {
		if node, ok := node.(map[string]interface{}); ok {
			if f == nil || f.Pass(node) {
				if ip, err := jsonpath.Read(node, "$.http.publish_address"); err == nil {
					if ip, ok := ip.(string); ok {
						IPList = append(IPList, ip)
					}
				} else {
					return nil, err
				}
			}
		}
	}
	return IPList, nil
}

// create ES host list using scheme, [user, password] and hosts
func (p *ElasticsearchOutput) assebleHosts() (hosts []string) {
	hosts = make([]string, 0)
	for _, host := range p.hosts {
		if len(p.user) > 0 {
			hosts = append(hosts, fmt.Sprintf("%s://%s:%s@%s", p.scheme, p.user, p.password, host))
		} else {
			hosts = append(hosts, fmt.Sprintf("%s://%s", p.scheme, host))
		}
	}
	return
}

// Emit adds the event to bulkProcessor
func (p *ElasticsearchOutput) Emit(event map[string]interface{}) {
	var (
		index      string = p.index.Render(event).(string)
		index_type string = p.index_type.Render(event).(string)
		op         string = p.action
		es_version int    = p.es_version
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

	if p.source_field == nil && p.bytes_source_field == nil {
		p.bulkProcessor.add(&Action{op, index, index_type, id, routing, event, nil, es_version})
	} else if p.bytes_source_field != nil {
		t := p.bytes_source_field.Render(event)
		if t == nil {
			p.bulkProcessor.add(&Action{op, index, index_type, id, routing, event, nil, es_version})
		} else {
			p.bulkProcessor.add(&Action{op, index, index_type, id, routing, event, (t.([]byte)), es_version})
		}
	} else {
		t := p.source_field.Render(event)
		if t == nil {
			p.bulkProcessor.add(&Action{op, index, index_type, id, routing, event, nil, es_version})
		} else {
			p.bulkProcessor.add(&Action{op, index, index_type, id, routing, event, []byte(t.(string)), es_version})
		}
	}
}

func (outputPlugin *ElasticsearchOutput) Shutdown() {
	outputPlugin.bulkProcessor.awaitclose(30 * time.Second)
}
