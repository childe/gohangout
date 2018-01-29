package filter

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/golang/glog"
)

type JsonFilter struct {
	BaseFilter

	field     string
	target    string
	overwrite bool
}

func NewJsonFilter(config map[interface{}]interface{}) *JsonFilter {
	plugin := &JsonFilter{
		BaseFilter: NewBaseFilter(config),
		overwrite:  true,
		target:     "",
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

	return plugin
}

func (plugin *JsonFilter) Process(event map[string]interface{}) (map[string]interface{}, bool) {
	if s, ok := event[plugin.field]; ok {
		if reflect.TypeOf(s).Kind() != reflect.String {
			return event, false
		}
		o := make(map[string]interface{})
		d := json.NewDecoder(strings.NewReader(s.(string)))
		d.UseNumber()
		err := d.Decode(&o)
		if err != nil {
			return event, false
		}

		if plugin.target == "" {
			if plugin.overwrite {
				for k, v := range o {
					event[k] = v
				}
			} else {
				for k, v := range o {
					if _, ok := event[k]; !ok {
						event[k] = v
					}
				}
			}
		} else {
			event[plugin.target] = o
		}
		return event, true
	} else {
		return event, false
	}
	return event, false
}
