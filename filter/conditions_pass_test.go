package filter

import (
	"testing"
	"time"
)

func TestConditions(t *testing.T) {
	var (
		config map[interface{}]interface{}
		f      *DropFilter
		event  map[string]interface{}
		pass   bool
	)
	// test DropFilter with condition
	config = make(map[interface{}]interface{})
	conditions := make([]interface{}, 3)
	conditions[0] = "{{if .name}}y{{end}}"
	conditions[1] = "{{if .name.first}}y{{end}}"
	conditions[2] = `{{if eq .name.first "dehua"}}y{{end}}`
	config["if"] = conditions
	f = NewDropFilter(config)

	// should drop
	event = make(map[string]interface{})
	event["@timestamp"] = time.Now().Unix()
	event["name"] = map[string]interface{}{"first": "dehua"}

	pass = f.Pass(event)

	if pass == false {
		t.Error("should pass the conditions")
	}

	// should not drop
	event = make(map[string]interface{})
	event["@timestamp"] = time.Now().Unix()
	event["name"] = map[string]interface{}{"last": "liu"}

	pass = f.Pass(event)

	if pass == true {
		t.Error("should not pass the conditions")
	}
}
