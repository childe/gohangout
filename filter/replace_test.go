package filter

import "testing"

func TestReplaceFilter(t *testing.T) {
	config := make(map[any]any)
	fields := make(map[any]any)
	fields["msg"] = []any{"'", `"`}
	config["fields"] = fields
	f := BuildFilter("Replace", config)

	event := make(map[string]any)
	event["msg"] = `this is 'cat'`

	event, ok := f.Filter(event)
	t.Log(event)
	if !ok {
		t.Error("ReplaceFilter error")
	}

	if event["msg"] != `this is "cat"` {
		t.Error(event["msg"])
	}

	config = make(map[any]any)
	fields = make(map[any]any)
	fields["name1"] = []any{"wang", "Wang", 1}
	fields["name2"] = []any{"en", "eng"}
	config["fields"] = fields
	f = BuildFilter("Replace", config)

	event = make(map[string]any)
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
