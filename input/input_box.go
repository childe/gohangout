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

	simple bool
}

func NewInputBox(input Input, filters []filter.Filter, outputs []output.Output, config map[interface{}]interface{}) InputBox {
	box := InputBox{
		input:   input,
		filters: filters,
		outputs: outputs,
		config:  config,
		simple:  true,
	}
	for _, f := range filters {
		if f.IfSimple() == false {
			box.simple = false
			break
		}
	}
	if box.simple {
		glog.Info("box is simple")
	} else {
		glog.Info("box is not simple")
	}

	return box
}

func (box *InputBox) beatSimple() {
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
				} else {
					event = filterPlugin.PostProcess(event, success)
				}
			}
		}
		if event == nil {
			continue
		}

		for _, outputPlugin := range box.outputs {
			if outputPlugin.Pass(event) {
				outputPlugin.Emit(event)
			}
		}
	}
}
func (box *InputBox) beatNotSimple() {
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

func (box *InputBox) Beat() {
	if box.simple {
		box.beatSimple()
	} else {
		box.beatNotSimple()
	}
}
