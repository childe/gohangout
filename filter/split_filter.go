package filter

import (
	"strings"

	"github.com/childe/gohangout/field_setter"
	"github.com/childe/gohangout/value_render"
	"github.com/golang/glog"
)

type SplitFilter struct {
	*BaseFilter

	config       map[interface{}]interface{}
	fields       []field_setter.FieldSetter
	fieldsLength int
	sep          string
	maxSplit     int
	src          value_render.ValueRender
	overwrite    bool
	ignoreBlank  bool
}

func NewSplitFilter(config map[interface{}]interface{}) *SplitFilter {
	plugin := &SplitFilter{
		BaseFilter:  NewBaseFilter(config),
		config:      config,
		fields:      make([]field_setter.FieldSetter, 0),
		overwrite:   true,
		sep:         "",
		ignoreBlank: true,
		maxSplit:    -1,
	}

	if ignoreBlank, ok := config["ignoreBlank"]; ok {
		plugin.ignoreBlank = ignoreBlank.(bool)
	}

	if overwrite, ok := config["overwrite"]; ok {
		plugin.overwrite = overwrite.(bool)
	}

	if maxSplit, ok := config["maxSplit"]; ok {
		plugin.maxSplit = maxSplit.(int)
	}

	if src, ok := config["src"]; ok {
		plugin.src = value_render.GetValueRender2(src.(string))
	} else {
		plugin.src = value_render.GetValueRender2("message")
	}

	if sep, ok := config["sep"]; ok {
		plugin.sep = sep.(string)
	}

	if plugin.sep == "" {
		glog.Fatal("sep must be set in split filter plugin")
	}

	if fieldsI, ok := config["fields"]; ok {
		for _, f := range fieldsI.([]interface{}) {
			plugin.fields = append(plugin.fields, field_setter.NewFieldSetter(f.(string)))
		}
	} else {
		glog.Fatal("fileds must be set in split filter plugin")
	}
	plugin.fieldsLength = len(plugin.fields)

	return plugin
}

func (plugin *SplitFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	src := plugin.src.Render(event)
	if src == nil {
		return event, false
	}

	values := strings.SplitN(src.(string), plugin.sep, plugin.maxSplit)

	if len(values) != plugin.fieldsLength {
		return event, false
	}

	for i, f := range plugin.fields {
		if values[i] == "" && plugin.ignoreBlank {
			continue
		}
		event = f.SetField(event, values[i], "", plugin.overwrite)
	}
	return event, true
}
