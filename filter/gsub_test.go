package filter

import (
	"reflect"
	"testing"
)

type gsubTestCase struct {
	event  map[string]any
	config map[any]any

	wantEvent map[string]any
	wantOK    bool
}

func TestGsub(t *testing.T) {
	for _, c := range []gsubTestCase{
		{
			event: map[string]any{
				"msg1": "corp/com",
				"msg2": `trip#corp?com\cn`,
			},
			config: map[any]any{
				"fields": []map[string]string{
					{"field": "msg1", "src": "/", "repl": "_"},
					{"field": "msg2", "src": `[\\?#-]`, "repl": "."},
				},
			},
			wantEvent: map[string]any{
				"msg1": "corp_com",
				"msg2": "trip.corp.com.cn",
			},
			wantOK: true,
		},
		{
			event: map[string]any{
				"msg": "corp.com",
			},
			config: map[any]any{
				"fields": []map[string]string{
					{"field": "msg", "src": "(^\\w+)", "repl": "xxx-$1-yyy"},
				},
			},
			wantEvent: map[string]any{
				"msg": "xxx-corp-yyy.com",
			},
			wantOK: true,
		},
		{
			event: map[string]any{
				"msg": map[string]any{
					"data": "corp.com",
				},
			},
			config: map[any]any{
				"fields": []map[string]string{
					{"field": "[msg][data]", "src": "(^\\w+)", "repl": "xxx-$1-yyy"},
				},
			},
			wantEvent: map[string]any{
				"msg": map[string]any{
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
