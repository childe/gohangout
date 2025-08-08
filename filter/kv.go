package filter

import (
	"strings"

	"github.com/childe/gohangout/field_setter"
	"github.com/childe/gohangout/topology"
	"github.com/childe/gohangout/value_render"
)

// KVConfig defines the configuration structure for KV filter
type KVConfig struct {
	Src         string   `mapstructure:"src"`
	Target      string   `mapstructure:"target"`
	FieldSplit  string   `mapstructure:"field_split"`
	ValueSplit  string   `mapstructure:"value_split"`
	Trim        string   `mapstructure:"trim"`
	TrimKey     string   `mapstructure:"trim_key"`
	IncludeKeys []string `mapstructure:"include_keys"`
	ExcludeKeys []string `mapstructure:"exclude_keys"`
}

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
	// Parse configuration using mapstructure
	var kvConfig KVConfig

	SafeDecodeConfig("KV", config, &kvConfig)

	// Validate required fields
	ValidateRequiredFields("KV", map[string]interface{}{
		"src":         kvConfig.Src,
		"field_split": kvConfig.FieldSplit,
		"value_split": kvConfig.ValueSplit,
	})

	plugin := &KVFilter{
		config:      config,
		fields:      make(map[field_setter.FieldSetter]value_render.ValueRender),
		target:      kvConfig.Target,
		field_split: kvConfig.FieldSplit,
		value_split: kvConfig.ValueSplit,
		trim:        kvConfig.Trim,
		trim_key:    kvConfig.TrimKey,
	}
	
	plugin.src = value_render.GetValueRender2(kvConfig.Src)

	// Convert include_keys slice to map
	plugin.include_keys = make(map[string]bool)
	for _, key := range kvConfig.IncludeKeys {
		plugin.include_keys[key] = true
	}

	// Convert exclude_keys slice to map
	plugin.exclude_keys = make(map[string]bool)
	for _, key := range kvConfig.ExcludeKeys {
		plugin.exclude_keys[key] = true
	}

	return plugin
}

func (p *KVFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	msg, err := p.src.Render(event)
	if err != nil || msg == nil {
		return event, false
	}
	A := strings.Split(msg.(string), p.field_split)

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
