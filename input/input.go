package input

import (
	"plugin"
	"strings"

	"github.com/childe/gohangout/topology"
	"github.com/golang/glog"
)

type BuildInputFunc func(map[interface{}]interface{}) topology.Input

var registeredInput map[string]BuildInputFunc = make(map[string]BuildInputFunc)

// Register is used by input plugins to register themselves
func Register(inputType string, bf BuildInputFunc) {
	if _, ok := registeredInput[inputType]; ok {
		glog.Errorf("%s has been registered, ignore %T", inputType, bf)
		return
	}
	registeredInput[inputType] = bf
}

// GetInput return topoloty.Input from builtin plugins or from a 3rd party plugin
func GetInput(inputType string, config map[interface{}]interface{}) topology.Input {
	if v, ok := registeredInput[inputType]; ok {
		return v(config)
	}
	glog.Info("could not load %s input plugin, try third party plugin", inputType)

	pluginPath := inputType
	if !strings.HasSuffix(pluginPath, ".so") {
		pluginPath = inputType + ".so"
	}
	_, err := plugin.Open(pluginPath)
	if err != nil {
		glog.Errorf("could not open %s: %s", pluginPath, err)
		return nil
	}

	if v, ok := registeredInput[inputType]; ok {
		return v(config)
	}
	glog.Errorf("could not load %s input plugin", inputType)
	return nil
}
