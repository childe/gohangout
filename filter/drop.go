package filter

import "github.com/childe/gohangout/topology"

type DropFilter struct {
	config map[interface{}]interface{}
}

func init() {
	Register("Drop", newDropFilter)
}

func newDropFilter(config map[interface{}]interface{}) topology.Filter {
	plugin := &DropFilter{
		config: config,
	}
	return plugin
}

func (plugin *DropFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	return nil, true
}
