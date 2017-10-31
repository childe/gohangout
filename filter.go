package main

import "reflect"

type Filter interface {
	process(map[string]interface{}) (map[string]interface{}, bool)
	postProcess(map[string]interface{}, bool) map[string]interface{}
}

func getFilter(filterType string, config map[interface{}]interface{}) Filter {
	switch filterType {
	case "Add":
		return NewAddFilter(config)
	case "Grok":
		return NewGrokFilter(config)
	}
	return nil
}

type BaseFilter struct {
	config map[interface{}]interface{}
}

func (f *BaseFilter) process(event map[string]interface{}) (map[string]interface{}, bool) {
	return event, true
}
func (f *BaseFilter) postProcess(event map[string]interface{}, success bool) map[string]interface{} {
	if success {
		if remove_fields, ok := f.config["remove_fields"]; ok {
			for _, field := range remove_fields.([]interface{}) {
				delete(event, field.(string))
			}
		}
	} else {
		if v, ok := f.config["failTag"]; ok {
			failTag := v.(string)
			if tags, ok := event["tags"]; ok {
				if reflect.TypeOf(tags).Kind() == reflect.String {
					event["tags"] = []string{tags.(string), failTag}
				} else if reflect.TypeOf(tags).Kind() == reflect.Array {
					event["tags"] = append(tags.([]interface{}), failTag)
				} else {
				}
			} else {
				event["tags"] = failTag
			}
		}
	}
	return event
}
