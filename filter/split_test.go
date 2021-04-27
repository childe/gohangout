package filter

import "testing"

func TestSplitFilter1(t *testing.T) {
	config := make(map[interface{}]interface{})
	fields := []interface{}{"loglevel", "date", "time", "message"}
	config["src"] = "message"
	config["fields"] = fields
	config["sep"] = " "
	config["maxSplit"] = 4
	config["trim"] = "[]"
	f := BuildFilter("Split", config)

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
}

func TestSplitFilter2(t *testing.T) {
	config := make(map[interface{}]interface{})
	fields := []interface{}{"loglevel", "logtime", "message"}
	config["src"] = "message"
	config["fields"] = fields
	config["sep"] = "] "
	config["maxSplit"] = 3
	config["trim"] = "[]"
	f := BuildFilter("Split", config)

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

	if event["logtime"] != "2019-03-21 23:59:59,998" {
		t.Errorf("logtime error: %#v", event)
	}

	if event["message"] != "messages ..." {
		t.Errorf("message error: %#v", event)
	}
}

// dynamic sep
func TestSplitFilter3(t *testing.T) {
	config := make(map[interface{}]interface{})
	fields := []interface{}{"loglevel", "date", "time", "message"}
	config["src"] = "message"
	config["fields"] = fields
	config["sep"] = "[sep]"
	config["dynamicSep"] = true
	config["maxSplit"] = 4
	config["trim"] = "[]"
	f := BuildFilter("Split", config)

	event := make(map[string]interface{})
	event["message"] = `[INFO] [2019-03-21 23:59:59,998] messages ...`
	event["sep"] = " "

	event, ok := f.Filter(event)
	t.Log(event)
	if !ok {
		t.Error("SplitFilter error")
	}

	if event["loglevel"] != "INFO" {
		t.Errorf("loglevel error: %#v", event)
	}

	if event["date"] != "2019-03-21" {
		t.Errorf("logtime error: %#v", event)
	}

	if event["time"] != "23:59:59,998" {
		t.Errorf("logtime error: %#v", event)
	}

	if event["message"] != "messages ..." {
		t.Errorf("message error: %#v", event)
	}
}

// length of fileds do not match length of splited
func TestSplitFilter4(t *testing.T) {
	config := make(map[interface{}]interface{})
	fields := []interface{}{"loglevel", "date", "time"}
	config["src"] = "message"
	config["fields"] = fields
	config["sep"] = "[sep]"
	config["dynamicSep"] = true
	config["maxSplit"] = 4
	config["trim"] = "[]"
	f := BuildFilter("Split", config)

	event := make(map[string]interface{})
	event["message"] = `[INFO] [2019-03-21 23:59:59,998] messages ...`
	event["sep"] = " "

	event, ok := f.Filter(event)
	t.Log(event)
	if !ok {
		t.Error("SplitFilter error")
	}

	if event["loglevel"] != "INFO" {
		t.Errorf("loglevel error: %#v", event)
	}

	if event["date"] != "2019-03-21" {
		t.Errorf("logtime error: %#v", event)
	}

	if event["time"] != "23:59:59,998" {
		t.Errorf("logtime error: %#v", event)
	}

	if event["message"] != `[INFO] [2019-03-21 23:59:59,998] messages ...` {
		t.Errorf("message error: %#v", event)
	}
}
