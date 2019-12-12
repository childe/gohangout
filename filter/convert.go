package filter

import (
	"encoding/json"
	"errors"
	"reflect"
	"strconv"

	"github.com/childe/gohangout/field_setter"
	"github.com/childe/gohangout/value_render"
	"github.com/golang/glog"
)

type Converter interface {
	convert(v interface{}) (interface{}, error)
}

var ConvertUnknownFormat error = errors.New("unknown format")

type IntConverter struct{}

func (c *IntConverter) convert(v interface{}) (interface{}, error) {
	if reflect.TypeOf(v).String() == "json.Number" {
		i, err := v.(json.Number).Int64()
		if err == nil {
			return (int)(i), err
		} else {
			return i, err
		}
	}
	if reflect.TypeOf(v).Kind() == reflect.String {
		i, err := strconv.ParseInt(v.(string), 0, 64)
		if err == nil {
			return (int)(i), err
		} else {
			return i, err
		}
	}
	return nil, ConvertUnknownFormat
}

type FloatConverter struct{}

func (c *FloatConverter) convert(v interface{}) (interface{}, error) {
	if reflect.TypeOf(v).String() == "json.Number" {
		return v.(json.Number).Float64()
	}
	if reflect.TypeOf(v).Kind() == reflect.String {
		return strconv.ParseFloat(v.(string), 64)
	}
	return nil, ConvertUnknownFormat
}

type BoolConverter struct{}

func (c *BoolConverter) convert(v interface{}) (interface{}, error) {
	return strconv.ParseBool(v.(string))
}

type ConveterAndRender struct {
	converter    Converter
	valueRender  value_render.ValueRender
	removeIfFail bool
	settoIfFail  interface{}
}

type ConvertFilter struct {
	config map[interface{}]interface{}
	fields map[field_setter.FieldSetter]ConveterAndRender
}

func (l *MethodLibrary) NewConvertFilter(config map[interface{}]interface{}) *ConvertFilter {
	plugin := &ConvertFilter{
		config: config,
		fields: make(map[field_setter.FieldSetter]ConveterAndRender),
	}

	if fieldsValue, ok := config["fields"]; ok {
		for f, vI := range fieldsValue.(map[interface{}]interface{}) {
			v := vI.(map[interface{}]interface{})
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

func (plugin *ConvertFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	for fs, conveterAndRender := range plugin.fields {
		originanV := conveterAndRender.valueRender.Render(event)
		if originanV == nil {
			continue
		}
		v, err := conveterAndRender.converter.convert(originanV)
		if err == nil {
			event = fs.SetField(event, v, "", true)
		} else {
			glog.V(10).Infof("convert error: %s", err)
			if conveterAndRender.removeIfFail {
				event = fs.SetField(event, nil, "", true)
			} else if conveterAndRender.settoIfFail != nil {
				event = fs.SetField(event, conveterAndRender.settoIfFail, "", true)
			}
		}
	}
	return event, true
}
