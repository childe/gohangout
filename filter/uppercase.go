package filter

import (
	"strings"

	"github.com/childe/gohangout/field_setter"
	"github.com/childe/gohangout/topology"
	"github.com/childe/gohangout/value_render"
	"k8s.io/klog/v2"
)

// UppercaseConfig defines the configuration structure for Uppercase filter
type UppercaseConfig struct {
	Fields []string `mapstructure:"fields"`
}

type UppercaseFilter struct {
	config map[any]any
	fields map[field_setter.FieldSetter]value_render.ValueRender
}

func init() {
	Register("Uppercase", newUppercaseFilter)
}

func newUppercaseFilter(config map[any]any) topology.Filter {
	// Parse configuration using mapstructure
	var uppercaseConfig UppercaseConfig

	SafeDecodeConfig("Uppercase", config, &uppercaseConfig)

	// Validate required fields
	ValidateRequiredFields("Uppercase", map[string]any{
		"fields": uppercaseConfig.Fields,
	})
	if len(uppercaseConfig.Fields) == 0 {
		klog.Fatal("Uppercase filter: 'fields' cannot be empty")
	}

	plugin := &UppercaseFilter{
		config: config,
		fields: make(map[field_setter.FieldSetter]value_render.ValueRender),
	}

	// Create field setters and value renders
	for _, field := range uppercaseConfig.Fields {
		fieldSetter := field_setter.NewFieldSetter(field)
		if fieldSetter == nil {
			klog.Fatalf("Uppercase filter: could not build field setter from '%s'", field)
		}
		plugin.fields[fieldSetter] = value_render.GetValueRender2(field)
	}

	return plugin
}

// 如果字段不是字符串, 返回false, 其它返回true
func (plugin *UppercaseFilter) Filter(event map[string]any) (map[string]any, bool) {
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
