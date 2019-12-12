package filter

type DropFilter struct {
	config map[interface{}]interface{}
}

func (l *MethodLibrary) NewDropFilter(config map[interface{}]interface{}) *DropFilter {
	plugin := &DropFilter{
		config: config,
	}
	return plugin
}

func (plugin *DropFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	return nil, true
}
