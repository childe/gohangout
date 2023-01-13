package filter

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/childe/gohangout/field_setter"
	"github.com/childe/gohangout/topology"
	"github.com/childe/gohangout/value_render"
	"github.com/golang/glog"
	"github.com/spf13/cast"
)

type Converter interface {
	convert(v interface{}) (interface{}, error)
}

var ErrConvertUnknownFormat error = errors.New("unknown format")

type IntConverter struct{}

func (c *IntConverter) convert(v interface{}) (interface{}, error) {
	return cast.ToInt64E(v)
}

type UIntConverter struct{}

func (c *UIntConverter) convert(v interface{}) (interface{}, error) {
	return cast.ToUint64E(v)
}

type FloatConverter struct{}

func (c *FloatConverter) convert(v interface{}) (interface{}, error) {
	return cast.ToFloat64E(v)
}

type BoolConverter struct{}

func (c *BoolConverter) convert(v interface{}) (interface{}, error) {
	if v, ok := v.(string); ok {
		rst, err := strconv.ParseBool(v)
		if err != nil {
			return nil, err
		} else {
			return rst, err
		}
	}
	return nil, ErrConvertUnknownFormat
}

type StringConverter struct{}

func (c *StringConverter) convert(v interface{}) (interface{}, error) {
	if r, ok := v.(json.Number); ok {
		return r.String(), nil
	}

	if r, ok := v.(string); ok {
		return r, nil
	}

	jsonString, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return string(jsonString), nil
}

type ArrayIntConverter struct{}

func (c *ArrayIntConverter) convert(v interface{}) (interface{}, error) {
	if v1, ok1 := v.([]interface{}); ok1 {
		var t2 = []int{}
		for _, i := range v1 {
			j, err := i.(json.Number).Int64()
			// j, err := strconv.ParseInt(i.String(), 0, 64)
			if err != nil {
				return nil, ErrConvertUnknownFormat
			}
			t2 = append(t2, (int)(j))
		}
		return t2, nil
	}
	return nil, ErrConvertUnknownFormat
}

type ArrayFloatConverter struct{}

func (c *ArrayFloatConverter) convert(v interface{}) (interface{}, error) {
	if v1, ok1 := v.([]interface{}); ok1 {
		var t2 = []float64{}
		for _, i := range v1 {
			j, err := i.(json.Number).Float64()
			if err != nil {
				return nil, ErrConvertUnknownFormat
			}
			t2 = append(t2, (float64)(j))
		}
		return t2, nil
	}
	return nil, ErrConvertUnknownFormat
}

type ConveterAndRender struct {
	converter    Converter
	valueRender  value_render.ValueRender
	removeIfFail bool
	settoIfFail  interface{}
	settoIfNil   interface{}
}

type ConvertFilter struct {
	config map[interface{}]interface{}
	fields map[field_setter.FieldSetter]ConveterAndRender
}

func init() {
	Register("Convert", newConvertFilter)
}

func newConvertFilter(config map[interface{}]interface{}) topology.Filter {
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
			setto_if_nil := v["setto_if_nil"]

			var converter Converter
			if to == "float" {
				converter = &FloatConverter{}
			} else if to == "int" {
				converter = &IntConverter{}
			} else if to == "uint" {
				converter = &UIntConverter{}
			} else if to == "bool" {
				converter = &BoolConverter{}
			} else if to == "string" {
				converter = &StringConverter{}
			} else if to == "array(int)" {
				converter = &ArrayIntConverter{}
			} else if to == "array(float)" {
				converter = &ArrayFloatConverter{}
			} else {
				glog.Fatal("can only convert to int/float/bool/array(int)/array(float)")
			}

			plugin.fields[fieldSetter] = ConveterAndRender{
				converter:    converter,
				valueRender:  value_render.GetValueRender2(f.(string)),
				removeIfFail: remove_if_fail,
				settoIfFail:  setto_if_fail,
				settoIfNil:   setto_if_nil,
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
			if conveterAndRender.settoIfNil != nil {
				event = fs.SetField(event, conveterAndRender.settoIfNil, "", true)
			}
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
