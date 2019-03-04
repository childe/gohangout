package value_render

import (
	"testing"
	"time"
)

func TestIndexRender(t *testing.T) {
	var event map[string]interface{}
	var template string
	var vr ValueRender
	var indexname string

	// only timestamp
	event = make(map[string]interface{})
	event["@timestamp"], _ = time.Parse("2006-01-02T15:04:05", "2019-03-04T14:21:00")

	template = "nginx-%{+2006.01.02}"

	vr = NewIndexRender(template)
	indexname = vr.Render(event).(string)
	t.Log(indexname)

	if indexname != "nginx-2019.03.04" {
		t.Errorf("%s != nginx-2019.03.04\n", indexname)
	}

	// only timestamp, NOT endswith timestamp
	event = make(map[string]interface{})
	event["@timestamp"], _ = time.Parse("2006-01-02T15:04:05", "2019-03-04T14:21:00")

	template = "nginx-%{+2006.01.02}-log"

	vr = NewIndexRender(template)
	indexname = vr.Render(event).(string)
	t.Log(indexname)

	if indexname != "nginx-2019.03.04-log" {
		t.Errorf("%s != nginx-2019.03.04-log\n", indexname)
	}

	// timestamp, appid
	event = make(map[string]interface{})
	event["@timestamp"], _ = time.Parse("2006-01-02T15:04:05", "2019-03-04T14:21:00")
	event["appid"] = "100234"

	template = "nginx-%{appid}-%{+2006.01.02}"

	vr = NewIndexRender(template)
	indexname = vr.Render(event).(string)
	t.Log(indexname)

	if indexname != "nginx-100234-2019.03.04" {
		t.Errorf("%s != nginx-100234-2019.03.04\n", indexname)
	}

	// timestamp exists, appid missing
	event = make(map[string]interface{})
	event["@timestamp"], _ = time.Parse("2006-01-02T15:04:05", "2019-03-04T14:21:00")

	template = "nginx-%{appid}-%{+2006.01.02}"

	vr = NewIndexRender(template)
	indexname = vr.Render(event).(string)
	t.Log(indexname)

	if indexname != "nginx-null-2019.03.04" {
		t.Errorf("%s != nginx-null-2019.03.04\n", indexname)
	}

}
