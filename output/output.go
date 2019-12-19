package output

import (
	"fmt"
	"plugin"
	"reflect"
	"strings"

	"github.com/childe/gohangout/condition_filter"
	"github.com/childe/gohangout/topology"
	"github.com/golang/glog"
)

type MethodLibrary struct{}

var methodLibrary *MethodLibrary = &MethodLibrary{}

func BuildOutput(outputType string, config map[interface{}]interface{}) *topology.OutputBox {
	var output topology.Output
	var err error

	method := reflect.ValueOf(methodLibrary).MethodByName("New" + outputType + "Output")
	if method.IsValid() {
		output = method.Call([]reflect.Value{reflect.ValueOf(config)})[0].Interface().(topology.Output)
	} else {
		glog.Info("use third party plugin")

		if !strings.HasSuffix(outputType, ".so") {
			outputType = outputType + ".so"
		}
		output, err = getOutputFromPlugin(outputType, config)
		if err != nil {
			glog.Fatal("could not load plugin from %s. try %s.so", outputType, outputType)
		}
	}

	return &topology.OutputBox{
		output,
		condition_filter.NewConditionFilter(config),
	}
}

func getOutputFromPlugin(pluginPath string, config map[interface{}]interface{}) (topology.Output, error) {
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
	if filter, ok := rst.(topology.Output); !ok {
		return nil, fmt.Errorf("`New` func in %s dose not return Output Interface", pluginPath)
	} else {
		return filter, nil
	}
}
