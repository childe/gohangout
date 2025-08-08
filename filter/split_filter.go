package filter

import (
	"strings"

	"github.com/childe/gohangout/field_setter"
	"github.com/childe/gohangout/topology"
	"github.com/childe/gohangout/value_render"
	"k8s.io/klog/v2"
)

// SplitConfig defines the configuration structure for Split filter
type SplitConfig struct {
	Src         string   `mapstructure:"src"`
	Sep         string   `mapstructure:"sep"`
	MaxSplit    int      `mapstructure:"maxSplit"`
	Fields      []string `mapstructure:"fields"`
	IgnoreBlank bool     `mapstructure:"ignore_blank"`
	Overwrite   bool     `mapstructure:"overwrite"`
	Trim        string   `mapstructure:"trim"`
	DynamicSep  bool     `mapstructure:"dynamicSep"`
}

type SplitFilter struct {
	config       map[interface{}]interface{}
	fields       []field_setter.FieldSetter
	fieldsLength int
	sep          string
	sepRender    value_render.ValueRender
	maxSplit     int
	trim         string
	src          value_render.ValueRender
	overwrite    bool
	ignoreBlank  bool
	dynamicSep   bool
}

func init() {
	Register("Split", newSplitFilter)
}

func newSplitFilter(config map[interface{}]interface{}) topology.Filter {
	// Parse configuration using mapstructure
	var splitConfig SplitConfig
	// Set default values
	splitConfig.Src = "message"
	splitConfig.MaxSplit = -1
	splitConfig.IgnoreBlank = true
	splitConfig.Overwrite = true

	SafeDecodeConfig("Split", config, &splitConfig)

	// Validate required fields
	ValidateRequiredFields("Split", map[string]interface{}{
		"sep":    splitConfig.Sep,
		"fields": splitConfig.Fields,
	})
	if len(splitConfig.Fields) == 0 {
		klog.Fatal("Split filter: 'fields' cannot be empty")
	}

	plugin := &SplitFilter{
		config:      config,
		fields:      make([]field_setter.FieldSetter, 0),
		overwrite:   splitConfig.Overwrite,
		sep:         splitConfig.Sep,
		trim:        splitConfig.Trim,
		ignoreBlank: splitConfig.IgnoreBlank,
		dynamicSep:  splitConfig.DynamicSep,
		maxSplit:    splitConfig.MaxSplit,
	}

	plugin.src = value_render.GetValueRender2(splitConfig.Src)

	if plugin.dynamicSep {
		plugin.sepRender = value_render.GetValueRender(plugin.sep)
	}

	// Convert field names to field setters
	for _, fieldName := range splitConfig.Fields {
		plugin.fields = append(plugin.fields, field_setter.NewFieldSetter(fieldName))
	}
	plugin.fieldsLength = len(plugin.fields)

	return plugin
}

func (plugin *SplitFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	src, err := plugin.src.Render(event)
	if err != nil || src == nil {
		return event, false
	}

	var sep string
	if plugin.dynamicSep {
		s, err := plugin.sepRender.Render(event)
		if err != nil {
			return event, false
		}
		var ok bool
		if sep, ok = s.(string); !ok {
			return event, false
		}
	} else {
		sep = plugin.sep
	}
	values := strings.SplitN(src.(string), sep, plugin.maxSplit)

	if len(values) < plugin.fieldsLength {
		return event, false
	}

	for i, f := range plugin.fields {
		if values[i] == "" && plugin.ignoreBlank {
			continue
		}
		if plugin.trim == "" {
			event = f.SetField(event, values[i], "", plugin.overwrite)
		} else {
			event = f.SetField(event, strings.Trim(values[i], plugin.trim), "", plugin.overwrite)
		}
	}
	return event, true
}
