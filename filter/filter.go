package filter

import (
	"plugin"
	"reflect"

	"github.com/childe/gohangout/condition_filter"
	"github.com/childe/gohangout/field_deleter"
	"github.com/childe/gohangout/field_setter"
	"github.com/childe/gohangout/output"
	"github.com/childe/gohangout/value_render"
	"github.com/golang/glog"
)

type Nexter interface {
	Process(map[string]interface{}) map[string]interface{}
}

type FilterNexter struct {
	Next *FilterBox
}

func (n *FilterNexter) Process(event map[string]interface{}) map[string]interface{} {
	return n.Next.Process(event)
}

type OutputNexter struct {
	Next output.Output
}

func (n *OutputNexter) Process(event map[string]interface{}) map[string]interface{} {
	if n.Next.Pass(event) {
		n.Next.Emit(event)
	}
	return nil
}

type OutputsNexter struct {
	Next []output.Output
}

func (n *OutputsNexter) Process(event map[string]interface{}) map[string]interface{} {
	for _, o := range n.Next {
		if o.Pass(event) {
			o.Emit(event)
		}
	}
	return nil
}

type NilNexter struct {
}

func (n *NilNexter) Process(event map[string]interface{}) map[string]interface{} {
	return event
}

type Filter interface {
	Filter(map[string]interface{}) (map[string]interface{}, bool)
}

func BuildFilterBoxes(config map[string]interface{}, nexter Nexter) []*FilterBox {
	if _, ok := config["filters"]; !ok {
		return nil
	}

	filtersI := config["filters"].([]interface{})
	filters := make([]Filter, len(filtersI))

	for i := 0; i < len(filters); i++ {
		for filterTypeI, filterConfigI := range filtersI[i].(map[interface{}]interface{}) {
			filterType := filterTypeI.(string)
			glog.Infof("filter type: %s", filterType)
			filterConfig := filterConfigI.(map[interface{}]interface{})
			glog.Infof("filter config: %v", filterConfig)

			filterPlugin := BuildFilter(filterType, filterConfig)

			filters[i] = filterPlugin
		}
	}

	boxes := make([]*FilterBox, len(filters))
	for i := 0; i < len(filters); i++ {
		for _, cfg := range filtersI[i].(map[interface{}]interface{}) {
			boxes[i] = NewFilterBox(cfg.(map[interface{}]interface{}))
			boxes[i].filter = filters[i]
		}
	}

	for i := 0; i < len(filters)-1; i++ {
		boxes[i].nexter = &FilterNexter{boxes[i+1]}
	}

	boxes[len(boxes)-1].nexter = nexter

	for i, filter := range filters {
		v := reflect.ValueOf(filter)
		f := v.MethodByName("SetBelongTo")
		if f.IsValid() {
			f.Call([]reflect.Value{reflect.ValueOf(boxes[i])})
		}
	}

	return boxes
}

func BuildFilter(filterType string, config map[interface{}]interface{}) Filter {
	switch filterType {
	case "Add":
		f := NewAddFilter(config)
		return f
	case "Remove":
		f := NewRemoveFilter(config)
		return f
	case "Rename":
		f := NewRenameFilter(config)
		return f
	case "Lowercase":
		f := NewLowercaseFilter(config)
		return f
	case "Uppercase":
		f := NewUppercaseFilter(config)
		return f
	case "Split":
		f := NewSplitFilter(config)
		return f
	case "Grok":
		f := NewGrokFilter(config)
		return f
	case "Date":
		f := NewDateFilter(config)
		return f
	case "Drop":
		f := NewDropFilter(config)
		return f
	case "Json":
		f := NewJsonFilter(config)
		return f
	case "Translate":
		f := NewTranslateFilter(config)
		return f
	case "Convert":
		f := NewConvertFilter(config)
		return f
	case "URLDecode":
		f := NewURLDecodeFilter(config)
		return f
	case "Replace":
		f := NewReplaceFilter(config)
		return f
	case "KV":
		f := NewKVFilter(config)
		return f
	case "IPIP":
		f := NewIPIPFilter(config)
		return f
	case "Filters":
		f := NewFiltersFilter(config)
		return f
	case "LinkMetric":
		f := NewLinkMetricFilter(config)
		return f
	case "LinkStatsMetric":
		f := NewLinkStatsMetricFilter(config)
		return f
	//case "Filters":
	//f := NewFiltersFilter(config, nextFilter, outputs)
	//return f
	default:
		p, err := plugin.Open(filterType)
		if err != nil {
			glog.Fatalf("could not open %s", filterType)
		}
		new, err := p.Lookup("New")
		if err != nil {
			glog.Fatalf("could not find New function in %s", filterType)
		}
		return new.(func(map[interface{}]interface{}) Filter)(config)
	}
}

type FilterBox struct {
	filter Filter

	nexter          Nexter
	conditionFilter *condition_filter.ConditionFilter

	config map[interface{}]interface{}

	failTag      string
	removeFields []field_deleter.FieldDeleter
	addFields    map[field_setter.FieldSetter]value_render.ValueRender
}

func NewFilterBox(config map[interface{}]interface{}) *FilterBox {
	f := FilterBox{
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

func (f *FilterBox) PostProcess(event map[string]interface{}, success bool) map[string]interface{} {
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

func (b *FilterBox) Process(event map[string]interface{}) map[string]interface{} {
	var rst bool

	if b.conditionFilter.Pass(event) {
		event, rst = b.filter.Filter(event)
		if event == nil {
			return nil
		}
		event = b.PostProcess(event, rst)
	}

	return b.nexter.Process(event)
}
