package filter

import "testing"

func TestReplaceFilter(t *testing.T) {
	config := make(map[interface{}]interface{})
	fields := make(map[interface{}]interface{})
	fields["msg"] = []interface{}{"'", `"`}
	config["fields"] = fields
	f := methodLibrary.NewReplaceFilter(config)

	event := make(map[string]interface{})
	event["msg"] = `this is 'cat'`

	event, ok := f.Filter(event)
	t.Log(event)
	if !ok {
		t.Error("ReplaceFilter error")
	}

	if event["msg"] != `this is "cat"` {
		t.Error(event["msg"])
	}

	config = make(map[interface{}]interface{})
	fields = make(map[interface{}]interface{})
	fields["name1"] = []interface{}{"wang", "Wang", 1}
	fields["name2"] = []interface{}{"en", "eng"}
	config["fields"] = fields
	f = methodLibrary.NewReplaceFilter(config)

	event = make(map[string]interface{})
	event["name1"] = "wang wangwang"
	event["name2"] = "wang henhen"

	event, ok = f.Filter(event)
	t.Log(event)
	if !ok {
		t.Error("ReplaceFilter error")
	}

	if event["name1"] != "Wang wangwang" {
		t.Error(event["name1"])
	}

	if event["name2"] != "wang hengheng" {
		t.Error(event["name2"])
	}
}
