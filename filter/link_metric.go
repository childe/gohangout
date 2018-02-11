package filter

import (
	"reflect"
	"strings"
	"time"

	"github.com/golang-collections/collections/stack"
	"github.com/golang/glog"
)

type LinkMetricFilter struct {
	BaseFilter

	config        map[interface{}]interface{}
	timestamp     string
	batchWindow   int64
	reserveWindow int64
	overwrite     bool

	fields []string

	metric       map[int64]interface{}
	metricToEmit map[int64]interface{}
}

func NewLinkMetricFilter(config map[interface{}]interface{}) *LinkMetricFilter {
	plugin := &LinkMetricFilter{
		BaseFilter:   NewBaseFilter(config),
		config:       config,
		overwrite:    true,
		metric:       make(map[int64]interface{}),
		metricToEmit: make(map[int64]interface{}),
	}

	if overwrite, ok := config["overwrite"]; ok {
		plugin.overwrite = overwrite.(bool)
	}

	if fieldsLink, ok := config["fieldsLink"]; ok {
		plugin.fields = strings.Split(fieldsLink.(string), "->")
	} else {
		glog.Fatal("fieldsLink must be set in linkmetric filter plugin")
	}

	if timestamp, ok := config["timestamp"]; ok {
		plugin.timestamp = timestamp.(string)
	} else {
		plugin.timestamp = "@timestamp"
	}

	if batchWindow, ok := config["batchWindow"]; ok {
		plugin.batchWindow = int64(batchWindow.(int))
	} else {
		glog.Fatal("batchWindow must be set in linkmetric filter plugin")
	}

	if reserveWindow, ok := config["reserveWindow"]; ok {
		plugin.reserveWindow = int64(reserveWindow.(int)) * 1000
	} else {
		glog.Fatal("reserveWindow must be set in linkmetric filter plugin")
	}

	ticker := time.NewTicker(time.Second * time.Duration(plugin.batchWindow))
	go func() {
		for range ticker.C {
			if len(plugin.metric) > 0 && len(plugin.metricToEmit) == 0 {
				plugin.metricToEmit = plugin.metric
				plugin.metric = make(map[int64]interface{})
			}
		}
	}()
	return plugin
}

func (plugin *LinkMetricFilter) updateMetric(event map[string]interface{}) {
	var timestamp int64
	if v, ok := event[plugin.timestamp]; ok {
		if reflect.TypeOf(v).String() != "time.Time" {
			glog.V(10).Infof("timestamp must be time.Time, but it's %s", reflect.TypeOf(v).String())
			return
		}
		timestamp = v.(time.Time).Unix()
	} else {
		glog.V(10).Infof("not timestamp in event. %s", event)
		return
	}

	diff := time.Now().Unix() - timestamp
	if diff > plugin.reserveWindow || diff < 0 {
		return
	}

	timestamp -= timestamp % plugin.batchWindow
	var set map[string]interface{} = nil
	if v, ok := plugin.metric[timestamp]; ok {
		set = v.(map[string]interface{})
	} else {
		set = make(map[string]interface{})
		plugin.metric[timestamp] = set
	}

	var fieldValue string
	for _, field := range plugin.fields {
		fieldValueI := event[field]
		if fieldValueI == nil {
			return
		}
		fieldValue = fieldValueI.(string)
		if v, ok := set[fieldValue]; ok {
			set = v.(map[string]interface{})
		} else {
			set[fieldValue] = make(map[string]interface{})
			set = set[fieldValue].(map[string]interface{})
		}
	}

	if count, ok := set["count"]; ok {
		set["count"] = 1 + count.(int)
	} else {
		set["count"] = 1
	}
}

func (plugin *LinkMetricFilter) Process(event map[string]interface{}) (map[string]interface{}, bool) {
	plugin.updateMetric(event)
	return event, true
}

func (plugin *LinkMetricFilter) EmitExtraEvents(sTo *stack.Stack) []map[string]interface{} {
	/*
	   if (metricToEmit.size() == 0) {
	       return null;
	   }
	   List<Map<String, Object>> events = new ArrayList();

	   this.metricToEmit.forEach((timestamp, s) -> {
	       this.metricToEvents((Map) s, 0).forEach((Map<String, Object> e) -> {
	           e.put(this.timestamp, timestamp);
	           this.postProcess(e, true);
	           events.add(e);
	       });
	   });

	   this.metricToEmit.clear();
	   this.lastEmitTime = System.currentTimeMillis();

	   return events;
	*/
	if len(plugin.metricToEmit) == 0 {
		return nil
	}
	for timestamp, sI := range plugin.metricToEmit {
		s := sI.(map[string]interface{})
		s["@timestamp"] = timestamp
		sTo.Push(s)
	}
	return nil
}
