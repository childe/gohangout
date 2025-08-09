package filter

import (
	"math"
	"strings"
	"sync"
	"time"

	"github.com/childe/gohangout/topology"
	"k8s.io/klog/v2"
)

type stats struct {
	count int
	min   float64
	max   float64
	sum   float64
}

type LinkStatsMetricFilter struct {
	next              topology.Processor
	config            map[any]any
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

	metric       map[int64]any
	metricToEmit map[int64]any

	mutex sync.Locker
}

func (f *LinkStatsMetricFilter) SetBelongTo(next topology.Processor) {
	f.next = next
}

func init() {
	Register("LinkStatsMetric", newLinkStatsMetricFilter)
}

func newLinkStatsMetricFilter(config map[any]any) topology.Filter {
	p := &LinkStatsMetricFilter{
		config:       config,
		metric:       make(map[int64]any),
		metricToEmit: make(map[int64]any),

		mutex: &sync.Mutex{},
	}

	if fieldsLink, ok := config["fieldsLink"]; ok {
		p.fields = strings.Split(fieldsLink.(string), "->")
		p.fieldsLength = len(p.fields)
		p.fieldsWithoutLast = p.fields[:p.fieldsLength-1]
		p.lastField = p.fields[p.fieldsLength-1]
	} else {
		klog.Fatal("fieldsLink must be set in linkstatmetric filter plugin")
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
		klog.Fatal("batchWindow must be set in linkstatmetric filter plugin")
	}

	if reserveWindow, ok := config["reserveWindow"]; ok {
		p.reserveWindow = int64(reserveWindow.(int))
	} else {
		klog.Fatal("reserveWindow must be set in linkstatmetric filter plugin")
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
			klog.Errorf("invalid accumulateMode: %s. set to cumulative", accumulateMode)
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

func (f *LinkStatsMetricFilter) metricToEvents(metrics map[any]any, level int) []map[string]any {
	var (
		fieldName string                   = f.fields[level]
		events    []map[string]any = make([]map[string]any, 0)
	)

	if level == f.fieldsLength-1 {
		for _, statsI := range metrics {
			s := statsI.(*stats)
			event := make(map[string]any)
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
		for _, e := range f.metricToEvents(nextLevelMetrics.(map[any]any), level+1) {
			event := make(map[string]any)
			event[fieldName] = fieldValue
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

		f.metricToEmit = make(map[int64]any)
		for k, v := range f.metric {
			if k <= timestamp-f.batchWindow*f.windowOffset {
				f.metricToEmit[k] = v
			}
		}

		if f.accumulateMode == 1 {
			f.metric = make(map[int64]any)
		} else {
			newMetric := make(map[int64]any)
			for k, v := range f.metric {
				if k >= timestamp-f.reserveWindow {
					newMetric[k] = v
				}
			}
			f.metric = newMetric
		}
	}
}
func (f *LinkStatsMetricFilter) updateMetric(event map[string]any) {
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
			klog.V(20).Infof("timestamp is not time.Time type")
			return
		} else {
			timestamp = t.Unix()
		}
	} else {
		klog.V(20).Infof("no timestamp in event. %s", event)
		return
	}

	diff := time.Now().Unix() - timestamp
	if diff > f.reserveWindow || diff < 0 {
		return
	}

	timestamp -= timestamp % f.batchWindow
	var set map[any]any = nil
	if v, ok := f.metric[timestamp]; ok {
		set = v.(map[any]any)
	} else {
		set = make(map[any]any)
		f.metric[timestamp] = set
	}

	for _, field := range f.fieldsWithoutLast {
		fieldValue := event[field]
		if fieldValue == nil {
			return
		}
		if v, ok := set[fieldValue]; ok {
			set = v.(map[any]any)
		} else {
			set[fieldValue] = make(map[any]any)
			set = set[fieldValue].(map[any]any)
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

	var event map[string]any
	for timestamp, metrics := range f.metricToEmit {
		for _, event = range f.metricToEvents(metrics.(map[any]any), 0) {
			event[f.timestamp] = time.Unix(timestamp, 0)
			f.next.Process(event)
		}
	}
	f.metricToEmit = make(map[int64]any)
}

func (f *LinkStatsMetricFilter) Filter(event map[string]any) (map[string]any, bool) {
	f.updateMetric(event)
	if f.dropOriginalEvent {
		return nil, false
	}
	return event, false
}
