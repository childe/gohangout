package filter

import (
	"github.com/childe/gohangout/field_deleter"
	"github.com/childe/gohangout/field_setter"
	"github.com/childe/gohangout/topology"
	"github.com/childe/gohangout/value_render"
	"k8s.io/klog/v2"
)

type gsd struct {
	g value_render.ValueRender
	s field_setter.FieldSetter
	d field_deleter.FieldDeleter
}

// RenameConfig defines the configuration structure for Rename filter
type RenameConfig struct {
	Fields map[string]string `mapstructure:"fields"`
}

type RenameFilter struct {
	config map[any]any
	fields map[string]gsd
}

func init() {
	Register("Rename", newRenameFilter)
}

func newRenameFilter(config map[any]any) topology.Filter {
	// Parse configuration using mapstructure
	var renameConfig RenameConfig

	SafeDecodeConfig("Rename", config, &renameConfig)

	// Validate required fields
	ValidateRequiredFields("Rename", map[string]any{
		"fields": renameConfig.Fields,
	})
	if len(renameConfig.Fields) == 0 {
		klog.Fatal("Rename filter: 'fields' cannot be empty")
	}

	plugin := &RenameFilter{
		config: config,
		fields: make(map[string]gsd),
	}

	// Process field mappings
	for src, dst := range renameConfig.Fields {
		g := value_render.GetValueRender2(src)
		s := field_setter.NewFieldSetter(dst)
		d := field_deleter.NewFieldDeleter(src)
		plugin.fields[src] = gsd{g, s, d}
	}

	return plugin
}

func (plugin *RenameFilter) Filter(event map[string]any) (map[string]any, bool) {
	for _, _gsd := range plugin.fields {
		v, err := _gsd.g.Render(event)
		if err == nil {
			_gsd.s.SetField(event, v, "", true)
			_gsd.d.Delete(event)
		}
	}
	return event, true
}
