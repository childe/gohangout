package filter

import (
	"regexp"

	"github.com/childe/gohangout/field_setter"
	"github.com/childe/gohangout/topology"
	"github.com/childe/gohangout/value_render"
	"github.com/golang/glog"
	"github.com/mitchellh/mapstructure"
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

func newGsubFilter(config map[interface{}]interface{}) topology.Filter {
	gsubFilter := &GsubFilter{}
	fields, ok := config["fields"]
	if !ok {
		glog.Fatal("fields must be set in gsub filter")
	}

	err := mapstructure.Decode(fields, &gsubFilter.fields)
	if err != nil {
		glog.Fatal("decode fields config in gusb error:", err)
	}

	for _, config := range gsubFilter.fields {
		config.rs.r = value_render.GetValueRender2(config.Field)
		config.rs.s = field_setter.NewFieldSetter(config.Field)

		config.srcRegexp = regexp.MustCompile(config.Src)
	}

	return gsubFilter
}

// Filter implements topology.Filter.
// One field config fails if could not get src or src is not string.
// Filter returns false if either field config fails.
func (f *GsubFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	rst := true
	for _, config := range f.fields {
		v := config.rs.r.Render(event)
		if v == nil {
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
