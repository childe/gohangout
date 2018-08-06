package filter

import (
	"strconv"

	"github.com/childe/gohangout/field_setter"
	"github.com/childe/gohangout/value_render"
	"github.com/golang/glog"
)

type Converter interface {
	convert(v string) (interface{}, error)
}

type IntConverter struct{}

func (c *IntConverter) convert(v string) (interface{}, error) {
	i, err := strconv.ParseInt(v, 0, 32)
	return (int)(i), err
}

type FloatConverter struct{}

func (c *FloatConverter) convert(v string) (interface{}, error) {
	return strconv.ParseFloat(v, 64)
}

type BoolConverter struct{}

func (c *BoolConverter) convert(v string) (interface{}, error) {
	return strconv.ParseBool(v)
}

type ConveterAndRender struct {
	converter    Converter
	valueRender  value_render.ValueRender
	removeIfFail bool
	settoIfFail  interface{}
}

type ConvertFilter struct {
	BaseFilter

	config map[interface{}]interface{}
	fields map[field_setter.FieldSetter]ConveterAndRender
}

func NewConvertFilter(config map[interface{}]interface{}) *ConvertFilter {
	plugin := &ConvertFilter{
		BaseFilter: NewBaseFilter(config),
		config:     config,
		fields:     make(map[field_setter.FieldSetter]ConveterAndRender),
	}

	if fieldsValue, ok := config["fields"]; ok {
		for f, vI := range fieldsValue.(map[interface{}]interface{}) {
			v := vI.(map[string]interface{})
			fieldSetter := field_setter.NewFieldSetter(f.(string))
			if fieldSetter == nil {
				glog.Fatalf("could build field setter from %s", f.(string))
			}

			to := v["to"].(string)
			remove_if_fail := false
			if I, ok := v["remove_if_fail"]; ok {
				remove_if_fail = I.(bool)
			}
			setto_if_fail := v["setto_if_fail"]

			var converter Converter
			if to == "float" {
				converter = &FloatConverter{}
			} else if to == "int" {
				converter = &IntConverter{}
			} else if to == "bool" {
				converter = &BoolConverter{}
			} else {
				glog.Fatal("can only convert to int/float/bool")
			}
			plugin.fields[fieldSetter] = ConveterAndRender{
				converter,
				value_render.GetValueRender2(f.(string)),
				remove_if_fail,
				setto_if_fail,
			}
		}
	} else {
		glog.Fatal("fileds must be set in convert filter plugin")
	}
	return plugin
}

func (plugin *ConvertFilter) Process(event map[string]interface{}) (map[string]interface{}, bool) {
	for fs, conveterAndRender := range plugin.fields {
		originanV := conveterAndRender.valueRender.Render(event)
		v, err := conveterAndRender.converter.convert(originanV.(string))
		if err == nil {
			event = fs.SetField(event, v, "", true)
		} else {
			glog.V(10).Infof("convert error: %s", err)
			if conveterAndRender.settoIfFail != nil {
				event = fs.SetField(event, conveterAndRender.settoIfFail, "", true)
			} else if conveterAndRender.removeIfFail {
				event = fs.SetField(event, nil, "", true)
			}
		}
	}
	return event, true
}
