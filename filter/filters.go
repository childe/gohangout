package filter

import (
	"reflect"

	"github.com/childe/gohangout/topology"
	"github.com/golang/glog"
)

type FiltersFilter struct {
	config        map[interface{}]interface{}
	processorNode *topology.ProcessorNode
	filterBoxes   []*topology.FilterBox
}

func (l *MethodLibrary) NewFiltersFilter(config map[interface{}]interface{}) *FiltersFilter {
	f := &FiltersFilter{
		config: config,
	}

	_config := make(map[string]interface{})
	for k, v := range config {
		_config[k.(string)] = v
	}

	f.filterBoxes = topology.BuildFilterBoxes(_config, BuildFilter)
	if len(f.filterBoxes) == 0 {
		glog.Fatal("no filters configured in Filters")
	}

	for _, b := range f.filterBoxes {
		f.processorNode = topology.AppendProcessorsToLink(f.processorNode, b)
	}

	return f
}

func (f *FiltersFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	return f.processorNode.Process(event), true
}

func (f *FiltersFilter) SetBelongTo(next topology.Processor) {
	var b *topology.FilterBox = f.filterBoxes[len(f.filterBoxes)-1]
	v := reflect.ValueOf(b.Filter)
	fun := v.MethodByName("SetBelongTo")
	if fun.IsValid() {
		fun.Call([]reflect.Value{reflect.ValueOf(next)})
	}
}
