package filter

import "testing"

func TestGrokFilter(t *testing.T) {
	config := make(map[interface{}]interface{})
	match := make([]interface{}, 2)
	match[0] = `(?P<logtime>\S+ \S+) \[(?P<level>\w+)\] (?P<msg>.*)$`
	match[1] = `(?P<logtime>\S+ \S+)`
	config["match"] = match
	config["src"] = "message"

	f := NewGrokFilter(config)

	event := make(map[string]interface{})
	event["message"] = "2018-07-12T14:45:00 +0800 [info] message"

	event, ok := f.Filter(event)
	t.Log(event)

	if ok == false {
		t.Error("grok filter fail")
	}

	if v, ok := event["msg"]; !ok {
		t.Error("msg field should exist")
	} else {
		if v != "message" {
			t.Error("msg field do not match")
		}
	}
}
