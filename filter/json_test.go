package filter

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestJson(t *testing.T) {
	type testCase struct {
		event   map[string]any
		config  map[any]any
		want    map[string]any
		success bool
	}

	cases := []testCase{
		{
			map[string]any{
				"message": `{"a":1,"b":2}`,
				"a":       10,
			},
			map[any]any{
				"field":     "message",
				"overwrite": true,
			},
			map[string]any{
				"message": `{"a":1,"b":2}`,
				"a":       json.Number("1"),
				"b":       json.Number("2"),
			},
			true,
		},
		{
			map[string]any{
				"message": `{"a":1,"b":2}`,
				"a":       10,
			},
			map[any]any{
				"field":     "message",
				"overwrite": false,
			},
			map[string]any{
				"message": `{"a":1,"b":2}`,
				"a":       10,
				"b":       json.Number("2"),
			},
			true,
		},
		{
			map[string]any{
				"message": `{"a":1,"b":2}`,
				"a":       10,
			},
			map[any]any{
				"field":     "message",
				"overwrite": false,
				"target":    "c",
			},
			map[string]any{
				"message": `{"a":1,"b":2}`,
				"a":       10,
				"c": map[string]any{
					"a": json.Number("1"),
					"b": json.Number("2"),
				},
			},
			true,
		},
		{
			map[string]any{
				"message": `{"message":"hello","b":2}`,
				"a":       10,
			},
			map[any]any{
				"field":     "message",
				"overwrite": true,
			},
			map[string]any{
				"message": "hello",
				"a":       10,
				"b":       json.Number("2"),
			},
			true,
		},
		{
			map[string]any{
				"message": `{"message":"hello","b":2}`,
				"a":       10,
			},
			map[any]any{
				"field":     "$.message",
				"overwrite": false,
			},
			map[string]any{
				"message": `{"message":"hello","b":2}`,
				"a":       10,
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

func TestIncludeExclude(t *testing.T) {
	type testCase struct {
		event   map[string]any
		config  map[any]any
		want    map[string]any
		success bool
	}

	cases := []testCase{
		{
			map[string]any{
				"message": `{"a":1,"b":2, "c": 3}`,
			},
			map[any]any{
				"field":     "message",
				"overwrite": true,
				"include":   []any{"a", "b"},
			},
			map[string]any{
				"message": `{"a":1,"b":2, "c": 3}`,
				"a":       json.Number("1"),
				"b":       json.Number("2"),
			},
			true,
		},
		{
			map[string]any{
				"message": `{"a":1,"b":2, "c": 3}`,
			},
			map[any]any{
				"field":     "message",
				"overwrite": true,
				"exclude":   []any{"a", "b"},
			},
			map[string]any{
				"message": `{"a":1,"b":2, "c": 3}`,
				"c":       json.Number("3"),
			},
			true,
		},
		{
			map[string]any{
				"message": `{"a":1,"b":2, "c": 3}`,
			},
			map[any]any{
				"field":     "message",
				"overwrite": true,
				"include":   []any{"a", "b"},
				"exclude":   []any{"a", "b"},
			},
			map[string]any{
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
