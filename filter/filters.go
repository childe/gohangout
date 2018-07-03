package filter

import "github.com/golang-collections/collections/stack"

type FiltersFilter struct {
	BaseFilter

	config  map[interface{}]interface{}
	filters []Filter
	event   map[string]interface{}
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
	plugin.filters = GetFilters(_config)
	return plugin
}

func (plugin *FiltersFilter) Process(event map[string]interface{}) (map[string]interface{}, bool) {
	plugin.event = event
	return nil, false
}

func (plugin *FiltersFilter) EmitExtraEvents(outputS *stack.Stack) []map[string]interface{} {
	var (
		event   map[string]interface{}
		events  []map[string]interface{}
		success bool

		sFrom *stack.Stack = stack.New()
		sTo   *stack.Stack = stack.New()
	)

	sFrom.Push(plugin.event)

	for _, filterPlugin := range plugin.filters {
		for sFrom.Len() > 0 {
			event = sFrom.Pop().(map[string]interface{})
			if filterPlugin.Pass(event) == false {
				sTo.Push(event)
				continue
			}
			event, success = filterPlugin.Process(event)
			if event != nil {
				event = filterPlugin.PostProcess(event, success)
				if event != nil {
					sTo.Push(event)
				}
			}
			events = filterPlugin.EmitExtraEvents(sTo)
			if events != nil {
				for _, event := range events {
					sTo.Push(event)
				}
			}
		}
		sFrom, sTo = sTo, sFrom
	}
	for sFrom.Len() > 0 {
		outputS.Push(sFrom.Pop())
	}
	return nil
}
