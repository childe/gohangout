package filter

type FiltersFilter struct {
	*BaseFilter

	config  map[interface{}]interface{}
	filters []Filter
}

func NewFiltersFilter(config map[interface{}]interface{}) *FiltersFilter {
	plugin := &FiltersFilter{
		BaseFilter: NewBaseFilter(config),
		config:     config,
	}

	_config := make(map[string]interface{})
	for k, v := range config {
		_config[k.(string)] = v
	}
	//plugin.filters = BuildFilters(_config)
	return plugin
}

func (plugin *FiltersFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	return nil, false
}
