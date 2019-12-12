package filter

import (
	"net/url"
	"reflect"

	"github.com/childe/gohangout/field_setter"
	"github.com/childe/gohangout/value_render"
	"github.com/golang/glog"
)

type URLDecodeFilter struct {
	config map[interface{}]interface{}
	fields map[field_setter.FieldSetter]value_render.ValueRender
}

func (l *MethodLibrary) NewURLDecodeFilter(config map[interface{}]interface{}) *URLDecodeFilter {
	plugin := &URLDecodeFilter{
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
		glog.Fatal("fileds must be set in URLDecode filter plugin")
	}
	return plugin
}

// 如果字段不是字符串, 返回false, 其它返回true
func (plugin *URLDecodeFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	success := true
	for s, v := range plugin.fields {
		value := v.Render(event)
		if value != nil {
			if reflect.TypeOf(value).Kind() != reflect.String {
				success = false
				continue
			}
			rst, err := url.QueryUnescape(value.(string))
			if err != nil {
				success = false
				continue
			}
			s.SetField(event, rst, "", true)
		}
	}
	return event, success
}
