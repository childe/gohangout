package filter

import (
	"github.com/childe/gohangout/field_setter"
	"github.com/childe/gohangout/value_render"
	"github.com/golang/glog"
)

type AddFilter struct {
	config    map[interface{}]interface{}
	fields    map[field_setter.FieldSetter]value_render.ValueRender
	overwrite bool
}

func (l *MethodLibrary) NewAddFilter(config map[interface{}]interface{}) *AddFilter {
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
				glog.Fatalf("could build field setter from %s", f.(string))
			}
			plugin.fields[fieldSetter] = value_render.GetValueRender(v.(string))
		}
	} else {
		glog.Fatal("fields must be set in add filter plugin")
	}
	return plugin
}

func (plugin *AddFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	for fs, v := range plugin.fields {
		event = fs.SetField(event, v.Render(event), "", plugin.overwrite)
	}
	return event, true
}
