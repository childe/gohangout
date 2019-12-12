package filter

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/childe/gohangout/topology"
	"github.com/golang/glog"
)

type stats struct {
	count int
	min   float64
	max   float64
	sum   float64
}

type LinkStatsMetricFilter struct {
	next              topology.Processor
	config            map[interface{}]interface{}
	timestamp         string
	batchWindow       int64
	reserveWindow     int64
	dropOriginalEvent bool
	windowOffset      int64
	accumulateMode    int
	reduce            bool

	fields            []string
	fieldsWithoutLast []string
	lastField         string
	fieldsLength      int

	metric       map[int64]interface{}
	metricToEmit map[int64]interface{}

	mutex sync.Locker
}

func (f *LinkStatsMetricFilter) SetBelongTo(next topology.Processor) {
	f.next = next
}

func (l *MethodLibrary) NewLinkStatsMetricFilter(config map[interface{}]interface{}) *LinkStatsMetricFilter {
	p := &LinkStatsMetricFilter{
		config:       config,
		metric:       make(map[int64]interface{}),
		metricToEmit: make(map[int64]interface{}),

		mutex: &sync.Mutex{},
	}

	if fieldsLink, ok := config["fieldsLink"]; ok {
		p.fields = strings.Split(fieldsLink.(string), "->")
		p.fieldsLength = len(p.fields)
		p.fieldsWithoutLast = p.fields[:p.fieldsLength-1]
		p.lastField = p.fields[p.fieldsLength-1]
	} else {
		glog.Fatal("fieldsLink must be set in linkstatmetric filter plugin")
	}

	if timestamp, ok := config["timestamp"]; ok {
		p.timestamp = timestamp.(string)
	} else {
		p.timestamp = "@timestamp"
	}

	if dropOriginalEvent, ok := config["drop_original_event"]; ok {
		p.dropOriginalEvent = dropOriginalEvent.(bool)
	} else {
		p.dropOriginalEvent = false
	}

	if batchWindow, ok := config["batchWindow"]; ok {
		p.batchWindow = int64(batchWindow.(int))
	} else {
		glog.Fatal("batchWindow must be set in linkstatmetric filter plugin")
	}

	if reserveWindow, ok := config["reserveWindow"]; ok {
		p.reserveWindow = int64(reserveWindow.(int))
	} else {
		glog.Fatal("reserveWindow must be set in linkstatmetric filter plugin")
	}

	if reduce, ok := config["reduce"]; ok {
		p.reduce = reduce.(bool)
	}

	if accumulateModeI, ok := config["accumulateMode"]; ok {
		accumulateMode := accumulateModeI.(string)
		switch accumulateMode {
		case "cumulative":
			p.accumulateMode = 0
		case "separate":
			p.accumulateMode = 1
		default:
			glog.Errorf("invalid accumulateMode: %s. set to cumulative", accumulateMode)
			p.accumulateMode = 0
		}
	} else {
		p.accumulateMode = 0
	}

	if windowOffset, ok := config["windowOffset"]; ok {
		p.windowOffset = (int64)(windowOffset.(int))
	} else {
		p.windowOffset = 0
	}

	ticker := time.NewTicker(time.Second * time.Duration(p.batchWindow))
	go func() {
		for range ticker.C {
			p.swap_Metric_MetricToEmit()
			p.emitMetrics()
		}
	}()
	return p
}

func (f *LinkStatsMetricFilter) metricToEvents(metrics map[interface{}]interface{}, level int) []map[string]interface{} {
	var (
		fieldName string                   = f.fields[level]
		events    []map[string]interface{} = make([]map[string]interface{}, 0)
	)

	if level == f.fieldsLength-1 {
		for _, statsI := range metrics {
			s := statsI.(*stats)
			event := make(map[string]interface{})
			event["count"] = s.count
			event["sum"] = s.sum
			event["min"] = s.min
			event["max"] = s.max
			event["mean"] = s.sum / float64(s.count)
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

func (f *LinkStatsMetricFilter) swap_Metric_MetricToEmit() {
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
func (f *LinkStatsMetricFilter) updateMetric(event map[string]interface{}) {
	var value float64
	var (
		count int
		min   float64
		max   float64
		sum   float64
	)

	if f.reduce {
		if c, ok := event["count"]; !ok {
			return
		} else {
			count = c.(int)
		}
		if s, ok := event["sum"]; !ok {
			return
		} else {
			sum = s.(float64)
		}
		if s, ok := event["min"]; !ok {
			return
		} else {
			min = s.(float64)
		}
		if s, ok := event["max"]; !ok {
			return
		} else {
			max = s.(float64)
		}
	} else {
		fieldValueI := event[f.lastField]
		if fieldValueI == nil {
			return
		}
		value = fieldValueI.(float64)

		count, min, max, sum = 1, value, value, value
	}

	var timestamp int64
	if v, ok := event[f.timestamp]; ok {
		if t, ok := v.(time.Time); !ok {
			glog.V(20).Infof("timestamp is not time.Time type")
			return
		} else {
			timestamp = t.Unix()
		}
	} else {
		glog.V(20).Infof("no timestamp in event. %s", event)
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

	if statsI, ok := set[f.lastField]; ok {
		s := statsI.(*stats)
		s.count = count + s.count
		s.sum = sum + s.sum
		s.min = math.Min(s.min, min)
		s.max = math.Max(s.max, max)
		set[f.lastField] = s
	} else {
		set[f.lastField] = &stats{count, min, max, sum}
	}
}

func (f *LinkStatsMetricFilter) emitMetrics() {
	if len(f.metricToEmit) == 0 {
		return
	}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	var event map[string]interface{}
	for timestamp, metrics := range f.metricToEmit {
		for _, event = range f.metricToEvents(metrics.(map[interface{}]interface{}), 0) {
			event[f.timestamp] = time.Unix(timestamp, 0)
			f.next.Process(event)
		}
	}
	f.metricToEmit = make(map[int64]interface{})
}

func (f *LinkStatsMetricFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	f.updateMetric(event)
	if f.dropOriginalEvent {
		return nil, false
	}
	return event, false
}
