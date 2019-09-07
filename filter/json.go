package filter

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/golang/glog"
)

type JsonFilter struct {
	field     string
	target    string
	overwrite bool
	remove_fields    []string
	simplejsonString bool

}

func NewJsonFilter(config map[interface{}]interface{}) *JsonFilter {
	plugin := &JsonFilter{
		overwrite: true,
		target:    "",
		remove_fields:    []string{},
		simplejsonString: false,

	}

	if field, ok := config["field"]; ok {
		plugin.field = field.(string)
	} else {
		glog.Fatal("field must be set in Json filter")
	}

	if overwrite, ok := config["overwrite"]; ok {
		plugin.overwrite = overwrite.(bool)
	}

	if v, ok := config["remove_fields"]; ok {
        for _,  remove_field := range v.([]interface{}) {
            plugin.remove_fields = append(plugin.remove_fields, remove_field.(string))
        }
	}

	if simplejsonString, ok := config["simplejsonString"]; ok {
        plugin.simplejsonString = simplejsonString.(bool)
    }

	if target, ok := config["target"]; ok {
		plugin.target = target.(string)
	}

	return plugin
}

func (plugin *JsonFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	if s, ok := event[plugin.field]; ok {
		if reflect.TypeOf(s).Kind() != reflect.String {
			return event, false
		}
		var o interface{} = nil
		d := json.NewDecoder(strings.NewReader(s.(string)))
		d.UseNumber()
		err := d.Decode(&o)
		if err != nil || o == nil {
			return event, false
		}

		if plugin.target == "" {
			if reflect.TypeOf(o).Kind() != reflect.Map {
				glog.Errorf("%s is not map. must set target", plugin.field)
				return event, false
			}
			if plugin.overwrite {
				for k, v := range o.(map[string]interface{}) {
					if plugin.simplejsonString {
                        switch vv := v.(type){
                            case []interface{}, map[interface{}]interface{}, map[string]interface{}:
                                mjson, _ := json.Marshal(vv)
                                event[k] = string(mjson)
                            default:
                                event[k] = fmt.Sprintf("%+v", v)
                        }
                    } else {
                        event[k] = v
                    }
				}
			} else {
				for k, v := range o.(map[string]interface{}) {
					if _, ok := event[k]; !ok {
						if plugin.simplejsonString {
							switch vv := v.(type){
								case []interface{}, map[interface{}]interface{}, map[string]interface{}:
									mjson, _ := json.Marshal(vv)
									event[k] = string(mjson)
								default:
									event[k] = fmt.Sprintf("%+v", v)
							}
	
						} else {
							event[k] = v
						}
					}
				}
			}
		} else {
			event[plugin.target] = o
		}
		// 当remove_fields 存在，删除该key
		for _, remove_field := range plugin.remove_fields {
			delete(event, remove_field)
		}
		return event, true
	} 
	return event, false
}
