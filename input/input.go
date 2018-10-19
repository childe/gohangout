package input

import (
	"github.com/childe/gohangout/filter"
	"github.com/childe/gohangout/output"
)

type Input interface {
	readOneEvent() map[string]interface{}
	Shutdown()

	GotoNext(map[string]interface{})
}

func GetInput(inputType string, config map[interface{}]interface{}, nextFilter filter.Filter, outputs []output.Output) Input {
	switch inputType {
	case "Stdin":
		f := NewStdinInput(config)
		f.BaseInput.nextFilter = nextFilter
		f.BaseInput.outputs = outputs
		return f
	case "Kafka":
		f := NewKafkaInput(config)
		f.BaseInput.nextFilter = nextFilter
		f.BaseInput.outputs = outputs
		return f
	}
	return nil
}

type BaseInput struct {
	nextFilter filter.Filter
	outputs    []output.Output
}

func (i *BaseInput) GotoNext(event map[string]interface{}) {
	if i.nextFilter != nil {
		i.nextFilter.Process(event)
	} else {
		for _, outputPlugin := range i.outputs {
			if outputPlugin.Pass(event) {
				outputPlugin.Emit(event)
			}
		}
	}
}
