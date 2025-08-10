package filter

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/childe/cast"
	"github.com/childe/gohangout/field_setter"
	"github.com/childe/gohangout/topology"
	"github.com/childe/gohangout/value_render"
	"k8s.io/klog/v2"
)

type Converter interface {
	convert(v any) (any, error)
}

var ErrConvertUnknownFormat error = errors.New("unknown format")

type IntConverter struct{}

func (c *IntConverter) convert(v any) (any, error) {
	return cast.ToInt64E(v)
}

type UIntConverter struct{}

func (c *UIntConverter) convert(v any) (any, error) {
	return cast.ToUint64E(v)
}

type FloatConverter struct{}

func (c *FloatConverter) convert(v any) (any, error) {
	return cast.ToFloat64E(v)
}

type BoolConverter struct{}

func (c *BoolConverter) convert(v any) (any, error) {
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

func (c *StringConverter) convert(v any) (any, error) {
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

func (c *ArrayIntConverter) convert(v any) (any, error) {
	if v1, ok1 := v.([]any); ok1 {
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

func (c *ArrayFloatConverter) convert(v any) (any, error) {
	if v1, ok1 := v.([]any); ok1 {
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
	settoIfFail  any
	settoIfNil   any
}

// FieldConvertConfig defines the configuration for a single field conversion
type FieldConvertConfig struct {
	To           string      `json:"to"`
	RemoveIfFail bool        `json:"remove_if_fail"`
	SettoIfFail  any `json:"setto_if_fail"`
	SettoIfNil   any `json:"setto_if_nil"`
}

// ConvertConfig defines the configuration structure for Convert filter
type ConvertConfig struct {
	Fields map[string]FieldConvertConfig `json:"fields"`
}

type ConvertFilter struct {
	config map[any]any
	fields map[field_setter.FieldSetter]ConveterAndRender
}

func init() {
	Register("Convert", newConvertFilter)
}

func newConvertFilter(config map[any]any) topology.Filter {
	plugin := &ConvertFilter{
		config: config,
		fields: make(map[field_setter.FieldSetter]ConveterAndRender),
	}

	// Parse configuration using SafeDecodeConfig
	var convertConfig ConvertConfig
	
	SafeDecodeConfig("Convert", config, &convertConfig)

	// Validate required fields
	if convertConfig.Fields == nil || len(convertConfig.Fields) == 0 {
		panic("Convert filter: 'fields' is required and cannot be empty")
	}

	// Process each field conversion
	for fieldName, fieldConfig := range convertConfig.Fields {
		fieldSetter := field_setter.NewFieldSetter(fieldName)
		if fieldSetter == nil {
			panic(fmt.Sprintf("Convert filter: could not build field setter from '%s'", fieldName))
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
			panic(fmt.Sprintf("Convert filter: field '%s' has invalid 'to' value '%s'. Must be one of: int/uint/float/bool/string/array(int)/array(float)", fieldName, fieldConfig.To))
		}

		plugin.fields[fieldSetter] = ConveterAndRender{
			converter:    converter,
			valueRender:  value_render.GetValueRender2(fieldName),
			removeIfFail: fieldConfig.RemoveIfFail,
			settoIfFail:  convertJSONNumber(fieldConfig.SettoIfFail),
			settoIfNil:   convertJSONNumber(fieldConfig.SettoIfNil),
		}
	}

	return plugin
}

// convertJSONNumber converts json.Number to appropriate Go types for backward compatibility
func convertJSONNumber(value any) any {
	if jsonNum, ok := value.(json.Number); ok {
		numStr := jsonNum.String()
		// If it contains a decimal point, treat as float
		if strings.Contains(numStr, ".") {
			if floatVal, err := jsonNum.Float64(); err == nil {
				return floatVal
			}
		} else {
			// Otherwise, try to convert to int
			if intVal, err := jsonNum.Int64(); err == nil {
				return int(intVal)
			}
		}
		// If conversion fails, return as string
		return numStr
	}
	return value
}

func (plugin *ConvertFilter) Filter(event map[string]any) (map[string]any, bool) {
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
