package output

import (
	"plugin"
	"strings"

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

func BuildOutput(outputType string, config map[interface{}]interface{}) *topology.OutputBox {
	var output topology.Output
	if strings.HasSuffix(outputType, ".so") {
		_, err := plugin.Open(outputType)
		if err != nil {
			glog.Errorf("could not open %s: %s", outputType, err)
		}
	}

	if v, ok := registeredOutput[outputType]; ok {
		output = v(config)
	} else {
		glog.Errorf("could not load %s output plugin", outputType)
		return nil
	}

	return &topology.OutputBox{
		output,
		condition_filter.NewConditionFilter(config),
	}
}
