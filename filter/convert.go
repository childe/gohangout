package filter

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/childe/cast"
	"github.com/childe/gohangout/field_setter"
	"github.com/childe/gohangout/topology"
	"github.com/childe/gohangout/value_render"
	"github.com/mitchellh/mapstructure"
	"k8s.io/klog/v2"
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

// FieldConvertConfig defines the configuration for a single field conversion
type FieldConvertConfig struct {
	To           string      `mapstructure:"to"`
	RemoveIfFail bool        `mapstructure:"remove_if_fail"`
	SettoIfFail  interface{} `mapstructure:"setto_if_fail"`
	SettoIfNil   interface{} `mapstructure:"setto_if_nil"`
}

// ConvertConfig defines the configuration structure for Convert filter
type ConvertConfig struct {
	Fields map[string]FieldConvertConfig `mapstructure:"fields"`
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

	// Parse configuration using mapstructure
	var convertConfig ConvertConfig
	
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &convertConfig,
		ErrorUnused:      false,
	})
	if err != nil {
		klog.Fatalf("Convert filter: failed to create config decoder: %v", err)
	}

	if err := decoder.Decode(config); err != nil {
		klog.Fatalf("Convert filter configuration error: %v", err)
	}

	// Validate required fields
	if convertConfig.Fields == nil || len(convertConfig.Fields) == 0 {
		klog.Fatal("Convert filter: 'fields' is required and cannot be empty")
	}

	// Process each field conversion
	for fieldName, fieldConfig := range convertConfig.Fields {
		fieldSetter := field_setter.NewFieldSetter(fieldName)
		if fieldSetter == nil {
			klog.Fatalf("Convert filter: could not build field setter from '%s'", fieldName)
		}

		// Validate and get converter
		var converter Converter
		switch fieldConfig.To {
		case "float":
			converter = &FloatConverter{}
		case "int":
			converter = &IntConverter{}
		case "uint":
			converter = &UIntConverter{}
		case "bool":
			converter = &BoolConverter{}
		case "string":
			converter = &StringConverter{}
		case "array(int)":
			converter = &ArrayIntConverter{}
		case "array(float)":
			converter = &ArrayFloatConverter{}
		default:
			klog.Fatalf("Convert filter: field '%s' has invalid 'to' value '%s'. Must be one of: int/uint/float/bool/string/array(int)/array(float)", fieldName, fieldConfig.To)
		}

		plugin.fields[fieldSetter] = ConveterAndRender{
			converter:    converter,
			valueRender:  value_render.GetValueRender2(fieldName),
			removeIfFail: fieldConfig.RemoveIfFail,
			settoIfFail:  fieldConfig.SettoIfFail,
			settoIfNil:   fieldConfig.SettoIfNil,
		}
	}

	return plugin
}

func (plugin *ConvertFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	for fs, conveterAndRender := range plugin.fields {
		originanV, err := conveterAndRender.valueRender.Render(event)
		if err != nil || originanV == nil {
			if conveterAndRender.settoIfNil != nil {
				event = fs.SetField(event, conveterAndRender.settoIfNil, "", true)
			}
			continue
		}
		v, err := conveterAndRender.converter.convert(originanV)
		if err == nil {
			event = fs.SetField(event, v, "", true)
		} else {
			klog.V(10).Infof("convert error: %s", err)
			if conveterAndRender.removeIfFail {
				event = fs.SetField(event, nil, "", true)
			} else if conveterAndRender.settoIfFail != nil {
				event = fs.SetField(event, conveterAndRender.settoIfFail, "", true)
			}
		}
	}
	return event, true
}
