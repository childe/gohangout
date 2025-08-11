package input

import (
	"fmt"
	"plugin"

	"github.com/childe/gohangout/topology"
	"k8s.io/klog/v2"
)

type BuildInputFunc func(map[any]any) topology.Input

var registeredInput map[string]BuildInputFunc = make(map[string]BuildInputFunc)

// Register is used by input plugins to register themselves
func Register(inputType string, bf BuildInputFunc) {
	if _, ok := registeredInput[inputType]; ok {
		klog.Errorf("%s has been registered, ignore %T", inputType, bf)
		return
	}
	registeredInput[inputType] = bf
}

// GetInput return topoloty.Input from builtin plugins or from a 3rd party plugin
func GetInput(inputType string, config map[any]any) topology.Input {
	if v, ok := registeredInput[inputType]; ok {
		return v(config)
	}
	klog.Infof("could not load %s input plugin, try third party plugin", inputType)

	pluginPath := inputType
	output, err := getInputFromPlugin(pluginPath, config)
	if err != nil {
		klog.Errorf("could not load %s: %v", pluginPath, err)
		return nil
	}
	return output
}

func getInputFromPlugin(pluginPath string, config map[any]any) (topology.Input, error) {
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("could not open %s: %v", pluginPath, err)
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
	input, ok := rst.(topology.Input)
	if !ok {
		return nil, fmt.Errorf("`New` func in %s dose not return Input Interface", pluginPath)
	}
	return input, nil
}
