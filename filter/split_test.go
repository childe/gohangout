package filter

import "testing"

func TestSplitFilter(t *testing.T) {
	config := make(map[interface{}]interface{})
	fields := []interface{}{"loglevel", "date", "time", "message"}
	config["src"] = "message"
	config["fields"] = fields
	config["sep"] = " "
	config["maxSplit"] = 4
	config["trim"] = "[]"
	f := methodLibrary.NewSplitFilter(config)

	event := make(map[string]interface{})
	event["message"] = `[INFO] [2019-03-21 23:59:59,998] messages ...`

	event, ok := f.Filter(event)
	t.Log(event)
	if !ok {
		t.Error("SplitFilter error")
	}

	if event["loglevel"] != "INFO" {
		t.Errorf("loglevel error: %#v", event)
	}

	if event["date"] != "2019-03-21" {
		t.Errorf("date error: %#v", event)
	}

	if event["time"] != "23:59:59,998" {
		t.Errorf("time error: %#v", event)
	}

	if event["message"] != "messages ..." {
		t.Errorf("message error: %#v", event)
	}

	config = make(map[interface{}]interface{})
	fields = []interface{}{"loglevel", "logtime", "message"}
	config["src"] = "message"
	config["fields"] = fields
	config["sep"] = "] "
	config["maxSplit"] = 3
	config["trim"] = "[]"
	f = methodLibrary.NewSplitFilter(config)

	event = make(map[string]interface{})
	event["message"] = `[INFO] [2019-03-21 23:59:59,998] messages ...`

	event, ok = f.Filter(event)
	t.Log(event)
	if !ok {
		t.Error("SplitFilter error")
	}

	if event["loglevel"] != "INFO" {
		t.Errorf("loglevel error: %#v", event)
	}

	if event["logtime"] != "2019-03-21 23:59:59,998" {
		t.Errorf("logtime error: %#v", event)
	}

	if event["message"] != "messages ..." {
		t.Errorf("message error: %#v", event)
	}

	// dynamic sep
	config = make(map[interface{}]interface{})
	fields = []interface{}{"loglevel", "logtime", "message"}
	config["src"] = "message"
	config["fields"] = fields
	config["sep"] = "] "
	config["dynamicSep"] = true
	config["maxSplit"] = 3
	config["trim"] = "[]"
	f = methodLibrary.NewSplitFilter(config)

	event = make(map[string]interface{})
	event["message"] = `[INFO] [2019-03-21 23:59:59,998] messages ...`
	event["sep"] = " "

	event, ok = f.Filter(event)
	t.Log(event)
	if !ok {
		t.Error("SplitFilter error")
	}

	if event["loglevel"] != "INFO" {
		t.Errorf("loglevel error: %#v", event)
	}

	if event["logtime"] != "2019-03-21 23:59:59,998" {
		t.Errorf("logtime error: %#v", event)
	}

	if event["message"] != "messages ..." {
		t.Errorf("message error: %#v", event)
	}
}
