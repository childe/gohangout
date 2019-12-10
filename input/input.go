package input

import (
	"plugin"

	"github.com/golang/glog"
)

type Input interface {
	readOneEvent() map[string]interface{}
	Shutdown()
}

func GetInput(inputType string, config map[interface{}]interface{}) Input {
	switch inputType {
	case "Stdin":
		f := NewStdinInput(config)
		return f
	case "Kafka":
		f := NewKafkaInput(config)
		return f
	case "Random":
		f := NewRandomInput(config)
		return f
	case "TCP":
		f := NewTCPInput(config)
		return f
	default:
		p, err := plugin.Open(inputType)
		if err != nil {
			glog.Fatalf("could not open %s: %s", inputType, err)
		}
		newFunc, err := p.Lookup("New")
		if err != nil {
			glog.Fatalf("could not find New function in %s: %s", inputType, err)
		}
		return newFunc.(func(map[interface{}]interface{}) interface{})(config).(Input)
	}
}
