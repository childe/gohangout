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

	stopped bool
}

func NewInputBox(input Input, filters []filter.Filter, outputs []output.Output, config map[interface{}]interface{}) InputBox {
	return InputBox{
		input:   input,
		filters: filters,
		outputs: outputs,
		config:  config,

		stopped: false,
	}
}

func (box *InputBox) Beat() {
	var (
		event   map[string]interface{}
		success bool

		sFrom *stack.Stack = stack.New()
		sTo   *stack.Stack = stack.New()
	)

	for box.stopped == false {
		event = box.input.readOneEvent()
		if event == nil {
			glog.Info("receive nil message. shutdown")
			break
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
					filterPlugin.EmitExtraEvents(sTo)
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

	box.shutdown()
}

func (box *InputBox) Shutdown() {
	box.stopped = true
}

func (box *InputBox) shutdown() {
	glog.Infof("try to shutdown input %T", box.input)
	box.input.Shutdown()
	for _, o := range box.outputs {
		glog.Infof("try to shutdown output %T", o)
		o.Shutdown()
	}
}
