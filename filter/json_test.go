package filter

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestJson(t *testing.T) {
	type testCase struct {
		event   map[string]interface{}
		config  map[interface{}]interface{}
		want    map[string]interface{}
		success bool
	}

	cases := []testCase{
		{
			map[string]interface{}{
				"message": `{"a":1,"b":2}`,
				"a":       10,
			},
			map[interface{}]interface{}{
				"field":     "message",
				"overwrite": true,
			},
			map[string]interface{}{
				"message": `{"a":1,"b":2}`,
				"a":       json.Number("1"),
				"b":       json.Number("2"),
			},
			true,
		},
		{
			map[string]interface{}{
				"message": `{"a":1,"b":2}`,
				"a":       10,
			},
			map[interface{}]interface{}{
				"field":     "message",
				"overwrite": false,
			},
			map[string]interface{}{
				"message": `{"a":1,"b":2}`,
				"a":       10,
				"b":       json.Number("2"),
			},
			true,
		},
		{
			map[string]interface{}{
				"message": `{"a":1,"b":2}`,
				"a":       10,
			},
			map[interface{}]interface{}{
				"field":     "message",
				"overwrite": false,
				"target":    "c",
			},
			map[string]interface{}{
				"message": `{"a":1,"b":2}`,
				"a":       10,
				"c": map[string]interface{}{
					"a": json.Number("1"),
					"b": json.Number("2"),
				},
			},
			true,
		},
	}

	for _, c := range cases {
		f := newJSONFilter(c.config)
		got, ok := f.Filter(c.event)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("config: %#v event: %v: want %#v, got %#v", c.config, c.event, c.want, got)
		}

		if ok != c.success {
			t.Errorf("config: %#v event: %v: want %v, got %v", c.config, c.event, c.success, ok)
		}
	}
}

func TestIncludeExclude(t *testing.T) {
	type testCase struct {
		event   map[string]interface{}
		config  map[interface{}]interface{}
		want    map[string]interface{}
		success bool
	}

	cases := []testCase{
		{
			map[string]interface{}{
				"message": `{"a":1,"b":2, "c": 3}`,
			},
			map[interface{}]interface{}{
				"field":     "message",
				"overwrite": true,
				"include":   []interface{}{"a", "b"},
			},
			map[string]interface{}{
				"message": `{"a":1,"b":2, "c": 3}`,
				"a":       json.Number("1"),
				"b":       json.Number("2"),
			},
			true,
		},
		{
			map[string]interface{}{
				"message": `{"a":1,"b":2, "c": 3}`,
			},
			map[interface{}]interface{}{
				"field":     "message",
				"overwrite": true,
				"exclude":   []interface{}{"a", "b"},
			},
			map[string]interface{}{
				"message": `{"a":1,"b":2, "c": 3}`,
				"c":       json.Number("3"),
			},
			true,
		},
		{
			map[string]interface{}{
				"message": `{"a":1,"b":2, "c": 3}`,
			},
			map[interface{}]interface{}{
				"field":     "message",
				"overwrite": true,
				"include":   []interface{}{"a", "b"},
				"exclude":   []interface{}{"a", "b"},
			},
			map[string]interface{}{
				"message": `{"a":1,"b":2, "c": 3}`,
				"a":       json.Number("1"),
				"b":       json.Number("2"),
			},
			true,
		},
	}

	for _, c := range cases {
		f := newJSONFilter(c.config)
		got, ok := f.Filter(c.event)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("config: %#v event: %v: want %#v, got %#v", c.config, c.event, c.want, got)
		}

		if ok != c.success {
			t.Errorf("config: %#v event: %v: want %v, got %v", c.config, c.event, c.success, ok)
		}
	}
}
