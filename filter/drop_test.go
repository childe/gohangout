package filter

import (
	"testing"
	"time"
)

func TestDropFilter(t *testing.T) {
	var (
		config map[any]any
		event  map[string]any
		ok     bool
	)

	// test DropFilter without any condition
	config = make(map[any]any)
	f := BuildFilter("Drop", config)

	event = make(map[string]any)
	event["@timestamp"] = time.Now().Unix()
	event["first"] = "dehua"
	event["last"] = "liu"

	event, ok = f.Filter(event)

	if ok == false {
		t.Error("drop filter fail")
	}

	if event != nil {
		t.Error("event should be nil after being dropped")
	}
}
