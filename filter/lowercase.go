package filter

import (
	"reflect"
	"strings"

	"github.com/childe/gohangout/field_setter"
	"github.com/childe/gohangout/value_render"
	"github.com/golang/glog"
)

type LowercaseFilter struct {
	config map[interface{}]interface{}
	fields map[field_setter.FieldSetter]value_render.ValueRender
}

func (l *MethodLibrary) NewLowercaseFilter(config map[interface{}]interface{}) *LowercaseFilter {
	plugin := &LowercaseFilter{
		config: config,
		fields: make(map[field_setter.FieldSetter]value_render.ValueRender),
	}

	if fieldsValue, ok := config["fields"]; ok {
		for _, field := range fieldsValue.([]interface{}) {
			fieldSetter := field_setter.NewFieldSetter(field.(string))
			if fieldSetter == nil {
				glog.Fatalf("could build field setter from %s", field.(string))
			}
			plugin.fields[fieldSetter] = value_render.GetValueRender2(field.(string))
		}
	} else {
		glog.Fatal("fileds must be set in remove filter plugin")
	}
	return plugin
}

// 如果字段不是字符串, 返回false, 其它返回true
func (plugin *LowercaseFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	success := true
	for s, v := range plugin.fields {
		value := v.Render(event)
		if value != nil {
			if reflect.TypeOf(value).Kind() != reflect.String {
				success = false
				continue
			}
			s.SetField(event, strings.ToLower(value.(string)), "", true)
		}
	}
	return event, success
}
