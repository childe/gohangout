package filter

import (
	"github.com/childe/gohangout/field_setter"
	"github.com/golang/glog"
)

type AddFilter struct {
	BaseFilter

	config    map[interface{}]interface{}
	fields    map[field_setter.FieldSetter]interface{}
	overwrite bool
}

func NewAddFilter(config map[interface{}]interface{}) *AddFilter {
	plugin := &AddFilter{
		BaseFilter: BaseFilter{config},
		config:     config,
		fields:     make(map[field_setter.FieldSetter]interface{}),
	}

	if fieldsValue, ok := config["fields"]; ok {
		for f, v := range fieldsValue.(map[interface{}]interface{}) {
			fieldSetter := field_setter.NewFieldSetter(f.(string))
			if fieldSetter == nil {
				glog.Fatalf("could build field setter from %s", f.(string))
			}
			plugin.fields[fieldSetter] = v
		}
	} else {
		glog.Fatal("fileds must be set in add filter plugin")
	}
	return plugin
}

func (plugin *AddFilter) Process(event map[string]interface{}) (map[string]interface{}, bool) {
	for fs, v := range plugin.fields {
		event = fs.SetField(event, v, "", plugin.overwrite)
	}
	return event, true
}
