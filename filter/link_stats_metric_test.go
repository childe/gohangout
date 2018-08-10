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

	f.swap_Metric_MetricToEmit()
	t.Logf("metricToEmit %v", f.metricToEmit)
	if len(f.metricToEmit) != 4 {
		t.Error(f.metricToEmit)
	}

	_15 := now - 15
	ts = _15 - _15%(int64)(batchWindow)
	if f.metricToEmit[ts] == nil {
		t.Errorf("_15 should be in metricToEmit")
	}
	_5 := now - 5
	ts = _5 - _5%(int64)(batchWindow)
	if f.metricToEmit[ts] == nil {
		t.Errorf("_5 should be in metricToEmit")
	}
	_0 := now
	ts = _0 - _0%(int64)(batchWindow)
	if f.metricToEmit[ts] == nil {
		t.Errorf("_0 should be in metricToEmit")
	}

	_10 := now - 10
	ts = _10 - _10%(int64)(batchWindow)
	_10_metric := f.metricToEmit[ts].(map[string]interface{})
	if len(_10_metric) != 2 {
		t.Errorf("_10 metric length should be 2")
	}
	if _10_metric["localhost"] == nil {
		t.Errorf("localhost should be in _10 metric")
	}
	if _10_metric["remote"] == nil {
		t.Errorf("remote should be in _10 metric")
	}

	localhost_metric := f.metricToEmit[ts].(map[string]interface{})["localhost"].(map[string]interface{})
	if len(localhost_metric) != 2 {
		t.Errorf("localhost metric length should be 2")
	}
	if localhost_metric["200"] == nil {
		t.Errorf("200 should be in localhost_metric")
	}
	if localhost_metric["301"] == nil {
		t.Errorf("301 should be in localhost_metric")
	}

	if len(localhost_metric["200"].(map[string]interface{})) != 1 {
		t.Errorf("%v", f.metricToEmit[ts])
	}
	if len(localhost_metric["200"].(map[string]interface{})["responseTime"].(map[string]float64)) != 2 {
		t.Errorf("%v", f.metricToEmit[ts])
	}
	if localhost_metric["200"].(map[string]interface{})["responseTime"].(map[string]float64)["count"] != 2 {
		t.Errorf("_10->localhost->200->responseTime-count should be 2")
	}
	if localhost_metric["200"].(map[string]interface{})["responseTime"].(map[string]float64)["sum"] != 20.4 {
		t.Errorf("_10->localhost->200->responseTime-sum should be 20.4")
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

	f.swap_Metric_MetricToEmit()
	t.Logf("metricToEmit %v", f.metricToEmit)
	if len(f.metricToEmit) != 2 {
		t.Error(f.metricToEmit)
	}

	_15 := now - 15
	ts = _15 - _15%(int64)(batchWindow)
	if f.metricToEmit[ts] == nil {
		t.Errorf("_15 should be in metricToEmit")
	}
	_5 := now - 5
	ts = _5 - _5%(int64)(batchWindow)
	if f.metricToEmit[ts] != nil {
		t.Errorf("_5 should not be in metricToEmit")
	}
	_0 := now
	ts = _0 - _0%(int64)(batchWindow)
	if f.metricToEmit[ts] != nil {
		t.Errorf("_0 should not be in metricToEmit")
	}

	_10 := now - 10
	ts = _10 - _10%(int64)(batchWindow)
	_10_metric := f.metricToEmit[ts].(map[string]interface{})
	if len(_10_metric) != 2 {
		t.Errorf("_10 metric length should be 2")
	}
	if _10_metric["localhost"] == nil {
		t.Errorf("localhost should be in _10 metric")
	}
	if _10_metric["remote"] == nil {
		t.Errorf("remote should be in _10 metric")
	}

	localhost_metric := f.metricToEmit[ts].(map[string]interface{})["localhost"].(map[string]interface{})
	if len(localhost_metric) != 2 {
		t.Errorf("localhost metric length should be 2")
	}
	if localhost_metric["200"] == nil {
		t.Errorf("200 should be in localhost_metric")
	}
	if localhost_metric["301"] == nil {
		t.Errorf("301 should be in localhost_metric")
	}

	if len(localhost_metric["200"].(map[string]interface{})) != 1 {
		t.Errorf("%v", f.metricToEmit[ts])
	}
	if len(localhost_metric["200"].(map[string]interface{})["responseTime"].(map[string]float64)) != 2 {
		t.Errorf("%v", f.metricToEmit[ts])
	}
	if localhost_metric["200"].(map[string]interface{})["responseTime"].(map[string]float64)["count"] != 2 {
		t.Errorf("_10->localhost->200->responseTime-count should be 2")
	}
	if localhost_metric["200"].(map[string]interface{})["responseTime"].(map[string]float64)["sum"] != 20.4 {
		t.Errorf("_10->localhost->200->responseTime-sum should be 20.4")
	}
}

