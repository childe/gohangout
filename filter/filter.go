package filter

import (
	"reflect"

	"github.com/childe/gohangout/condition_filter"
	"github.com/childe/gohangout/field_deleter"
	"github.com/childe/gohangout/field_setter"
	"github.com/childe/gohangout/output"
	"github.com/childe/gohangout/value_render"
	"github.com/golang/glog"
)

type Filter interface {
	Pass(map[string]interface{}) bool
	Filter(map[string]interface{}) (map[string]interface{}, bool)
	Process(map[string]interface{})
	PostProcess(map[string]interface{}, bool) map[string]interface{}
}

func BuildFilters(config map[string]interface{}, nextFilter Filter, outputs []output.Output) []Filter {
	var (
		rst []Filter
	)

	if fsI, ok := config["filters"]; ok {
		filtersI := fsI.([]interface{})

		rst = make([]Filter, len(filtersI))

		// build last filter plugin, pass outputs to it
		for filterTypeI, filterConfigI := range filtersI[len(filtersI)-1].(map[interface{}]interface{}) {
			filterType := filterTypeI.(string)
			glog.Infof("filter type: %s", filterType)
			filterConfig := filterConfigI.(map[interface{}]interface{})
			glog.Infof("filter config: %v", filterConfig)

			nextFilter = BuildFilter(filterType, filterConfig, nextFilter, outputs)

			rst[len(filtersI)-1] = nextFilter
			break // len(filtersI[-1]) is 1
		}

		for i := len(filtersI) - 2; i >= 0; i-- {
			for filterTypeI, filterConfigI := range filtersI[i].(map[interface{}]interface{}) {
				filterType := filterTypeI.(string)
				glog.Infof("filter type: %s", filterType)
				filterConfig := filterConfigI.(map[interface{}]interface{})
				glog.Infof("filter config: %v", filterConfig)

				filterPlugin := BuildFilter(filterType, filterConfig, nextFilter, nil)

				rst[i] = filterPlugin
				nextFilter = filterPlugin
				break // len(filtersI[i]) is 1
			}
		}
		return rst
	} else {
		return nil
	}
}

func BuildFilter(filterType string, config map[interface{}]interface{}, nextFilter Filter, outputs []output.Output) Filter {
	switch filterType {
	case "Add":
		f := NewAddFilter(config)
		f.BaseFilter.filter = f
		f.BaseFilter.nextFilter = nextFilter
		f.outputs = outputs
		return f
	case "Remove":
		f := NewRemoveFilter(config)
		f.BaseFilter.filter = f
		f.BaseFilter.nextFilter = nextFilter
		f.outputs = outputs
		return f
	case "Rename":
		f := NewRenameFilter(config)
		f.BaseFilter.filter = f
		f.BaseFilter.nextFilter = nextFilter
		f.outputs = outputs
		return f
	case "Lowercase":
		f := NewLowercaseFilter(config)
		f.BaseFilter.filter = f
		f.BaseFilter.nextFilter = nextFilter
		f.outputs = outputs
		return f
	case "Split":
		f := NewSplitFilter(config)
		f.BaseFilter.filter = f
		f.BaseFilter.nextFilter = nextFilter
		f.outputs = outputs
		return f
	case "Grok":
		f := NewGrokFilter(config)
		f.BaseFilter.filter = f
		f.BaseFilter.nextFilter = nextFilter
		f.outputs = outputs
		return f
	case "Date":
		f := NewDateFilter(config)
		f.BaseFilter.filter = f
		f.BaseFilter.nextFilter = nextFilter
		f.outputs = outputs
		return f
	case "Drop":
		f := NewDropFilter(config)
		f.BaseFilter.filter = f
		f.BaseFilter.nextFilter = nextFilter
		f.outputs = outputs
		return f
	case "Json":
		f := NewJsonFilter(config)
		f.BaseFilter.filter = f
		f.BaseFilter.nextFilter = nextFilter
		f.outputs = outputs
		return f
	case "Translate":
		f := NewTranslateFilter(config)
		f.BaseFilter.filter = f
		f.BaseFilter.nextFilter = nextFilter
		f.outputs = outputs
		return f
	case "Convert":
		f := NewConvertFilter(config)
		f.BaseFilter.filter = f
		f.BaseFilter.nextFilter = nextFilter
		f.outputs = outputs
		return f
	case "IPIP":
		f := NewIPIPFilter(config)
		f.BaseFilter.filter = f
		f.BaseFilter.nextFilter = nextFilter
		f.outputs = outputs
		return f
	case "LinkMetric":
		f := NewLinkMetricFilter(config)
		f.BaseFilter.filter = f
		f.BaseFilter.nextFilter = nextFilter
		f.outputs = outputs
		return f
	case "LinkStatsMetric":
		f := NewLinkStatsMetricFilter(config)
		f.BaseFilter.filter = f
		f.BaseFilter.nextFilter = nextFilter
		f.outputs = outputs
		return f
	case "Filters":
		f := NewFiltersFilter(config, nextFilter, outputs)
		f.BaseFilter.filter = f
		f.BaseFilter.nextFilter = nextFilter
		f.outputs = outputs
		return f
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

	filter     Filter
	nextFilter Filter
	outputs    []output.Output
}

func NewBaseFilter(config map[interface{}]interface{}) *BaseFilter {
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
	return &f
}

func (f *BaseFilter) Pass(event map[string]interface{}) bool {
	return f.conditionFilter.Pass(event)
}

func (f *BaseFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	return event, true
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

func (b *BaseFilter) Process(event map[string]interface{}) {
	var rst bool

	if b.Pass(event) {
		event, rst = b.filter.Filter(event)
		if event == nil {
			return
		}
		event = b.filter.PostProcess(event, rst)
	}

	if b.nextFilter != nil {
		b.nextFilter.Process(event)
	} else {
		for _, outputPlugin := range b.outputs {
			if outputPlugin.Pass(event) {
				outputPlugin.Emit(event)
			}
		}
	}
}
