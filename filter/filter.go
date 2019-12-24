package filter

import (
	"fmt"
	"plugin"
	"reflect"
	"strings"

	"github.com/childe/gohangout/topology"
	"github.com/golang/glog"
)

type MethodLibrary struct{}

var methodLibrary *MethodLibrary = &MethodLibrary{}

func BuildFilter(filterType string, config map[interface{}]interface{}) topology.Filter {
	method := reflect.ValueOf(methodLibrary).MethodByName("New" + filterType + "Filter")
	if method.IsValid() {
		return method.Call([]reflect.Value{reflect.ValueOf(config)})[0].Interface().(topology.Filter)
	}

	glog.Info("use third party plugin")

	if !strings.HasSuffix(filterType, ".so") {
		filterType = filterType + ".so"
	}
	p, err := getFilterFromPlugin(filterType, config)
	if err != nil {
		glog.Fatalf("could not load plugin from %s. try %s.so", filterType, filterType)
	}
	return p
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
	if filter, ok := rst.(topology.Filter); !ok {
		return nil, fmt.Errorf("`New` func in %s dose not return Filter Interface", pluginPath)
	} else {
		return filter, nil
	}
}
