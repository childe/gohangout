package filter

import (
	"testing"
	"time"
)

func TestDropFilter(t *testing.T) {
	var (
		config map[interface{}]interface{}
		f      *DropFilter
		event  map[string]interface{}
		ok     bool
	)

	// test DropFilter without any condition
	config = make(map[interface{}]interface{})
	f = methodLibrary.NewDropFilter(config)

	event = make(map[string]interface{})
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
