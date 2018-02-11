package filter

import (
	"reflect"

	"github.com/childe/gohangout/condition_filter"
	"github.com/golang/glog"
)

type Filter interface {
	Pass(map[string]interface{}) bool
	Process(map[string]interface{}) (map[string]interface{}, bool)
	PostProcess(map[string]interface{}, bool) map[string]interface{}
	EmitExtraEvents() []map[string]interface{}
}

func GetFilter(filterType string, config map[interface{}]interface{}) Filter {
	switch filterType {
	case "Add":
		return NewAddFilter(config)
	case "Grok":
		return NewGrokFilter(config)
	case "Date":
		return NewDateFilter(config)
	case "Drop":
		return NewDropFilter(config)
	case "Json":
		return NewJsonFilter(config)
	}
	glog.Fatalf("could build %s filter plugin", filterType)
	return nil
}

type BaseFilter struct {
	config          map[interface{}]interface{}
	conditionFilter *condition_filter.ConditionFilter
}

func NewBaseFilter(config map[interface{}]interface{}) BaseFilter {
	f := BaseFilter{
		config:          config,
		conditionFilter: condition_filter.NewConditionFilter(config),
	}
	return f
}

func (f *BaseFilter) Pass(event map[string]interface{}) bool {
	return f.conditionFilter.Pass(event)
}

func (f *BaseFilter) Process(event map[string]interface{}) (map[string]interface{}, bool) {
	return event, true
}
func (f *BaseFilter) EmitExtraEvents() []map[string]interface{} {
	return nil
}
func (f *BaseFilter) PostProcess(event map[string]interface{}, success bool) map[string]interface{} {
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
