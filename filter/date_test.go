package filter

import (
	"strconv"
	"testing"
	"time"
)

var ts int64 = 1580212332
var tsMs int64 = 1580212332123

func TestUnix(t *testing.T) {
	for _, v := range []interface{}{ts, int(ts), "1580212332"} {
		p := &UnixParser{}
		r, err := p.Parse(v)

		if err != nil {
			t.Error(err)
		}

		t.Log(r)
		if r.Unix() != ts {
			t.Errorf("%v %d", v, r.Unix())
		}
	}
}

func TestUnixMs(t *testing.T) {
	for _, v := range []interface{}{tsMs, int(tsMs), "1580212332123"} {
		p := &UnixMSParser{}
		r, err := p.Parse(v)

		if err != nil {
			t.Fatalf("%v %s", v, err)
		}

		t.Log(r)
		t.Log(r.Unix())
		rr := r.UnixNano() / 1000000
		if rr != tsMs {
			t.Fatalf("%#v %d", v, rr)
		}
	}
}

func TestFormatParser(t *testing.T) {
	var tStr = "2020-01-28T19:52:12.123+08:00"
	p := &FormatParser{time.RFC3339, nil, false}
	r, err := p.Parse(tStr)

	if err != nil {
		t.Fatalf("%s", err)
	}

	rr := r.UnixNano() / 1000000
	if rr != tsMs {
		t.Fatalf("%v %d", tStr, rr)
	}
}

func TestFormatParserWithLocation(t *testing.T) {
	var tStr = "2020-01-28T19:52:12.123456"
	location, _ := time.LoadLocation("Asia/Shanghai")
	p := &FormatParser{"2006-01-02T15:04:05.9", location, false}
	r, err := p.Parse(tStr)

	if err != nil {
		t.Fatalf("%s", err)
	}

	rr := r.UnixNano() / 1000
	if rr != tsMs*1000+456 {
		t.Fatalf("%v %d", tStr, rr)
	}
}

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
