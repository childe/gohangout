package output

import (
	"encoding/json"
	"fmt"
	"github.com/childe/gohangout/topology"
	"github.com/childe/gohangout/value_render"
	"github.com/golang/glog"
	"gopkg.in/resty.v1"
	"strings"
	"sync"
	"time"
)

type ZincOutput struct {
	config map[interface{}]interface{}

	index        value_render.ValueRender
	bulkRequests []string

	bulkSize      int
	flushInterval int

	address  string
	user     string
	password string

	client *resty.Client

	lock sync.Mutex
}

type ZincAction struct {
	Index string
	Body  string
}

func init() {
	Register("Zinc", NewZincOutput)
}

const messageTemplate = `
{ "index" : { "_index" : "%s" } }
%s
`

func NewZincOutput(config map[interface{}]interface{}) topology.Output {
	output := &ZincOutput{config: config}
	if address, ok := config["address"]; ok {
		output.address = address.(string)
	} else {
		glog.Fatal("address must be set in zinc output")
	}
	if username, ok := config["user"]; ok {
		output.user = username.(string)
	} else {
		glog.Fatal("Zinc output config illegal, not specify username")
	}

	if password, ok := config["password"]; ok {
		output.password = password.(string)
	} else {
		glog.Fatal("Zinc output config illegal, not specify password")
	}

	if index, ok := config["index"]; ok {
		output.index = value_render.GetValueRender(index.(string))
	} else {
		glog.Fatal("Zinc output config illegal, not specify index")
	}

	var (
		bulkSize, flushInterval int
	)

	if v, ok := config["bulk_size"]; ok {
		bulkSize = v.(int)
	} else {
		bulkSize = DEFAULT_BULK_SIZE
	}
	output.bulkSize = bulkSize

	if v, ok := config["flush_interval"]; ok {
		flushInterval = v.(int)
	} else {
		flushInterval = DEFAULT_FLUSH_INTERVAL
	}
	output.flushInterval = flushInterval
	output.bulkRequests = make([]string, 0)
	output.client = resty.New()

	ticker := time.NewTicker(time.Second * time.Duration(output.flushInterval))
	go func() {
		for range ticker.C {
			output.flushRequests()
		}
	}()
	return output
}

func (z *ZincOutput) Emit(event map[string]interface{}) {
	index := z.index.Render(event).(string)
	marshal, err := json.Marshal(event)
	if err != nil {
		glog.Errorf("Marshal event failed: %v", err)
		return
	}
	z.lock.Lock()
	z.bulkRequests = append(z.bulkRequests, fmt.Sprintf(messageTemplate, index, string(marshal)))
	z.lock.Unlock()
	if len(z.bulkRequests) >= z.bulkSize {
		z.flushRequests()
	}
}

func (z *ZincOutput) flushRequests() {
	z.lock.Lock()
	requests := z.bulkRequests
	z.bulkRequests = make([]string, 0)
	z.lock.Unlock()
	_, err := z.client.R().
		SetBasicAuth(z.user, z.password).
		SetBody(strings.Join(requests, "\n")).
		Post(fmt.Sprintf("%s/api/_bulk", strings.TrimRight(z.address, "/")))
	if err != nil {
		glog.Errorf("Write messages to zinc failed: %v", err)
	} else {
		glog.Infof("Write messages to zinc success, size: %v", len(requests))
	}
}

func (z *ZincOutput) Shutdown() {
}
