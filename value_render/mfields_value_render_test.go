package value_render

import (
	"reflect"
	"testing"
)

type mfieldsTestCase struct {
	event  map[string]interface{}
	fields []string

	want interface{}
}

func TestMultiLevelValueRender(t *testing.T) {
	for _, c := range []mfieldsTestCase{
		{
			event: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": "c",
					},
				},
			},
			fields: []string{"a", "b", "c"},
			want:   "c",
		},
		{
			event: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": "c",
					},
				},
			},
			fields: []string{"a", "b", "d"},
			want:   nil,
		},
		{
			event: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": "c",
					},
				},
			},
			fields: []string{"a", "b", "c", "d"},
			want:   nil,
		},
		{
			event: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": "c",
					},
				},
			},
			fields: []string{"a", "b", "c", "d", "e"},
			want:   nil,
		},
		{
			event: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": 10,
					},
				},
			},
			fields: []string{"a", "b", "c"},
			want:   10,
		},
		{
			event: map[string]interface{}{
				"a": map[string]interface{}{
					"b": 10,
				},
			},
			fields: []string{"a", "b", "c"},
			want:   nil,
		},
	} {
		v := NewMultiLevelValueRender(c.fields)
		ans := v.Render(c.event)
		if !reflect.DeepEqual(ans, c.want) {
			t.Errorf("MultiLevelValueRender(%v) = %v, want %v", c.event, ans, c.want)
		}
	}
}