func _testLinkStatsMetricFilterAccumulateMode(t *testing.T, f *LinkStatsMetricFilter, now int64, count, sum float64) {
	if len(f.metricToEmit) != 4 {
		t.Error(f.metricToEmit)
	}

	var ts int64
	_15 := now - 15
	ts = _15 - _15%f.batchWindow
	if f.metricToEmit[ts] == nil {
		t.Errorf("_15 should be in metricToEmit")
	}
	_5 := now - 5
	ts = _5 - _5%f.batchWindow
	if f.metricToEmit[ts] == nil {
		t.Errorf("_5 should be in metricToEmit")
	}
	_0 := now
	ts = _0 - _0%f.batchWindow
	if f.metricToEmit[ts] == nil {
		t.Errorf("_0 should be in metricToEmit")
	}

	_10 := now - 10
	ts = _10 - _10%f.batchWindow
	_10_metric := f.metricToEmit[ts].(map[string]interface{})
	if len(_10_metric) != 2 {
		t.Errorf("_10 metric length should be 2")
	}
	if _10_metric["localhost"] == nil {
		t.Errorf("localhost should be in _10 metric")
	}
	if _10_metric["remote"] == nil {
		t.Errorf("remote should be in _10 metric")
	}

	localhost_metric := f.metricToEmit[ts].(map[string]interface{})["localhost"].(map[string]interface{})
	if len(localhost_metric) != 2 {
		t.Errorf("localhost metric length should be 2")
	}
	if localhost_metric["200"] == nil {
		t.Errorf("200 should be in localhost_metric")
	}
	if localhost_metric["301"] == nil {
		t.Errorf("301 should be in localhost_metric")
	}

	if len(localhost_metric["200"].(map[string]interface{})) != 1 {
		t.Errorf("%v", f.metricToEmit[ts])
	}
	if len(localhost_metric["200"].(map[string]interface{})["responseTime"].(map[string]float64)) != 2 {
		t.Errorf("%v", f.metricToEmit[ts])
	}
	if localhost_metric["200"].(map[string]interface{})["responseTime"].(map[string]float64)["count"] != count {
		t.Errorf("_10->localhost->200->responseTime-count should be %f", count)
	}
	if localhost_metric["200"].(map[string]interface{})["responseTime"].(map[string]float64)["sum"] != sum {
		t.Errorf("_10->localhost->200->responseTime-sum should be %f", sum)
	}
}

func TestLinkStatsMetricFilterCumulativeMode(t *testing.T) {
	var (
		config              map[interface{}]interface{}
		f                   *LinkStatsMetricFilter
		batchWindow         int    = 5
		reserveWindow       int    = 20
		windowOffset        int    = 0
		accumulateMode      string = "cumulative"
		drop_original_event        = true
	)

	config = make(map[interface{}]interface{})
	config["fieldsLink"] = "host->request_statusCode->responseTime"
	config["reserveWindow"] = reserveWindow
	config["batchWindow"] = batchWindow
	config["windowOffset"] = windowOffset
	config["accumulateMode"] = accumulateMode
	config["drop_original_event"] = drop_original_event

	f = NewLinkStatsMetricFilter(config)

	now := time.Now().Unix()
	for _, event := range createEvents(now) {
		f.Process(event)
	}
	f.swap_Metric_MetricToEmit()
	for _, event := range createEvents(now) {
		f.Process(event)
	}

	f.swap_Metric_MetricToEmit()
	t.Logf("metricToEmit: %v", f.metricToEmit)

	_testLinkStatsMetricFilterAccumulateMode(t, f, now, 4, 40.8)
}

func TestLinkStatsMetricFilterSeparateMode(t *testing.T) {
	var (
		config              map[interface{}]interface{}
		f                   *LinkStatsMetricFilter
		batchWindow         int    = 5
		reserveWindow       int    = 20
		windowOffset        int    = 0
		accumulateMode      string = "separate"
		drop_original_event        = true
	)

	config = make(map[interface{}]interface{})
	config["fieldsLink"] = "host->request_statusCode->responseTime"
	config["reserveWindow"] = reserveWindow
	config["batchWindow"] = batchWindow
	config["windowOffset"] = windowOffset
	config["accumulateMode"] = accumulateMode
	config["drop_original_event"] = drop_original_event

	f = NewLinkStatsMetricFilter(config)

	now := time.Now().Unix()
	for _, event := range createEvents(now) {
		f.Process(event)
	}
	f.swap_Metric_MetricToEmit()
	for _, event := range createEvents(now) {
		f.Process(event)
	}

	f.swap_Metric_MetricToEmit()
	t.Logf("metricToEmit: %v", f.metricToEmit)
	if len(f.metricToEmit) != 4 {
		t.Error(f.metricToEmit)
	}

	_testLinkStatsMetricFilterAccumulateMode(t, f, now, 2, 20.4)
}
