package filter

import (
	"encoding/json"
	"testing"
	"time"
)

func createEvents(now int64) []map[string]interface{} {
	var (
		event  map[string]interface{}
		events []map[string]interface{} = make([]map[string]interface{}, 0)
	)

	_15 := now - 15
	_10 := now - 10
	_5 := now - 5

	event = make(map[string]interface{})
	event["@timestamp"] = time.Unix(_15, 0)
	event["host"] = "localhost"
	event["request_statusCode"] = "200"
	event["responseTime"] = 10.1
	events = append(events, event)

	event = make(map[string]interface{})
	event["@timestamp"] = time.Unix(_10, 0)
	event["host"] = "localhost"
	event["request_statusCode"] = "301"
	event["responseTime"] = 0.1
	events = append(events, event)

	event = make(map[string]interface{})
	event["@timestamp"] = time.Unix(_10, 0)
	event["host"] = "localhost"
	event["request_statusCode"] = "200"
	event["responseTime"] = 10.1
	events = append(events, event)

	event = make(map[string]interface{})
	event["@timestamp"] = time.Unix(_10, 0)
	event["host"] = "localhost"
	event["request_statusCode"] = "200"
	event["responseTime"] = 10.3
	events = append(events, event)

	event = make(map[string]interface{})
	event["@timestamp"] = time.Unix(_10, 0)
	event["host"] = "remote"
	event["request_statusCode"] = "200"
	event["responseTime"] = 10.1
	events = append(events, event)

	event = make(map[string]interface{})
	event["@timestamp"] = time.Unix(_5, 0)
	event["host"] = "localhost"
	event["request_statusCode"] = "200"
	event["responseTime"] = 10.1
	events = append(events, event)

	event = make(map[string]interface{})
	event["@timestamp"] = time.Unix(now, 0)
	event["host"] = "localhost"
	event["request_statusCode"] = "200"
	event["responseTime"] = 10.1
	events = append(events, event)
	return events
}

