package filter

import "testing"

func TestReplaceFilter(t *testing.T) {
	config := make(map[interface{}]interface{})
	fields := make(map[interface{}]interface{})
	fields["msg"] = []interface{}{"'", `"`}
	config["fields"] = fields
	f := NewReplaceFilter(config)

	event := make(map[string]interface{})
	event["msg"] = `this is 'cat'`
	t.Log(event)

	event, ok := f.Filter(event)
	t.Log(event)
	if !ok {
		t.Error("ReplaceFilter error")
	}

	if event["msg"] != `this is "cat"` {
		t.Error(event["msg"])
	}
}
