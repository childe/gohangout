package filter

import (
	"reflect"
	"strings"

	"github.com/childe/gohangout/field_setter"
	"github.com/childe/gohangout/topology"
	"github.com/childe/gohangout/value_render"
	"k8s.io/klog/v2"
)

// LowercaseConfig defines the configuration structure for Lowercase filter
type LowercaseConfig struct {
	Fields []string `mapstructure:"fields"`
}

type LowercaseFilter struct {
	config map[interface{}]interface{}
	fields map[field_setter.FieldSetter]value_render.ValueRender
}

func init() {
	Register("Lowercase", newLowercaseFilter)
}

func newLowercaseFilter(config map[interface{}]interface{}) topology.Filter {
	// Parse configuration using mapstructure
	var lowercaseConfig LowercaseConfig

	SafeDecodeConfig("Lowercase", config, &lowercaseConfig)

	// Validate required fields
	ValidateRequiredFields("Lowercase", map[string]interface{}{
		"fields": lowercaseConfig.Fields,
	})
	if len(lowercaseConfig.Fields) == 0 {
		klog.Fatal("Lowercase filter: 'fields' cannot be empty")
	}

	plugin := &LowercaseFilter{
		config: config,
		fields: make(map[field_setter.FieldSetter]value_render.ValueRender),
	}

	// Create field setters and value renders
	for _, field := range lowercaseConfig.Fields {
		fieldSetter := field_setter.NewFieldSetter(field)
		if fieldSetter == nil {
			klog.Fatalf("Lowercase filter: could not build field setter from '%s'", field)
		}
		plugin.fields[fieldSetter] = value_render.GetValueRender2(field)
	}

	return plugin
}

// 如果字段不是字符串, 返回false, 其它返回true
func (plugin *LowercaseFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	success := true
	for s, v := range plugin.fields {
		value, err := v.Render(event)
		if err != nil || value != nil {
			if reflect.TypeOf(value).Kind() != reflect.String {
				success = false
				continue
			}
			s.SetField(event, strings.ToLower(value.(string)), "", true)
		}
	}
	return event, success
}
