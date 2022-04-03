package filter

import (
	"reflect"
	"testing"
)

type gsubTestCase struct {
	event  map[string]interface{}
	config map[interface{}]interface{}

	wantEvent map[string]interface{}
	wantOK    bool
}

func TestGsub(t *testing.T) {
	for _, c := range []gsubTestCase{
		{
			event: map[string]interface{}{
				"msg1": "corp/com",
				"msg2": `trip#corp?com\cn`,
			},
			config: map[interface{}]interface{}{
				"fields": []map[string]string{
					{"field": "msg1", "src": "/", "repl": "_"},
					{"field": "msg2", "src": `[\\?#-]`, "repl": "."},
				},
			},
			wantEvent: map[string]interface{}{
				"msg1": "corp_com",
				"msg2": "trip.corp.com.cn",
			},
			wantOK: true,
		},
		{
			event: map[string]interface{}{
				"msg": "corp.com",
			},
			config: map[interface{}]interface{}{
				"fields": []map[string]string{
					{"field": "msg", "src": "(^\\w+)", "repl": "xxx-$1-yyy"},
				},
			},
			wantEvent: map[string]interface{}{
				"msg": "xxx-corp-yyy.com",
			},
			wantOK: true,
		},
		{
			event: map[string]interface{}{
				"msg": map[string]interface{}{
					"data": "corp.com",
				},
			},
			config: map[interface{}]interface{}{
				"fields": []map[string]string{
					{"field": "[msg][data]", "src": "(^\\w+)", "repl": "xxx-$1-yyy"},
				},
			},
			wantEvent: map[string]interface{}{
				"msg": map[string]interface{}{
					"data": "xxx-corp-yyy.com",
				},
			},
			wantOK: true,
		},
	} {
		filter := newGsubFilter(c.config)
		event, ok := filter.Filter(c.event)
		if !reflect.DeepEqual(event, c.wantEvent) {
			t.Errorf("case %v error want %+v, got %+v", c, c.wantEvent, event)
		}
		if ok != c.wantOK {
			t.Errorf("case %v error. want %v, got %v", c, c.wantOK, ok)
		}
	}
}
