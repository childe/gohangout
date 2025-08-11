package filter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/childe/gohangout/field_setter"
	"github.com/childe/gohangout/topology"
	"github.com/childe/gohangout/value_render"
	"k8s.io/klog/v2"
)

type replaceConfig struct {
	s   field_setter.FieldSetter
	v   value_render.ValueRender
	old string
	new string
	n   int
}

// ReplaceConfig defines the configuration structure for Replace filter
type ReplaceFilterConfig struct {
	Fields map[string][]any `json:"fields"`
}

type ReplaceFilter struct {
	config map[any]any
	fields []replaceConfig
}

func init() {
	Register("Replace", newReplaceFilter)
}

func newReplaceFilter(config map[any]any) topology.Filter {
	// Parse configuration using SafeDecodeConfig helper
	var replaceFilterConfig ReplaceFilterConfig

	SafeDecodeConfig("Replace", config, &replaceFilterConfig)

	// Validate required fields
	ValidateRequiredFields("Replace", map[string]any{
		"fields": replaceFilterConfig.Fields,
	})
	if len(replaceFilterConfig.Fields) == 0 {
		klog.Fatal("Replace filter: 'fields' cannot be empty")
	}

	p := &ReplaceFilter{
		config: config,
		fields: make([]replaceConfig, 0),
	}

	// Process field replacements
	for fieldName, replaceParams := range replaceFilterConfig.Fields {
		fieldSetter := field_setter.NewFieldSetter(fieldName)
		if fieldSetter == nil {
			klog.Fatalf("Replace filter: could not build field setter from '%s'", fieldName)
		}

		v := value_render.GetValueRender2(fieldName)

		// Validate parameters length and types
		if len(replaceParams) < 2 || len(replaceParams) > 3 {
			klog.Fatalf("Replace filter: field '%s' must have 2 or 3 parameters [old, new] or [old, new, count]", fieldName)
		}

		// Extract old and new strings
		oldStr, ok := replaceParams[0].(string)
		if !ok {
			klog.Fatalf("Replace filter: field '%s' parameter 1 (old) must be string, got %T", fieldName, replaceParams[0])
		}
		newStr, ok := replaceParams[1].(string)
		if !ok {
			klog.Fatalf("Replace filter: field '%s' parameter 2 (new) must be string, got %T", fieldName, replaceParams[1])
		}

		// Extract count (optional)
		count := -1
		if len(replaceParams) == 3 {
			if countFloat, ok := replaceParams[2].(float64); ok {
				count = int(countFloat)
			} else if countInt, ok := replaceParams[2].(int); ok {
				count = countInt
			} else if countNum, ok := replaceParams[2].(json.Number); ok {
				if intVal, err := countNum.Int64(); err == nil {
					count = int(intVal)
				} else {
					panic(fmt.Sprintf("Replace filter: field '%s' parameter 3 (count) must be integer, got invalid number %s", fieldName, countNum.String()))
				}
			} else {
				panic(fmt.Sprintf("Replace filter: field '%s' parameter 3 (count) must be integer, got %T", fieldName, replaceParams[2]))
			}
		}

		t := replaceConfig{
			fieldSetter,
			v,
			oldStr,
			newStr,
			count,
		}
		p.fields = append(p.fields, t)
	}

	return p
}

// 如果字段不是字符串, 返回false, 其它返回true
func (p *ReplaceFilter) Filter(event map[string]any) (map[string]any, bool) {
	success := true
	for _, f := range p.fields {
		value, err := f.v.Render(event)
		if err != nil || value == nil {
			continue
		}
		if s, ok := value.(string); ok {
			new := strings.Replace(s, f.old, f.new, f.n)
			f.s.SetField(event, new, "", true)
		} else {
			success = false
		}
	}
	return event, success
}
