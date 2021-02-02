package filter

import (
	"plugin"
	"strings"

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

func BuildFilter(filterType string, config map[interface{}]interface{}) topology.Filter {
	if v, ok := registeredFilter[filterType]; ok {
		return v(config)
	}
	glog.Infof("could not load %s filter plugin, try third party plugin", filterType)

	pluginPath := filterType
	if !strings.HasSuffix(pluginPath, ".so") {
		pluginPath = filterType + ".so"
	}
	_, err := plugin.Open(pluginPath)
	if err != nil {
		glog.Errorf("could not open %s: %s", pluginPath, err)
		return nil
	}

	if v, ok := registeredFilter[filterType]; ok {
		return v(config)
	}
	glog.Errorf("could not load %s filter plugin", filterType)
	return nil
}
