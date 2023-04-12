package filter

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/childe/gohangout/topology"
	"github.com/golang/glog"
)

// JSONFilter will parse json string in `field` and put the result into `target` field
type JSONFilter struct {
	field     string
	target    string
	overwrite bool
	include   []string
	exclude   []string
}

func init() {
	Register("Json", newJSONFilter)
}

func newJSONFilter(config map[interface{}]interface{}) topology.Filter {
	plugin := &JSONFilter{
		overwrite: true,
		target:    "",
	}

	if field, ok := config["field"]; ok {
		plugin.field = field.(string)
	} else {
		glog.Fatal("field must be set in Json filter")
	}

	if overwrite, ok := config["overwrite"]; ok {
		plugin.overwrite = overwrite.(bool)
	}

	if target, ok := config["target"]; ok {
		plugin.target = target.(string)
	}

	if include, ok := config["include"]; ok {
		for _, i := range include.([]interface{}) {
			plugin.include = append(plugin.include, i.(string))
		}
	}
	if exclude, ok := config["exclude"]; ok {
		for _, i := range exclude.([]interface{}) {
			plugin.exclude = append(plugin.exclude, i.(string))
		}
	}

	return plugin
}

// Filter will parse json string in `field` and put the result into `target` field
func (plugin *JSONFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	s, ok := event[plugin.field]
	if !ok {
		return event, false
	}

	ss, ok := s.(string)
	if !ok {
		return event, false
	}

	var o interface{} = nil
	d := json.NewDecoder(strings.NewReader(ss))
	d.UseNumber()
	err := d.Decode(&o)
	if err != nil || o == nil {
		return event, false
	}

	if len(plugin.include) > 0 {
		oo := map[string]interface{}{}
		if o, ok := o.(map[string]interface{}); ok {
			for _, k := range plugin.include {
				oo[k] = o[k]
			}
		} else {
			glog.V(5).Infof("%s field is not map type, could not get `include` fields from it", plugin.field)
			return event, false
		}
		o = oo
	} else if len(plugin.exclude) > 0 {
		if o, ok := o.(map[string]interface{}); ok {
			for _, k := range plugin.exclude {
				delete(o, k)
			}
		} else {
			glog.V(5).Infof("%s field is not map type, could not get `include` fields from it", plugin.field)
			return event, false
		}
	}

	if plugin.target == "" {
		if reflect.TypeOf(o).Kind() != reflect.Map {
			glog.V(5).Infof("%s field is not map type, `target` must be set in config file", plugin.field)
			return event, false
		}
		if plugin.overwrite {
			for k, v := range o.(map[string]interface{}) {
				event[k] = v
			}
		} else {
			for k, v := range o.(map[string]interface{}) {
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
