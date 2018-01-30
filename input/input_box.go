package input

import (
	"github.com/childe/gohangout/filter"
	"github.com/childe/gohangout/output"
)

type InputBox struct {
	input   Input
	filters []filter.Filter
	outputs []output.Output
	config  map[interface{}]interface{}
}

func NewInputBox(input Input, filters []filter.Filter, outputs []output.Output, config map[interface{}]interface{}) InputBox {
	return InputBox{
		input:   input,
		filters: filters,
		outputs: outputs,
		config:  config,
	}
}

func (box *InputBox) prepare(event map[string]interface{}) map[string]interface{} {
	return event
}

func (box *InputBox) Beat() {
	//box.input.init(box.config)
	var (
		event   map[string]interface{}
		success bool
	)

	for {
		event = box.input.readOneEvent()
		if event == nil {
			continue
		}
		if typeValue, ok := box.config["type"]; ok {
			event["type"] = typeValue
		}
		if box.filters != nil {
			for _, filterPlugin := range box.filters {
				if filterPlugin.Pass(event) == false {
					continue
				}
				event, success = filterPlugin.Process(event)
				if event == nil {
					break
				}
				filterPlugin.PostProcess(event, success)
			}
		}
		if event == nil {
			continue
		}
		for _, outputPlugin := range box.outputs {
			outputPlugin.Emit(event)
		}
	}
}
