package filter

import (
	"fmt"
	"plugin"

	"github.com/childe/gohangout/topology"
	"github.com/golang/glog"
)

type BuildFilterFunc func(map[interface{}]interface{}) topology.Filter

var registeredFilter map[string]BuildFilterFunc = make(map[string]BuildFilterFunc)

// Register is used by input plugins to register themselves
func Register(filterType string, bf BuildFilterFunc) {
	if _, ok := registeredFilter[filterType]; ok {
		glog.Errorf("%s has been registered, ignore %T", filterType, bf)
		return
	}
	registeredFilter[filterType] = bf
}

// BuildFilter builds Filter from filter type and config. it firstly tries built-in filter, and then try 3rd party plugin
func BuildFilter(filterType string, config map[interface{}]interface{}) topology.Filter {
	if v, ok := registeredFilter[filterType]; ok {
		return v(config)
	}
	glog.Infof("could not load %s filter plugin, try third party plugin", filterType)

	pluginPath := filterType
	filter, err := getFilterFromPlugin(pluginPath, config)
	if err != nil {
		glog.Errorf("could not open %s: %s", pluginPath, err)
		return nil
	}
	return filter
}

func getFilterFromPlugin(pluginPath string, config map[interface{}]interface{}) (topology.Filter, error) {
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("could not open %s: %s", pluginPath, err)
	}
	newFunc, err := p.Lookup("New")
	if err != nil {
		return nil, fmt.Errorf("could not find `New` function in %s: %s", pluginPath, err)
	}

	f, ok := newFunc.(func(map[interface{}]interface{}) interface{})
	if !ok {
		return nil, fmt.Errorf("`New` func in %s format error", pluginPath)
	}

	rst := f(config)
	filter, ok := rst.(topology.Filter)
	if !ok {
		return nil, fmt.Errorf("`New` func in %s dose not return Filter Interface", pluginPath)
	}
	return filter, nil
}
