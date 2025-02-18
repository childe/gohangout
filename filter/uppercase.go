package filter

import (
	"strings"

	"github.com/childe/gohangout/field_setter"
	"github.com/childe/gohangout/topology"
	"github.com/childe/gohangout/value_render"
	"k8s.io/klog/v2"
)

type UppercaseFilter struct {
	config map[interface{}]interface{}
	fields map[field_setter.FieldSetter]value_render.ValueRender
}

func init() {
	Register("Uppercase", newUppercaseFilter)
}

func newUppercaseFilter(config map[interface{}]interface{}) topology.Filter {
	plugin := &UppercaseFilter{
		config: config,
		fields: make(map[field_setter.FieldSetter]value_render.ValueRender),
	}

	if fieldsValue, ok := config["fields"]; ok {
		for _, field := range fieldsValue.([]interface{}) {
			fieldSetter := field_setter.NewFieldSetter(field.(string))
			if fieldSetter == nil {
				klog.Fatalf("could build field setter from %s", field.(string))
			}
			plugin.fields[fieldSetter] = value_render.GetValueRender2(field.(string))
		}
	} else {
		klog.Fatal("fields must be set in remove filter plugin")
	}
	return plugin
}

// 如果字段不是字符串, 返回false, 其它返回true
func (plugin *UppercaseFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	success := true
	for s, v := range plugin.fields {
		value, err := v.Render(event)
		if err != nil || value == nil {
			success = false
			continue
		}
		if t, ok := value.(string); !ok {
			success = false
			continue
		} else {
			s.SetField(event, strings.ToUpper(t), "", true)
		}
	}
	return event, success
}
