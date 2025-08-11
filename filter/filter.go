package filter

import (
	"fmt"
	"plugin"

	"github.com/childe/gohangout/topology"
	"k8s.io/klog/v2"
)

type BuildFilterFunc func(map[any]any) topology.Filter

var registeredFilter map[string]BuildFilterFunc = make(map[string]BuildFilterFunc)

// Register is used by input plugins to register themselves
func Register(filterType string, bf BuildFilterFunc) {
	if _, ok := registeredFilter[filterType]; ok {
		klog.Errorf("%s has been registered, ignore %T", filterType, bf)
		return
	}
	registeredFilter[filterType] = bf
}

// BuildFilter builds Filter from filter type and config. it firstly tries built-in filter, and then try 3rd party plugin
func BuildFilter(filterType string, config map[any]any) topology.Filter {
	if v, ok := registeredFilter[filterType]; ok {
		return v(config)
	}
	klog.Infof("could not load %s filter plugin, try third party plugin", filterType)

	pluginPath := filterType
	filter, err := getFilterFromPlugin(pluginPath, config)
	if err != nil {
		klog.Errorf("could not open %s: %s", pluginPath, err)
		return nil
	}
	return filter
}

func getFilterFromPlugin(pluginPath string, config map[any]any) (topology.Filter, error) {
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("could not open %s: %s", pluginPath, err)
	}
	newFunc, err := p.Lookup("New")
	if err != nil {
		return nil, fmt.Errorf("could not find `New` function in %s: %s", pluginPath, err)
	}

	f, ok := newFunc.(func(map[any]any) any)
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
