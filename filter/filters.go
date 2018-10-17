package filter

import "github.com/childe/gohangout/output"

type FiltersFilter struct {
	*BaseFilter

	config  map[interface{}]interface{}
	filters []Filter
}

func NewFiltersFilter(config map[interface{}]interface{}, nextFilter Filter, outputs []output.Output) *FiltersFilter {
	plugin := &FiltersFilter{
		BaseFilter: NewBaseFilter(config),
		config:     config,
	}

	_config := make(map[string]interface{})
	for k, v := range config {
		_config[k.(string)] = v
	}
	plugin.filters = BuildFilters(_config, nextFilter, outputs)
	return plugin
}

func (f *FiltersFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	var rst bool
	for _, filter := range f.filters {
		event, rst = filter.Filter(event)
		if event == nil {
			return nil, false
		}
		event = filter.PostProcess(event, rst)
	}
	return event, false
}
