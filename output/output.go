package output

import (
	"plugin"
	"reflect"

	"github.com/childe/gohangout/condition_filter"
	"github.com/childe/gohangout/topology"
	"github.com/golang/glog"
)

type MethodLibrary struct{}

var methodLibrary *MethodLibrary = &MethodLibrary{}

func BuildOutput(outputType string, config map[interface{}]interface{}) *topology.OutputBox {
	var output topology.Output

	method := reflect.ValueOf(methodLibrary).MethodByName("New" + outputType + "Output")
	if method.IsValid() {
		output = method.Call([]reflect.Value{reflect.ValueOf(config)})[0].Interface().(topology.Output)
	} else {
		glog.Info("use third party plugin")

		p, err := plugin.Open(outputType)
		if err != nil {
			glog.Fatalf("could not open %s: %s", outputType, err)
		}
		newFunc, err := p.Lookup("New")
		if err != nil {
			glog.Fatalf("could not find New function in %s: %s", outputType, err)
		}
		output = newFunc.(func(map[interface{}]interface{}) interface{})(config).(topology.Output)
	}

	return &topology.OutputBox{
		output,
		condition_filter.NewConditionFilter(config),
	}
}
