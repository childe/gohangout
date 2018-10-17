package filter

type DropFilter struct {
	*BaseFilter
	config map[interface{}]interface{}
}

func NewDropFilter(config map[interface{}]interface{}) *DropFilter {
	plugin := &DropFilter{
		BaseFilter: NewBaseFilter(config),
		config:     config,
	}
	return plugin
}

func (plugin *DropFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	return nil, true
}
