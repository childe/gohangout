package value_render

import (
	"reflect"
	"testing"
)

type mfieldsTestCase struct {
	event  map[string]any
	fields []string

	hasError bool
	want     any
}

func TestMultiLevelValueRender(t *testing.T) {
	for _, c := range []mfieldsTestCase{
		{
			event: map[string]any{
				"a": map[string]any{
					"b": map[string]any{
						"c": "c",
					},
				},
			},
			fields:   []string{"a", "b", "c"},
			hasError: false,
			want:     "c",
		},
		{
			event: map[string]any{
				"a": map[string]any{
					"b": map[string]any{
						"c": "c",
					},
				},
			},
			fields:   []string{"a", "b", "d"},
			hasError: true,
			want:     nil,
		},
		{
			event: map[string]any{
				"a": map[string]any{
					"b": map[string]any{
						"c": "c",
					},
				},
			},
			fields:   []string{"a", "b", "c", "d"},
			hasError: true,
			want:     nil,
		},
		{
			event: map[string]any{
				"a": map[string]any{
					"b": map[string]any{
						"c": "c",
					},
				},
			},
			fields:   []string{"a", "b", "c", "d", "e"},
			hasError: true,
			want:     nil,
		},
		{
			event: map[string]any{
				"a": map[string]any{
					"b": map[string]any{
						"c": 10,
					},
				},
			},
			fields:   []string{"a", "b", "c"},
			hasError: false,
			want:     10,
		},
		{
			event: map[string]any{
				"a": map[string]any{
					"b": 10,
				},
			},
			fields:   []string{"a", "b", "c"},
			hasError: true,
			want:     nil,
		},
	} {
		v := NewMultiLevelValueRender(c.fields)
		got, err := v.Render(c.event)

		if c.hasError != (err != nil) {
			t.Errorf("if has error, case: %v, want %v, got %v", c, c.hasError, err != nil)
		}

		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("case: %v, want %q, got %q", c, c.want, got)
		}
	}
}
