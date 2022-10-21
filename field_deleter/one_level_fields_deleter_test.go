package field_deleter

import (
	"reflect"
	"testing"
)

func TestOneLevelDelete(t *testing.T) {
	for _, c := range []struct {
		field string
		event map[string]interface{}
		want  map[string]interface{}
	}{
		{
			event: map[string]interface{}{"hostname": "xxx"},
			field: "hostname",
			want:  map[string]interface{}{},
		},
		{
			event: map[string]interface{}{"hostname": "xxx"},
			field: "metadata",
			want:  map[string]interface{}{"hostname": "xxx"},
		},
		{
			event: map[string]interface{}{"metadata": map[string]interface{}{"hostname": "xxx"}},
			field: "metadata",
			want:  map[string]interface{}{},
		},
	} {
		deleter := NewOneLevelFieldDeleter(c.field)
		deleter.Delete(c.event)
		if !reflect.DeepEqual(c.event, c.want) {
			t.Errorf("got %v, want %v", c.event, c.want)
		}
	}
}
