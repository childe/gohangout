package filter

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/childe/gohangout/topology"
	"github.com/childe/gohangout/value_render"
	"github.com/mitchellh/mapstructure"
	"k8s.io/klog/v2"
)

// JsonConfig defines the configuration structure for Json filter
type JsonConfig struct {
	Field     string   `mapstructure:"field"`
	Target    string   `mapstructure:"target"`
	Overwrite bool     `mapstructure:"overwrite"`
	Include   []string `mapstructure:"include"`
	Exclude   []string `mapstructure:"exclude"`
}

// JSONFilter will parse json string in `field` and put the result into `target` field
type JSONFilter struct {
	field     string
	vr        value_render.ValueRender
	target    string
	overwrite bool
	include   []string
	exclude   []string
}

func init() {
	Register("Json", newJSONFilter)
}

func newJSONFilter(config map[any]any) topology.Filter {
	// Parse configuration using mapstructure
	var jsonConfig JsonConfig
	// Set default values
	jsonConfig.Overwrite = true

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &jsonConfig,
		ErrorUnused:      false,
	})
	if err != nil {
		klog.Fatalf("Json filter: failed to create config decoder: %v", err)
	}

	if err := decoder.Decode(config); err != nil {
		klog.Fatalf("Json filter configuration error: %v", err)
	}

	// Validate required fields
	if jsonConfig.Field == "" {
		klog.Fatal("Json filter: 'field' is required")
	}

	plugin := &JSONFilter{
		field:     jsonConfig.Field,
		target:    jsonConfig.Target,
		overwrite: jsonConfig.Overwrite,
		include:   jsonConfig.Include,
		exclude:   jsonConfig.Exclude,
	}
	plugin.vr = value_render.GetValueRender2(plugin.field)

	return plugin
}

// Filter will parse json string in `field` and put the result into `target` field
func (plugin *JSONFilter) Filter(event map[string]any) (map[string]any, bool) {
	f, err := plugin.vr.Render(event)
	if err != nil || f == nil {
		return event, false
	}

	ss, ok := f.(string)
	if !ok {
		return event, false
	}

	var o any = nil
	d := json.NewDecoder(strings.NewReader(ss))
	d.UseNumber()
	err = d.Decode(&o)
	if err != nil || o == nil {
		return event, false
	}

	if len(plugin.include) > 0 {
		oo := map[string]any{}
		if o, ok := o.(map[string]any); ok {
			for _, k := range plugin.include {
				oo[k] = o[k]
			}
		} else {
			klog.V(5).Infof("%s field is not map type, could not get `include` fields from it", plugin.field)
			return event, false
		}
		o = oo
	} else if len(plugin.exclude) > 0 {
		if o, ok := o.(map[string]any); ok {
			for _, k := range plugin.exclude {
				delete(o, k)
			}
		} else {
			klog.V(5).Infof("%s field is not map type, could not get `include` fields from it", plugin.field)
			return event, false
		}
	}

	if plugin.target == "" {
		if reflect.TypeOf(o).Kind() != reflect.Map {
			klog.V(5).Infof("%s field is not map type, `target` must be set in config file", plugin.field)
			return event, false
		}
		if plugin.overwrite {
			for k, v := range o.(map[string]any) {
				event[k] = v
			}
		} else {
			for k, v := range o.(map[string]any) {
				if _, ok := event[k]; !ok {
					event[k] = v
				}
			}
		}
	} else {
		event[plugin.target] = o
	}
	return event, true
}
