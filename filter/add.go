package filter

import (
	"github.com/childe/gohangout/field_setter"
	"github.com/childe/gohangout/topology"
	"github.com/childe/gohangout/value_render"
	"k8s.io/klog/v2"
)

type AddFilter struct {
	config    map[interface{}]interface{}
	fields    map[field_setter.FieldSetter]value_render.ValueRender
	overwrite bool
}

func init() {
	Register("Add", newAddFilter)
}

func newAddFilter(config map[interface{}]interface{}) topology.Filter {
	plugin := &AddFilter{
		config:    config,
		fields:    make(map[field_setter.FieldSetter]value_render.ValueRender),
		overwrite: true,
	}

	if overwrite, ok := config["overwrite"]; ok {
		plugin.overwrite = overwrite.(bool)
	}

	if fieldsValue, ok := config["fields"]; ok {
		for f, v := range fieldsValue.(map[interface{}]interface{}) {
			fieldSetter := field_setter.NewFieldSetter(f.(string))
			if fieldSetter == nil {
				klog.Fatalf("could build field setter from %s", f.(string))
			}
			plugin.fields[fieldSetter] = value_render.GetValueRender(v.(string))
		}
	} else {
		klog.Fatal("fields must be set in add filter plugin")
	}
	return plugin
}

func (plugin *AddFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	for fs, r := range plugin.fields {
		v, _ := r.Render(event)
		event = fs.SetField(event, v, "", plugin.overwrite)
	}
	return event, true
}
