package field_deleter

import (
	"reflect"
	"testing"
)

func TestOneLevelDelete(t *testing.T) {
	for _, c := range []struct {
		field string
		event map[string]any
		want  map[string]any
	}{
		{
			event: map[string]any{"hostname": "xxx"},
			field: "hostname",
			want:  map[string]any{},
		},
		{
			event: map[string]any{"hostname": "xxx"},
			field: "metadata",
			want:  map[string]any{"hostname": "xxx"},
		},
		{
			event: map[string]any{"metadata": map[string]any{"hostname": "xxx"}},
			field: "metadata",
			want:  map[string]any{},
		},
	} {
		deleter := NewOneLevelFieldDeleter(c.field)
		deleter.Delete(c.event)
		if !reflect.DeepEqual(c.event, c.want) {
			t.Errorf("got %v, want %v", c.event, c.want)
		}
	}
}
