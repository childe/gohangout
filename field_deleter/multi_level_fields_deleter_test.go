package field_deleter

import (
	"reflect"
	"testing"
)

func TestMultiLevelDelete(t *testing.T) {
	for _, c := range []struct {
		fields []string
		event  map[string]any
		want   map[string]any
	}{
		{
			event:  map[string]any{"hostname": "xxx"},
			fields: []string{"hostname"},
			want:   map[string]any{},
		},
		{
			event:  map[string]any{"hostname": "xxx"},
			fields: []string{"metadata", "hostname"},
			want:   map[string]any{"hostname": "xxx"},
		},
		{
			event:  map[string]any{"metadata": "xxx"},
			fields: []string{"metadata", "hostname"},
			want:   map[string]any{"metadata": "xxx"},
		},
		{
			event:  map[string]any{"metadata": map[string]any{"hostname": "xxx"}},
			fields: []string{"metadata", "hostname"},
			want:   map[string]any{"metadata": map[string]any{}},
		},
	} {
		deleter := NewMultiLevelFieldDeleter(c.fields)
		deleter.Delete(c.event)
		if !reflect.DeepEqual(c.event, c.want) {
			t.Errorf("got %v, want %v", c.event, c.want)
		}
	}
}
