package filter

import "github.com/childe/gohangout/topology"

type dropFilter struct {
	config map[any]any
}

func init() {
	Register("Drop", newDropFilter)
}

func newDropFilter(config map[any]any) topology.Filter {
	plugin := &dropFilter{
		config: config,
	}
	return plugin
}

func (plugin *dropFilter) Filter(event map[string]any) (map[string]any, bool) {
	return nil, true
}
