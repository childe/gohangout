package output

import (
	"fmt"
	"plugin"

	"github.com/childe/gohangout/condition_filter"
	"github.com/childe/gohangout/topology"
	"github.com/golang/glog"
)

type BuildOutputFunc func(map[interface{}]interface{}) topology.Output

var registeredOutput map[string]BuildOutputFunc = make(map[string]BuildOutputFunc)

// Register is used by output plugins to register themselves
func Register(outputType string, bf BuildOutputFunc) {
	if _, ok := registeredOutput[outputType]; ok {
		glog.Errorf("%s has been registered, ignore %T", outputType, bf)
		return
	}
	registeredOutput[outputType] = bf
}

// BuildOutput builds OutputBox. it firstly tries built-in plugin, and then try 3rd party plugin
func BuildOutput(outputType string, config map[interface{}]interface{}) *topology.OutputBox {
	var output topology.Output
	var err error
	if v, ok := registeredOutput[outputType]; ok {
		output = v(config)
	} else {
		glog.Info("use third party plugin")
		output, err = getOutputFromPlugin(outputType, config)
		if err != nil {
			glog.Errorf("could not load %s: %v", outputType, err)
			return nil
		}
	}

	return &topology.OutputBox{
		Output:          output,
		ConditionFilter: condition_filter.NewConditionFilter(config),
	}
}

func getOutputFromPlugin(pluginPath string, config map[interface{}]interface{}) (topology.Output, error) {
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("could not open %s: %v", pluginPath, err)
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
	filter, ok := rst.(topology.Output)
	if !ok {
		return nil, fmt.Errorf("`New` func in %s dose not return Output Interface", pluginPath)
	}
	return filter, nil
}
