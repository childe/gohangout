package filter

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
)

type LinkMetricFilter struct {
	config            map[interface{}]interface{}
	timestamp         string
	batchWindow       int64
	reserveWindow     int64
	overwrite         bool
	dropOriginalEvent bool
	windowOffset      int64
	accumulateMode    int

	fields            []string
	fieldsWithoutLast []string
	lastField         string
	fieldsLength      int

	metric       map[int64]interface{}
	metricToEmit map[int64]interface{}

	mutex sync.Locker
}

func NewLinkMetricFilter(config map[interface{}]interface{}) *LinkMetricFilter {
	plugin := &LinkMetricFilter{
		config:       config,
		overwrite:    true,
		metric:       make(map[int64]interface{}),
		metricToEmit: make(map[int64]interface{}),
		mutex:        &sync.Mutex{},
	}

	if overwrite, ok := config["overwrite"]; ok {
		plugin.overwrite = overwrite.(bool)
	}

	if fieldsLink, ok := config["fieldsLink"]; ok {
		plugin.fields = strings.Split(fieldsLink.(string), "->")
		plugin.fieldsLength = len(plugin.fields)
		plugin.fieldsWithoutLast = plugin.fields[:plugin.fieldsLength-1]
		plugin.lastField = plugin.fields[plugin.fieldsLength-1]
	} else {
		glog.Fatal("fieldsLink must be set in linkmetric filter plugin")
	}

	if timestamp, ok := config["timestamp"]; ok {
		plugin.timestamp = timestamp.(string)
	} else {
		plugin.timestamp = "@timestamp"
	}

	if dropOriginalEvent, ok := config["drop_original_event"]; ok {
		plugin.dropOriginalEvent = dropOriginalEvent.(bool)
	} else {
		plugin.dropOriginalEvent = false
	}

	if batchWindow, ok := config["batchWindow"]; ok {
		plugin.batchWindow = int64(batchWindow.(int))
	} else {
		glog.Fatal("batchWindow must be set in linkmetric filter plugin")
	}

	if reserveWindow, ok := config["reserveWindow"]; ok {
		plugin.reserveWindow = int64(reserveWindow.(int))
	} else {
		glog.Fatal("reserveWindow must be set in linkmetric filter plugin")
	}

	if accumulateModeI, ok := config["accumulateMode"]; ok {
		accumulateMode := accumulateModeI.(string)
		switch accumulateMode {
		case "cumulative":
			plugin.accumulateMode = 0
		case "separate":
			plugin.accumulateMode = 1
		default:
			glog.Errorf("invalid accumulateMode: %s. set to cumulative", accumulateMode)
			plugin.accumulateMode = 0
		}
	} else {
		plugin.accumulateMode = 0
	}

	if windowOffset, ok := config["windowOffset"]; ok {
		plugin.windowOffset = (int64)(windowOffset.(int))
	} else {
		plugin.windowOffset = 0
	}

	ticker := time.NewTicker(time.Second * time.Duration(plugin.batchWindow))
	go func() {
		for range ticker.C {
			plugin.swap_Metric_MetricToEmit()
		}
	}()
	return plugin
}

func (f *LinkMetricFilter) swap_Metric_MetricToEmit() {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.metric) > 0 && len(f.metricToEmit) == 0 {
		timestamp := time.Now().Unix()
		timestamp -= timestamp % f.batchWindow

		f.metricToEmit = make(map[int64]interface{})
		for k, v := range f.metric {
			if k <= timestamp-f.batchWindow*f.windowOffset {
				f.metricToEmit[k] = v
			}
		}

		if f.accumulateMode == 1 {
			f.metric = make(map[int64]interface{})
		} else {
			newMetric := make(map[int64]interface{})
			for k, v := range f.metric {
				if k >= timestamp-f.reserveWindow {
					newMetric[k] = v
				}
			}
			f.metric = newMetric
		}
	}
}

func (f *LinkMetricFilter) updateMetric(event map[string]interface{}) {
	lastFieldValue := event[f.lastField]
	if lastFieldValue == nil {
		return
	}

	var timestamp int64
	if v, ok := event[f.timestamp]; ok {
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
	if diff > f.reserveWindow || diff < 0 {
		return
	}

	timestamp -= timestamp % f.batchWindow
	var set map[interface{}]interface{} = nil
	if v, ok := f.metric[timestamp]; ok {
		set = v.(map[interface{}]interface{})
	} else {
		set = make(map[interface{}]interface{})
		f.metric[timestamp] = set
	}

	for _, field := range f.fieldsWithoutLast {
		fieldValue := event[field]
		if fieldValue == nil {
			return
		}
		if v, ok := set[fieldValue]; ok {
			set = v.(map[interface{}]interface{})
		} else {
			set[fieldValue] = make(map[interface{}]interface{})
			set = set[fieldValue].(map[interface{}]interface{})
		}
	}

	if count, ok := set[lastFieldValue]; ok {
		set[lastFieldValue] = 1 + count.(int)
	} else {
		set[lastFieldValue] = 1
	}

	f.EmitMetrics()
}

func (f *LinkMetricFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	f.updateMetric(event)
	if f.dropOriginalEvent {
		return nil, false
	}
	return event, false
}

func (f *LinkMetricFilter) metricToEvents(metrics map[interface{}]interface{}, level int) []map[string]interface{} {
	var (
		fieldName string                   = f.fields[level]
		events    []map[string]interface{} = make([]map[string]interface{}, 0)
	)

	if level == f.fieldsLength-1 {
		for fieldValue, count := range metrics {
			event := make(map[string]interface{})
			event[fmt.Sprintf("%s", fieldName)] = fieldValue
			event["count"] = count
			events = append(events, event)
		}
		return events
	}

	for fieldValue, nextLevelMetrics := range metrics {
		for _, e := range f.metricToEvents(nextLevelMetrics.(map[interface{}]interface{}), level+1) {
			event := make(map[string]interface{})
			event[fmt.Sprintf("%s", fieldName)] = fieldValue
			for k, v := range e {
				event[k] = v
			}
			events = append(events, event)
		}
	}

	return events
}

func (f *LinkMetricFilter) EmitMetrics() {
	if len(f.metricToEmit) == 0 {
		return
	}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	//var event map[string]interface{}
	//for timestamp, metrics := range f.metricToEmit {
	//for _, event = range f.metricToEvents(metrics.(map[interface{}]interface{}), 0) {
	//event[f.timestamp] = time.Unix(timestamp, 0)
	//event = f.PostProcess(event, true)

	//if f.BaseFilter.nextFilter != nil {
	//f.BaseFilter.nextFilter.Process(event)
	//} else {
	//for _, outputPlugin := range f.BaseFilter.outputs {
	//if outputPlugin.Pass(event) {
	//outputPlugin.Emit(event)
	//}
	//}
	//}

	//}
	//}
	//f.metricToEmit = make(map[int64]interface{})
	return
}
