package filter

import "github.com/childe/gohangout/topology"

type dropFilter struct {
	config map[interface{}]interface{}
}

func init() {
	Register("Drop", newDropFilter)
}

func newDropFilter(config map[interface{}]interface{}) topology.Filter {
	plugin := &dropFilter{
		config: config,
	}
	return plugin
}

func (plugin *dropFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	return nil, true
}
