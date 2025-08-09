package topology

import (
	"reflect"

	"github.com/childe/gohangout/condition_filter"
	"github.com/childe/gohangout/field_deleter"
	"github.com/childe/gohangout/field_setter"
	"github.com/childe/gohangout/value_render"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog/v2"
)

type Filter interface {
	Filter(map[string]any) (map[string]any, bool)
}

type FilterBox struct {
	Filter Filter

	conditionFilter *condition_filter.ConditionFilter

	promCounter prometheus.Counter

	config map[any]any

	failTag      string
	removeFields []field_deleter.FieldDeleter
	addFields    map[field_setter.FieldSetter]value_render.ValueRender
}

func NewFilterBox(config map[any]any) *FilterBox {
	f := FilterBox{
		config:          config,
		conditionFilter: condition_filter.NewConditionFilter(config),
		promCounter:     GetPromCounter(config),
	}

	if v, ok := config["failTag"]; ok {
		f.failTag = v.(string)
	} else {
		f.failTag = ""
	}

	if remove_fields, ok := config["remove_fields"]; ok {
		f.removeFields = make([]field_deleter.FieldDeleter, 0)
		for _, field := range remove_fields.([]any) {
			f.removeFields = append(f.removeFields, field_deleter.NewFieldDeleter(field.(string)))
		}
	} else {
		f.removeFields = nil
	}

	if add_fields, ok := config["add_fields"]; ok {
		f.addFields = make(map[field_setter.FieldSetter]value_render.ValueRender)
		for k, v := range add_fields.(map[any]any) {
			fieldSetter := field_setter.NewFieldSetter(k.(string))
			if fieldSetter == nil {
				klog.Fatalf("could build field setter from %s", k.(string))
			}
			f.addFields[fieldSetter] = value_render.GetValueRender(v.(string))
		}
	} else {
		f.addFields = nil
	}
	return &f
}

func (f *FilterBox) PostProcess(event map[string]any, success bool) map[string]any {
	if success {
		for fs, r := range f.addFields {
			v, _ := r.Render(event)
			event = fs.SetField(event, v, "", false)
		}
		if f.removeFields != nil {
			for _, d := range f.removeFields {
				d.Delete(event)
			}
		}
	} else {
		if f.failTag != "" {
			if tags, ok := event["tags"]; ok {
				if reflect.TypeOf(tags).Kind() == reflect.String {
					event["tags"] = []any{tags.(string), f.failTag}
				} else if reflect.TypeOf(tags).Kind() == reflect.Array {
					event["tags"] = append(tags.([]any), f.failTag)
				}
			} else {
				event["tags"] = f.failTag
			}
		}
	}
	return event
}

func (b *FilterBox) Process(event map[string]any) map[string]any {
	var rst bool

	if b.conditionFilter.Pass(event) {
		if b.promCounter != nil {
			b.promCounter.Inc()
		}
		event, rst = b.Filter.Filter(event)
		if event == nil {
			return nil
		}
		event = b.PostProcess(event, rst)
	}
	return event
}

type buildFilterFunc func(filterType string, config map[any]any) Filter

func BuildFilterBoxes(config map[string]any, buildFilter buildFilterFunc) []*FilterBox {
	if _, ok := config["filters"]; !ok {
		return nil
	}

	filtersI := config["filters"].([]any)
	filters := make([]Filter, len(filtersI))

	for i := 0; i < len(filters); i++ {
		for filterTypeI, filterConfigI := range filtersI[i].(map[any]any) {
			filterType := filterTypeI.(string)
			klog.Infof("filter type: %s", filterType)
			filterConfig := filterConfigI.(map[any]any)
			klog.Infof("filter config: %v", filterConfig)

			filterPlugin := buildFilter(filterType, filterConfig)

			filters[i] = filterPlugin
		}
	}

	boxes := make([]*FilterBox, len(filters))
	for i := 0; i < len(filters); i++ {
		for _, cfg := range filtersI[i].(map[any]any) {
			boxes[i] = NewFilterBox(cfg.(map[any]any))
			boxes[i].Filter = filters[i]
		}
	}

	return boxes
}
