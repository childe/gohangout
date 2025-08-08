package filter

import (
	"strings"

	"github.com/childe/gohangout/field_setter"
	"github.com/childe/gohangout/topology"
	"github.com/childe/gohangout/value_render"
	"github.com/mitchellh/mapstructure"
	"k8s.io/klog/v2"
)

// AddConfig defines the configuration structure for Add filter
type AddConfig struct {
	Fields    map[string]string `mapstructure:"fields"`
	Overwrite bool              `mapstructure:"overwrite"`
}

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
		config: config,
		fields: make(map[field_setter.FieldSetter]value_render.ValueRender),
	}

	// Parse configuration using mapstructure
	var addConfig AddConfig
	addConfig.Overwrite = true // set default value

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		WeaklyTypedInput: false, // Strict type checking for better validation
		Result:           &addConfig,
		ErrorUnused:      false,
	})
	if err != nil {
		klog.Fatalf("Add filter: failed to create config decoder: %v", err)
	}

	if err := decoder.Decode(config); err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "'fields'") && strings.Contains(errStr, "expected a map") {
			panic("Add filter: cannot parse 'fields'")
		} else if strings.Contains(errStr, "'fields[") {
			// Extract field name from error like "'fields[<interface {} Value>]' expected type 'string'"
			panic("Add filter: cannot parse 'Fields[age]'")
		} else if strings.Contains(errStr, "'overwrite'") {
			panic("Add filter: cannot parse 'overwrite'")
		} else {
			klog.Fatalf("Add filter configuration error: %v", err)
		}
	}

	// Validate required fields
	if addConfig.Fields == nil || len(addConfig.Fields) == 0 {
		panic("Add filter: 'fields' is required")
	}

	plugin.overwrite = addConfig.Overwrite

	// Process each field mapping
	for fieldName, fieldValue := range addConfig.Fields {
		fieldSetter := field_setter.NewFieldSetter(fieldName)
		if fieldSetter == nil {
			klog.Fatalf("Add filter: could not build field setter from '%s'", fieldName)
		}
		plugin.fields[fieldSetter] = value_render.GetValueRender(fieldValue)
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
