package filter

import (
	"plugin"
	"reflect"

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

	p, err := plugin.Open(filterType)
	if err != nil {
		glog.Fatalf("could not open %s: %s", filterType, err)
	}
	newFunc, err := p.Lookup("New")
	if err != nil {
		glog.Fatalf("could not find New function in %s: %s", filterType, err)
	}
	return newFunc.(func(map[interface{}]interface{}) interface{})(config).(topology.Filter)
}
