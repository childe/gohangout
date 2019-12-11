package filter

import (
	"github.com/childe/gohangout/topology"
	"github.com/golang/glog"
)

type FiltersFilter struct {
	config        map[interface{}]interface{}
	processorNode *topology.ProcessorNode
}

func NewFiltersFilter(config map[interface{}]interface{}) *FiltersFilter {
	f := &FiltersFilter{
		config: config,
	}

	_config := make(map[string]interface{})
	for k, v := range config {
		_config[k.(string)] = v
	}

	// TODO set next topology.Processor
	filterBoxes := BuildFilterBoxes(_config, nil)
	if len(filterBoxes) == 0 {
		glog.Fatal("no filters configured in Filters")
	}

	for _, b := range filterBoxes {
		f.processorNode = topology.AppendProcessorsToLink(f.processorNode, b)
	}

	return f
}

func (f *FiltersFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	return f.processorNode.Process(event), true
}
