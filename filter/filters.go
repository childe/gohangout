package filter

import "github.com/golang/glog"

type FiltersFilter struct {
	config      map[interface{}]interface{}
	filterBoxes []*FilterBox
}

func NewFiltersFilter(config map[interface{}]interface{}) *FiltersFilter {
	f := &FiltersFilter{
		config: config,
	}

	_config := make(map[string]interface{})
	for k, v := range config {
		_config[k.(string)] = v
	}
	f.filterBoxes = BuildFilterBoxes(_config, nil)
	if len(f.filterBoxes) == 0 {
		glog.Fatal("no filters configured in Filters")
	}
	return f
}

func (f *FiltersFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	return f.filterBoxes[0].Process(event), true
}
