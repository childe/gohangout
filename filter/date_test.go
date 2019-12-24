package filter

import (
	"strconv"
	"testing"
	"time"
)

func TestDateFilter(t *testing.T) {
	config := make(map[interface{}]interface{})
	config["location"] = "Asia/Shanghai"
	config["src"] = "@timestamp"
	config["formats"] = []interface{}{"RFC3339", "UNIX"}
	f := methodLibrary.NewDateFilter(config)

	event := make(map[string]interface{})
	event["@timestamp"] = time.Now().Unix()
	t.Log(event)

	event, ok := f.Filter(event)
	t.Log(event)

	if ok == false {
		t.Error("fail")
	}

	event["@timestamp"] = strconv.Itoa((int)(time.Now().Unix()))
	t.Log(event)

	event, ok = f.Filter(event)
	t.Log(event)

	if ok == false {
		t.Error("fail")
	}

	event["@timestamp"] = "2018-01-23T17:06:05+08:00"
	t.Log(event)

	event, ok = f.Filter(event)
	t.Log(event)

	if ok == false {
		t.Error("fail")
	}

	config["location"] = "Etc/UTC"
	config["formats"] = []interface{}{"2006-01-02T15:04:05"}
	f = methodLibrary.NewDateFilter(config)
	event["@timestamp"] = "2018-01-23T17:06:05"
	t.Log(event)

	event, ok = f.Filter(event)
	t.Log(event)

	if ok == false {
		t.Error("fail")
	}

}