func TestLinkStatsMetricFilter(t *testing.T) {
	var (
		config              map[interface{}]interface{}
		f                   *LinkStatsMetricFilter
		ok                  bool
		batchWindow         int = 5
		reserveWindow       int = 20
		windowOffset        int = 0
		ts                  int64
		drop_original_event = true
	)

	config = make(map[interface{}]interface{})
	config["fieldsLink"] = "host->request_statusCode->responseTime"
	config["reserveWindow"] = reserveWindow
	config["batchWindow"] = batchWindow
	config["windowOffset"] = windowOffset
	config["drop_original_event"] = drop_original_event

	f = NewLinkStatsMetricFilter(config)
	now := time.Now().Unix()
	for _, event := range createEvents(now) {
		f.Process(event)
	}

	t.Logf("metric: %v", f.metric)
	b, _ := json.Marshal(f.metric)
	t.Logf("metric: %s", b)

	if ok == true {
		t.Error("LinkStatsMetric filter process should return false")
	}

	if len(f.metric) != 4 {
		t.Errorf("%v", f.metricToEmit[ts])
	}

	_15 := now - 15
	_10 := now - 10

	ts = _15 - _15%(int64)(batchWindow)
	t.Logf("_15: %d", ts)
	if len(f.metric[ts].(map[string]interface{})) != 1 {
		t.Errorf("%v", f.metric[ts])
	}
	if len(f.metric[ts].(map[string]interface{})["localhost"].(map[string]interface{})) != 1 {
		t.Errorf("%v", f.metric[ts])
	}
	if len(f.metric[ts].(map[string]interface{})["localhost"].(map[string]interface{})["200"].(map[string]interface{})) != 1 {
		t.Errorf("%v", f.metric[ts])
	}
	if len(f.metric[ts].(map[string]interface{})["localhost"].(map[string]interface{})["200"].(map[string]interface{})["responseTime"].(map[string]float64)) != 2 {
		t.Errorf("%v", f.metric[ts])
	}
	if f.metric[ts].(map[string]interface{})["localhost"].(map[string]interface{})["200"].(map[string]interface{})["responseTime"].(map[string]float64)["sum"] != 10.1 {
		t.Errorf("%v", f.metric[ts])
	}

	ts = _10 - _10%(int64)(batchWindow)
	t.Logf("_10: %d", ts)
	if len(f.metric[ts].(map[string]interface{})) != 2 {
		t.Errorf("%v", f.metric[ts])
	}
	if len(f.metric[ts].(map[string]interface{})["localhost"].(map[string]interface{})) != 2 {
		t.Errorf("%v", f.metric[ts])
	}
	if len(f.metric[ts].(map[string]interface{})["localhost"].(map[string]interface{})["200"].(map[string]interface{})) != 1 {
		t.Errorf("%v", f.metric[ts])
	}
	if len(f.metric[ts].(map[string]interface{})["localhost"].(map[string]interface{})["200"].(map[string]interface{})["responseTime"].(map[string]float64)) != 2 {
		t.Errorf("%v", f.metric[ts])
	}
	if f.metric[ts].(map[string]interface{})["localhost"].(map[string]interface{})["200"].(map[string]interface{})["responseTime"].(map[string]float64)["sum"] != 20.4 {
		t.Errorf("%v", f.metric[ts])
	}

	f.swap_Metric_MetricToEmit()
	t.Logf("metricToEmit: %v", f.metricToEmit)
	if len(f.metricToEmit) != 4 {
		t.Error(f.metricToEmit)
	}

	ts = _10 - _10%(int64)(batchWindow)
	t.Logf("_10: %d", ts)
	if len(f.metricToEmit[ts].(map[string]interface{})) != 2 {
		t.Errorf("%v", f.metricToEmit[ts])
	}
	if len(f.metricToEmit[ts].(map[string]interface{})["localhost"].(map[string]interface{})) != 2 {
		t.Errorf("%v", f.metricToEmit[ts])
	}
	if len(f.metricToEmit[ts].(map[string]interface{})["localhost"].(map[string]interface{})["200"].(map[string]interface{})) != 1 {
		t.Errorf("%v", f.metricToEmit[ts])
	}
	if len(f.metricToEmit[ts].(map[string]interface{})["localhost"].(map[string]interface{})["200"].(map[string]interface{})["responseTime"].(map[string]float64)) != 2 {
		t.Errorf("%v", f.metricToEmit[ts])
	}
	if f.metricToEmit[ts].(map[string]interface{})["localhost"].(map[string]interface{})["200"].(map[string]interface{})["responseTime"].(map[string]float64)["sum"] != 20.4 {
		t.Errorf("%v", f.metricToEmit[ts])
	}
}

