package input

import (
	"plugin"
	"reflect"

	"github.com/childe/gohangout/topology"
	"github.com/golang/glog"
)

type MethodLibrary struct{}

var methodLibrary *MethodLibrary = &MethodLibrary{}

func GetInput(inputType string, config map[interface{}]interface{}) topology.Input {
	method := reflect.ValueOf(methodLibrary).MethodByName("New" + inputType + "Input")
	if method.IsValid() {
		return method.Call([]reflect.Value{reflect.ValueOf(config)})[0].Interface().(topology.Input)
	}

	glog.Info("use third party plugin")

	p, err := plugin.Open(inputType)
	if err != nil {
		glog.Fatalf("could not open %s: %s", inputType, err)
	}
	newFunc, err := p.Lookup("New")
	if err != nil {
		glog.Fatalf("could not find New function in %s: %s", inputType, err)
	}
	return newFunc.(func(map[interface{}]interface{}) interface{})(config).(topology.Input)
}
