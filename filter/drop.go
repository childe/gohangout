package filter

import (
	"github.com/childe/gohangout/condition_filter"
	"github.com/childe/gohangout/topology"
)

type dropFilter struct {
	config map[interface{}]interface{}
	f      *condition_filter.ConditionFilter
}

func init() {
	Register("Drop", newDropFilter)
}

func newDropFilter(config map[interface{}]interface{}) topology.Filter {
	plugin := &dropFilter{
		config: config,
		f:      condition_filter.NewConditionFilter(config),
	}
	return plugin
}

func (plugin *dropFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {

	if plugin.f.Pass(event) {
		return nil, true
	}

	return event, false
}