func TestLinkStatsMetricFilterWindowOffset(t *testing.T) {
	var (
		config              map[interface{}]interface{}
		f                   *LinkStatsMetricFilter
		ok                  bool
		batchWindow         int = 5
		reserveWindow       int = 20
		windowOffset        int = 2
		ts                  int64
		drop_original_event = true
	)

	config = make(map[interface{}]interface{})
	config["fieldsLink"] = "host->request_statusCode->responseTime"
	config["reserveWindow"] = reserveWindow
	config["batchWindow"] = batchWindow
	config["windowOffset"] = windowOffset
	config["drop_original_event"] = drop_original_event

	f = NewLinkStatsMetricFilter(config)

	now := time.Now().Unix()
	for _, event := range createEvents(now) {
		f.Process(event)
	}

	t.Logf("metric: %v", f.metric)
	b, _ := json.Marshal(f.metric)
	t.Logf("metric: %s", b)

	if ok == true {
		t.Error("LinkStatsMetric filter process should return false")
	}

	if len(f.metric) != 4 {
		t.Errorf("%v", f.metricToEmit[ts])
	}

	_15 := now - 15
	_10 := now - 10
	_5 := now - 5
	_0 := now

	ts = _15 - _15%(int64)(batchWindow)
	if len(f.metric[ts].(map[string]interface{})) != 1 {
		t.Errorf("%v", f.metric[ts])
	}
	if len(f.metric[ts].(map[string]interface{})["localhost"].(map[string]interface{})) != 1 {
		t.Errorf("%v", f.metric[ts])
	}
	if len(f.metric[ts].(map[string]interface{})["localhost"].(map[string]interface{})["200"].(map[string]interface{})) != 1 {
		t.Errorf("%v", f.metric[ts])
	}
	if len(f.metric[ts].(map[string]interface{})["localhost"].(map[string]interface{})["200"].(map[string]interface{})["responseTime"].(map[string]float64)) != 2 {
		t.Errorf("%v", f.metric[ts])
	}
	if f.metric[ts].(map[string]interface{})["localhost"].(map[string]interface{})["200"].(map[string]interface{})["responseTime"].(map[string]float64)["sum"] != 10.1 {
		t.Errorf("%v", f.metric[ts])
	}

	ts = _10 - _10%(int64)(batchWindow)
	if len(f.metric[ts].(map[string]interface{})) != 2 {
		t.Errorf("%v", f.metric[ts])
	}
	if len(f.metric[ts].(map[string]interface{})["localhost"].(map[string]interface{})) != 2 {
		t.Errorf("%v", f.metric[ts])
	}
	if len(f.metric[ts].(map[string]interface{})["localhost"].(map[string]interface{})["200"].(map[string]interface{})) != 1 {
		t.Errorf("%v", f.metric[ts])
	}
	if len(f.metric[ts].(map[string]interface{})["localhost"].(map[string]interface{})["200"].(map[string]interface{})["responseTime"].(map[string]float64)) != 2 {
		t.Errorf("%v", f.metric[ts])
	}
	if f.metric[ts].(map[string]interface{})["localhost"].(map[string]interface{})["200"].(map[string]interface{})["responseTime"].(map[string]float64)["sum"] != 20.4 {
		t.Errorf("%v", f.metric[ts])
	}

	f.swap_Metric_MetricToEmit()
	t.Logf("metricToEmit %v", f.metricToEmit)
	if len(f.metricToEmit) != 2 {
		t.Error(f.metricToEmit)
	}

	ts = _10 - _10%(int64)(batchWindow)
	if len(f.metricToEmit[ts].(map[string]interface{})) != 2 {
		t.Errorf("%v", f.metricToEmit[ts])
	}
	if len(f.metricToEmit[ts].(map[string]interface{})) != 2 {
		t.Errorf("%v", f.metricToEmit[ts])
	}
	if len(f.metricToEmit[ts].(map[string]interface{})["localhost"].(map[string]interface{})) != 2 {
		t.Errorf("%v", f.metricToEmit[ts])
	}
	if len(f.metricToEmit[ts].(map[string]interface{})["localhost"].(map[string]interface{})["200"].(map[string]interface{})) != 1 {
		t.Errorf("%v", f.metricToEmit[ts])
	}
	if len(f.metricToEmit[ts].(map[string]interface{})["localhost"].(map[string]interface{})["200"].(map[string]interface{})["responseTime"].(map[string]float64)) != 2 {
		t.Errorf("%v", f.metricToEmit[ts])
	}
	if f.metricToEmit[ts].(map[string]interface{})["localhost"].(map[string]interface{})["200"].(map[string]interface{})["responseTime"].(map[string]float64)["sum"] != 20.4 {
		t.Errorf("%v", f.metricToEmit[ts])
	}

	ts = _5 - _5%(int64)(batchWindow)
	if f.metricToEmit[ts] != nil {
		t.Errorf("_5 should be nil")
	}

	ts = _0 - _0%(int64)(batchWindow)
	if f.metricToEmit[ts] != nil {
		t.Errorf("_0 should be nil")
	}
}
