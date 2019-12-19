package input

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

func GetInput(inputType string, config map[interface{}]interface{}) topology.Input {
	method := reflect.ValueOf(methodLibrary).MethodByName("New" + inputType + "Input")
	if method.IsValid() {
		return method.Call([]reflect.Value{reflect.ValueOf(config)})[0].Interface().(topology.Input)
	}

	glog.Info("use third party plugin")

	if !strings.HasSuffix(inputType, ".so") {
		inputType = inputType + ".so"
	}
	p, err := getInputFromPlugin(inputType, config)
	if err != nil {
		glog.Fatal("could not load plugin from %s. try %s.so", inputType, inputType)
	}
	return p
}

func getInputFromPlugin(pluginPath string, config map[interface{}]interface{}) (topology.Input, error) {
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("could not open %s: %s", pluginPath, err)
	}
	newFunc, err := p.Lookup("New")
	if err != nil {
		return nil, fmt.Errorf("could not find New function in %s: %s", pluginPath, err)
	}
	f, ok := newFunc.(func(map[interface{}]interface{}) interface{})
	if !ok {
		return nil, fmt.Errorf("`New` func in %s format error", pluginPath)
	}
	rst := f(config)
	if input, ok := rst.(topology.Input); !ok {
		return nil, fmt.Errorf("`New` func in %s dose not return Input Interface", pluginPath)
	} else {
		return input, nil
	}
}
