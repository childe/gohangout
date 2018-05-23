package input

import (
	"github.com/childe/gohangout/filter"
	"github.com/childe/gohangout/output"
	"github.com/golang-collections/collections/stack"
	"github.com/golang/glog"
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

func (box *InputBox) Beat() {
	var (
		event   map[string]interface{}
		events  []map[string]interface{}
		success bool

		sFrom *stack.Stack = stack.New()
		sTo   *stack.Stack = stack.New()
	)

	for {
		event = box.input.readOneEvent()
		if event == nil {
			continue
		}
		if typeValue, ok := box.config["type"]; ok {
			event["type"] = typeValue
		}
		sFrom.Push(event)

		if box.filters != nil {
			for _, filterPlugin := range box.filters {
				for sFrom.Len() > 0 {
					event = sFrom.Pop().(map[string]interface{})
					if filterPlugin.Pass(event) == false {
						sTo.Push(event)
						continue
					}
					event, success = filterPlugin.Process(event)
					if event != nil {
						event = filterPlugin.PostProcess(event, success)
						if event != nil {
							sTo.Push(event)
						}
					}
					events = filterPlugin.EmitExtraEvents(sTo)
					if events != nil {
						for _, event := range events {
							sTo.Push(event)
						}
					}
				}
				sFrom, sTo = sTo, sFrom
			}
		}

		for sFrom.Len() > 0 {
			event = sFrom.Pop().(map[string]interface{})
			for _, outputPlugin := range box.outputs {
				if outputPlugin.Pass(event) {
					outputPlugin.Emit(event)
				}
			}
		}
	}
}

func (box *InputBox) Shutdown() {
	glog.Infof("try to shutdown input %T", box.input)
	box.input.Shutdown()
	for _, o := range box.outputs {
		glog.Infof("try to shutdown output %T", o)
		o.Shutdown()
	}
}
