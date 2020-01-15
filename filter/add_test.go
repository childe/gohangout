package filter

import (
	"testing"
	"time"
)

func TestAddFilter(t *testing.T) {
	config := make(map[interface{}]interface{})
	fields := make(map[interface{}]interface{})
	fields["name"] = `{{.first}} {{.last}}`
	fields["name2"] = `$.name`
	config["fields"] = fields
	f := methodLibrary.NewAddFilter(config)

	event := make(map[string]interface{})
	event["@timestamp"] = time.Now().Unix()
	event["first"] = "dehua"
	event["last"] = "liu"
	t.Log(event)

	event, ok := f.Filter(event)
	t.Log(event)

	if ok == false {
		t.Error("add filter fail")
	}

	name, ok := event["name"]
	if ok == false {
		t.Error("add filter should add `name` field")
	}
	if name != "dehua liu" {
		t.Error("name field should be `dehua liu`")
	}

	name2, ok := event["name2"]
	if ok == false {
		t.Error("add filter should add `name2` field")
	}
	if name2 != "dehua liu" {
		t.Error("name2 field should be `dehua liu`")
	}
}
