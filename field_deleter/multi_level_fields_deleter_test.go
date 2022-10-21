package field_deleter

import (
	"reflect"
	"testing"
)

func TestMultiLevelDelete(t *testing.T) {
	for _, c := range []struct {
		fields []string
		event  map[string]interface{}
		want   map[string]interface{}
	}{
		{
			event:  map[string]interface{}{"hostname": "xxx"},
			fields: []string{"hostname"},
			want:   map[string]interface{}{},
		},
		{
			event:  map[string]interface{}{"hostname": "xxx"},
			fields: []string{"metadata", "hostname"},
			want:   map[string]interface{}{"hostname": "xxx"},
		},
		{
			event:  map[string]interface{}{"metadata": "xxx"},
			fields: []string{"metadata", "hostname"},
			want:   map[string]interface{}{"metadata": "xxx"},
		},
		{
			event:  map[string]interface{}{"metadata": map[string]interface{}{"hostname": "xxx"}},
			fields: []string{"metadata", "hostname"},
			want:   map[string]interface{}{"metadata": map[string]interface{}{}},
		},
	} {
		deleter := NewMultiLevelFieldDeleter(c.fields)
		deleter.Delete(c.event)
		if !reflect.DeepEqual(c.event, c.want) {
			t.Errorf("got %v, want %v", c.event, c.want)
		}
	}
}
