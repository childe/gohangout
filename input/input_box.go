package input

import (
	"github.com/childe/gohangout/output"
	"github.com/golang/glog"
)

type InputBox struct {
	input   Input
	outputs []output.Output
}

func NewInputBox(input Input, outputs []output.Output) *InputBox {
	return &InputBox{
		input:   input,
		outputs: outputs,
	}
}

func (box *InputBox) Beat() {
	var (
		event map[string]interface{}
	)

	for {
		event = box.input.readOneEvent()
		if event == nil {
			glog.Info("receive nil message. shutdown...")
			break
		}
		//if typeValue, ok := box.config["type"]; ok {
		//event["type"] = typeValue
		//}

		box.input.GotoNext(event)
	}

	box.Shutdown()
}

func (box *InputBox) Shutdown() {
	glog.Infof("try to shutdown input %T", box.input)
	box.input.Shutdown()
	for _, o := range box.outputs {
		glog.Infof("try to shutdown output %T", o)
		o.Shutdown()
	}
}
