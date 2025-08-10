package filter

import (
	"fmt"
	"regexp"

	"github.com/childe/gohangout/field_setter"
	"github.com/childe/gohangout/topology"
	"github.com/childe/gohangout/value_render"
)

type rs struct {
	r value_render.ValueRender
	s field_setter.FieldSetter
}

type oneFieldConfig struct {
	rs    rs
	Field string

	Src       string
	srcRegexp *regexp.Regexp

	Repl string
}

func init() {
	Register("Gsub", newGsubFilter)
}

// GsubFilter implements topology.Filter.
type GsubFilter struct {
	fields []*oneFieldConfig
}

func newGsubFilter(config map[any]any) topology.Filter {
	gsubFilter := &GsubFilter{}
	fields, ok := config["fields"]
	if !ok {
		panic("Gsub filter: 'fields' is required")
	}

	// Extract the decoded fields
	var tempStruct struct {
		Fields []*oneFieldConfig `json:"fields"`
	}
	SafeDecodeConfig("Gsub", map[any]any{"fields": fields}, &tempStruct)
	gsubFilter.fields = tempStruct.Fields

	if len(gsubFilter.fields) == 0 {
		panic("Gsub filter: 'fields' cannot be empty")
	}

	for _, fieldConfig := range gsubFilter.fields {
		if fieldConfig.Field == "" {
			panic("Gsub filter: field 'field' is required in each field config")
		}
		if fieldConfig.Src == "" {
			panic("Gsub filter: field 'src' is required in each field config")
		}
		if fieldConfig.Repl == "" {
			panic("Gsub filter: field 'repl' is required in each field config")
		}

		fieldConfig.rs.r = value_render.GetValueRender2(fieldConfig.Field)
		fieldConfig.rs.s = field_setter.NewFieldSetter(fieldConfig.Field)

		var err error
		fieldConfig.srcRegexp, err = regexp.Compile(fieldConfig.Src)
		if err != nil {
			panic(fmt.Sprintf("Gsub filter: invalid regex pattern '%s' for field '%s': %v", fieldConfig.Src, fieldConfig.Field, err))
		}
	}

	return gsubFilter
}

// Filter implements topology.Filter.
// One field config fails if could not get src or src is not string.
// Filter returns false if either field config fails.
func (f *GsubFilter) Filter(event map[string]any) (map[string]any, bool) {
	rst := true
	for _, config := range f.fields {
		v, err := config.rs.r.Render(event)
		if err != nil || v == nil {
			rst = false
			continue
		}
		if v, ok := v.(string); !ok {
			rst = false
			continue
		} else {
			rst := config.srcRegexp.ReplaceAllString(v, config.Repl)
			config.rs.s.SetField(event, rst, config.Field, true)
		}
	}
	return event, rst
}
