package filter

import (
	"reflect"

	"github.com/childe/gohangout/condition_filter"
	"github.com/childe/gohangout/field_deleter"
	"github.com/childe/gohangout/field_setter"
	"github.com/childe/gohangout/value_render"
	"github.com/golang-collections/collections/stack"
	"github.com/golang/glog"
)

type Filter interface {
	Pass(map[string]interface{}) bool
	Process(map[string]interface{}) (map[string]interface{}, bool)
	PostProcess(map[string]interface{}, bool) map[string]interface{}
	EmitExtraEvents(*stack.Stack)
}

func GetFilters(config map[string]interface{}) []Filter {
	if filterValue, ok := config["filters"]; ok {
		rst := make([]Filter, 0)
		filters := filterValue.([]interface{})
		for _, filterValue := range filters {
			filters := filterValue.(map[interface{}]interface{})
			for k, v := range filters {
				filterType := k.(string)
				glog.Infof("filter type:%s", filterType)
				filterConfig := v.(map[interface{}]interface{})
				glog.Infof("filter config:%v", filterConfig)
				filterPlugin := GetFilter(filterType, filterConfig)
				if filterPlugin == nil {
					glog.Fatalf("could build filter plugin from type (%s)", filterType)
				}
				rst = append(rst, filterPlugin)
			}
		}
		return rst
	} else {
		return nil
	}
}

func GetFilter(filterType string, config map[interface{}]interface{}) Filter {
	switch filterType {
	case "Add":
		return NewAddFilter(config)
	case "Remove":
		return NewRemoveFilter(config)
	case "Rename":
		return NewRenameFilter(config)
	case "Lowercase":
		return NewLowercaseFilter(config)
	case "Grok":
		return NewGrokFilter(config)
	case "Date":
		return NewDateFilter(config)
	case "Drop":
		return NewDropFilter(config)
	case "Json":
		return NewJsonFilter(config)
	case "Translate":
		return NewTranslateFilter(config)
	case "Convert":
		return NewConvertFilter(config)
	case "IPIP":
		return NewIPIPFilter(config)
	case "LinkMetric":
		return NewLinkMetricFilter(config)
	case "LinkStatMetric":
		return NewLinkStatMetricFilter(config)
	case "Filters":
		return NewFiltersFilter(config)
	}
	glog.Fatalf("could not build %s filter plugin", filterType)
	return nil
}

type BaseFilter struct {
	config          map[interface{}]interface{}
	conditionFilter *condition_filter.ConditionFilter

	failTag      string
	removeFields []field_deleter.FieldDeleter
	addFields    map[field_setter.FieldSetter]value_render.ValueRender
}

func NewBaseFilter(config map[interface{}]interface{}) BaseFilter {
	f := BaseFilter{
		config:          config,
		conditionFilter: condition_filter.NewConditionFilter(config),
	}
	if v, ok := config["failTag"]; ok {
		f.failTag = v.(string)
	} else {
		f.failTag = ""
	}

	if remove_fields, ok := config["remove_fields"]; ok {
		f.removeFields = make([]field_deleter.FieldDeleter, 0)
		for _, field := range remove_fields.([]interface{}) {
			f.removeFields = append(f.removeFields, field_deleter.NewFieldDeleter(field.(string)))
		}
	} else {
		f.removeFields = nil
	}

	if add_fields, ok := config["add_fields"]; ok {
		f.addFields = make(map[field_setter.FieldSetter]value_render.ValueRender)
		for k, v := range add_fields.(map[interface{}]interface{}) {
			fieldSetter := field_setter.NewFieldSetter(k.(string))
			if fieldSetter == nil {
				glog.Fatalf("could build field setter from %s", k.(string))
			}
			f.addFields[fieldSetter] = value_render.GetValueRender(v.(string))
		}
	} else {
		f.addFields = nil
	}
	return f
}

func (f *BaseFilter) Pass(event map[string]interface{}) bool {
	return f.conditionFilter.Pass(event)
}

func (f *BaseFilter) Process(event map[string]interface{}) (map[string]interface{}, bool) {
	return event, true
}
func (f *BaseFilter) EmitExtraEvents(*stack.Stack) {
	return
}
func (f *BaseFilter) PostProcess(event map[string]interface{}, success bool) map[string]interface{} {
	if success {
		if f.removeFields != nil {
			for _, d := range f.removeFields {
				d.Delete(event)
			}
		}
		for fs, v := range f.addFields {
			event = fs.SetField(event, v.Render(event), "", false)
		}
	} else {
		if f.failTag != "" {
			if tags, ok := event["tags"]; ok {
				if reflect.TypeOf(tags).Kind() == reflect.String {
					event["tags"] = []string{tags.(string), f.failTag}
				} else if reflect.TypeOf(tags).Kind() == reflect.Array {
					event["tags"] = append(tags.([]interface{}), f.failTag)
				} else {
				}
			} else {
				event["tags"] = f.failTag
			}
		}
	}
	return event
}
