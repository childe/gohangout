package filter

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/golang-collections/collections/stack"
	"github.com/golang/glog"
)

type LinkMetricFilter struct {
	BaseFilter

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
	baseFilter := NewBaseFilter(config)
	plugin := &LinkMetricFilter{
		BaseFilter:   baseFilter,
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

func (plugin *LinkMetricFilter) swap_Metric_MetricToEmit() {
	plugin.mutex.Lock()

	if len(plugin.metric) > 0 && len(plugin.metricToEmit) == 0 {
		timestamp := time.Now().Unix()
		timestamp -= timestamp % plugin.batchWindow

		plugin.metricToEmit = make(map[int64]interface{})
		for k, v := range plugin.metric {
			if k <= timestamp-plugin.batchWindow*plugin.windowOffset {
				plugin.metricToEmit[k] = v
			}
		}

		if plugin.accumulateMode == 1 {
			plugin.metric = make(map[int64]interface{})
		} else {
			newMetric := make(map[int64]interface{})
			for k, v := range plugin.metric {
				if k >= timestamp-plugin.reserveWindow {
					newMetric[k] = v
				}
			}
			plugin.metric = newMetric
		}
	}

	plugin.mutex.Unlock()
}
func (plugin *LinkMetricFilter) updateMetric(event map[string]interface{}) {
	lastFieldValue := event[plugin.lastField]
	if lastFieldValue == nil {
		return
	}

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
	var set map[interface{}]interface{} = nil
	if v, ok := plugin.metric[timestamp]; ok {
		set = v.(map[interface{}]interface{})
	} else {
		set = make(map[interface{}]interface{})
		plugin.metric[timestamp] = set
	}

	for _, field := range plugin.fieldsWithoutLast {
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
}

func (plugin *LinkMetricFilter) Process(event map[string]interface{}) (map[string]interface{}, bool) {
	plugin.updateMetric(event)
	if plugin.dropOriginalEvent {
		return nil, false
	}
	return event, false
}

func (plugin *LinkMetricFilter) metricToEvents(metrics map[interface{}]interface{}, level int) []map[string]interface{} {
	var (
		fieldName string                   = plugin.fields[level]
		events    []map[string]interface{} = make([]map[string]interface{}, 0)
	)

	if level == plugin.fieldsLength-1 {
		for fieldValue, count := range metrics {
			event := make(map[string]interface{})
			event[fmt.Sprintf("%s", fieldName)] = fieldValue
			event["count"] = count
			events = append(events, event)
		}
		return events
	}

	for fieldValue, nextLevelMetrics := range metrics {
		for _, e := range plugin.metricToEvents(nextLevelMetrics.(map[interface{}]interface{}), level+1) {
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

func (plugin *LinkMetricFilter) EmitExtraEvents(sTo *stack.Stack) {
	if len(plugin.metricToEmit) == 0 {
		return
	}
	var event map[string]interface{}
	for timestamp, metrics := range plugin.metricToEmit {
		for _, event = range plugin.metricToEvents(metrics.(map[interface{}]interface{}), 0) {
			event[plugin.timestamp] = time.Unix(timestamp, 0)
			event = plugin.PostProcess(event, true)
			sTo.Push(event)
		}
	}
	plugin.metricToEmit = make(map[int64]interface{})
	return
}
