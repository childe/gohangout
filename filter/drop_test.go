package filter

import (
	"testing"
	"time"
)

func TestDropFilter(t *testing.T) {
	var (
		config map[interface{}]interface{}
		event  map[string]interface{}
		ok     bool
	)

	// test DropFilter without any condition
	config = make(map[interface{}]interface{})

	config["if"] = []interface{}{
		`EQ($.level,"error")`,
	}

	f := BuildFilter("Drop", config)

	event = make(map[string]interface{})
	event["@timestamp"] = time.Now().Unix()
	event["first"] = "dehua"
	event["last"] = "liu"

	// test level = 1
	event["level"] = "1"

	event, ok = f.Filter(event)

	if event == nil || ok {
		t.Error("event should not be nil after being dropped and ok should be false")
	}

	// test level = error
	event["level"] = "error"
	event, ok = f.Filter(event)
	if event != nil || !ok {
		t.Error("event should be nil after being dropped and ok should be true")
	}
}
