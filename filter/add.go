package filter

import (
	"fmt"

	"github.com/childe/gohangout/field_setter"
	"github.com/childe/gohangout/topology"
	"github.com/childe/gohangout/value_render"
)

// AddConfig defines the configuration structure for Add filter
type AddConfig struct {
	Fields    map[string]string `json:"fields"`
	Overwrite bool              `json:"overwrite"`
}

type AddFilter struct {
	config    map[any]any
	fields    map[field_setter.FieldSetter]value_render.ValueRender
	overwrite bool
}

func init() {
	Register("Add", newAddFilter)
}

func newAddFilter(config map[any]any) topology.Filter {
	plugin := &AddFilter{
		config: config,
		fields: make(map[field_setter.FieldSetter]value_render.ValueRender),
	}

	// Parse configuration using SafeDecodeConfig helper
	var addConfig AddConfig
	addConfig.Overwrite = true // set default value

	SafeDecodeConfig("Add", config, &addConfig)

	// Validate required fields
	if len(addConfig.Fields) == 0 {
		panic("Add filter: 'fields' is required")
	}

	plugin.overwrite = addConfig.Overwrite

	// Process each field mapping
	for fieldName, fieldValue := range addConfig.Fields {
		fieldSetter := field_setter.NewFieldSetter(fieldName)
		if fieldSetter == nil {
			panic(fmt.Sprintf("Add filter: could not build field setter from '%s'", fieldName))
		}
		plugin.fields[fieldSetter] = value_render.GetValueRender(fieldValue)
	}

	return plugin
}

func (plugin *AddFilter) Filter(event map[string]any) (map[string]any, bool) {
	for fs, r := range plugin.fields {
		v, _ := r.Render(event)
		event = fs.SetField(event, v, "", plugin.overwrite)
	}
	return event, true
}
