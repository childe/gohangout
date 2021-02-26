package filter

import (
	"strings"

	"github.com/childe/gohangout/field_setter"
	"github.com/childe/gohangout/topology"
	"github.com/childe/gohangout/value_render"
	"github.com/golang/glog"
)

type KVFilter struct {
	config       map[interface{}]interface{}
	fields       map[field_setter.FieldSetter]value_render.ValueRender
	src          value_render.ValueRender
	target       string
	field_split  string
	value_split  string
	trim         string
	trim_key     string
	include_keys map[string]bool
	exclude_keys map[string]bool
}

func init() {
	Register("KV", newKVFilter)
}

func newKVFilter(config map[interface{}]interface{}) topology.Filter {
	plugin := &KVFilter{
		config: config,
		fields: make(map[field_setter.FieldSetter]value_render.ValueRender),
	}

	if src, ok := config["src"]; ok {
		plugin.src = value_render.GetValueRender2(src.(string))
	} else {
		glog.Fatal("src must be set in kv filter")
	}

	if target, ok := config["target"]; ok {
		plugin.target = target.(string)
	} else {
		plugin.target = ""
	}

	if field_split, ok := config["field_split"]; ok {
		plugin.field_split = field_split.(string)
	} else {
		glog.Fatal("field_split must be set in kv filter")
	}

	if value_split, ok := config["value_split"]; ok {
		plugin.value_split = value_split.(string)
	} else {
		glog.Fatal("value_split must be set in kv filter")
	}

	if trim, ok := config["trim"]; ok {
		plugin.trim = trim.(string)
	} else {
		plugin.trim = ""
	}

	if trim_key, ok := config["trim_key"]; ok {
		plugin.trim_key = trim_key.(string)
	} else {
		plugin.trim_key = ""
	}

	plugin.include_keys = make(map[string]bool)
	if include_keys, ok := config["include_keys"]; ok {
		for _, k := range include_keys.([]interface{}) {
			plugin.include_keys[k.(string)] = true
		}
	}

	plugin.exclude_keys = make(map[string]bool)
	if exclude_keys, ok := config["exclude_keys"]; ok {
		for _, k := range exclude_keys.([]interface{}) {
			plugin.exclude_keys[k.(string)] = true
		}
	}

	return plugin
}

func (p *KVFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	msg := p.src.Render(event)
	if msg == nil {
		return event, false
	}
	A := strings.Split(msg.(string), p.field_split)
	if len(A) == 1 {
		return event, false
	}

	var o map[string]interface{} = event
	if p.target != "" {
		o = make(map[string]interface{})
		event[p.target] = o
	}

	var success bool = true
	var key string
	for _, kv := range A {
		a := strings.SplitN(kv, p.value_split, 2)
		if len(a) != 2 {
			success = false
			continue
		}

		key = strings.Trim(a[0], p.trim_key)

		if _, ok := p.exclude_keys[key]; ok {
			continue
		}

		if _, ok := p.include_keys[key]; len(p.include_keys) == 0 || ok {
			o[key] = strings.Trim(a[1], p.trim)
		}
	}
	return event, success
}
