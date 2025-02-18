package filter

import (
	"github.com/childe/gohangout/field_deleter"
	"github.com/childe/gohangout/field_setter"
	"github.com/childe/gohangout/topology"
	"github.com/childe/gohangout/value_render"
	"k8s.io/klog/v2"
)

type gsd struct {
	g value_render.ValueRender
	s field_setter.FieldSetter
	d field_deleter.FieldDeleter
}

type RenameFilter struct {
	config map[interface{}]interface{}
	fields map[string]gsd
}

func init() {
	Register("Rename", newRenameFilter)
}

func newRenameFilter(config map[interface{}]interface{}) topology.Filter {
	plugin := &RenameFilter{
		config: config,
		fields: make(map[string]gsd),
	}

	if fieldsValue, ok := config["fields"]; ok {
		for src, dst := range fieldsValue.(map[interface{}]interface{}) {
			g := value_render.GetValueRender2(src.(string))
			s := field_setter.NewFieldSetter(dst.(string))
			d := field_deleter.NewFieldDeleter(src.(string))
			plugin.fields[src.(string)] = gsd{g, s, d}
		}
	} else {
		klog.Fatal("fields must be set in rename filter plugin")
	}
	return plugin
}

func (plugin *RenameFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	for _, _gsd := range plugin.fields {
		v := _gsd.g.Render(event)
		_gsd.s.SetField(event, v, "", true)
		_gsd.d.Delete(event)
	}
	return event, true
}
