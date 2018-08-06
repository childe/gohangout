package filter

import "testing"

func TestConvertFilter(t *testing.T) {
	config := make(map[interface{}]interface{})
	fields := make(map[interface{}]interface{})
	fields["responseSize"] = map[string]interface{}{
		"to":            "int",
		"setto_if_fail": 0,
	}
	fields["timeTaken"] = map[string]interface{}{
		"to":             "float",
		"remove_if_fail": true,
	}

	config["fields"] = fields
	f := NewConvertFilter(config)

	event := map[string]interface{}{
		"responseSize": "10",
		"timeTaken":    "0.010",
	}
	t.Log(event)

	event, ok := f.Process(event)
	t.Log(event)

	if ok == false {
		t.Error("ConvertFilter fail")
	}

	t.Log(event["responseSize"].(int64) == 10)
	if event["responseSize"].(int64) != 10 {
		t.Error("responseSize should be 10")
	}
	if event["timeTaken"].(float64) != 0.01 {
		t.Error("timeTaken should be 0.01")
	}
}
