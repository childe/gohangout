package filter

import "testing"

func TestConvertFilter(t *testing.T) {
	config := make(map[interface{}]interface{})
	fields := make(map[interface{}]interface{})
	fields["responseSize"] = map[interface{}]interface{}{
		"to":            "int",
		"setto_if_fail": 0,
	}
	fields["timeTaken"] = map[interface{}]interface{}{
		"to":             "float",
		"remove_if_fail": true,
	}

	config["fields"] = fields
	f := methodLibrary.NewConvertFilter(config)

	event := map[string]interface{}{
		"responseSize": "10",
		"timeTaken":    "0.010",
	}
	t.Log(event)

	event, ok := f.Filter(event)
	t.Log(event)

	if ok == false {
		t.Error("ConvertFilter fail")
	}

	if event["responseSize"].(int) != 10 {
		t.Error("responseSize should be 10")
	}
	if event["timeTaken"].(float64) != 0.01 {
		t.Error("timeTaken should be 0.01")
	}

	event = map[string]interface{}{
		"responseSize": "10.1",
		"timeTaken":    "abcd",
	}
	t.Log(event)

	event, ok = f.Filter(event)
	t.Log(event)

	if ok == false {
		t.Error("ConvertFilter fail")
	}

	if event["responseSize"].(int) != 0 {
		t.Error("responseSize should be 0")
	}
	if event["timeTaken"] != nil {
		t.Error("timeTaken should be nil")
	}
}
