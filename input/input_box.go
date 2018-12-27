package input

import (
	"github.com/childe/gohangout/output"
	"github.com/golang/glog"
)

type InputBox struct {
	input   Input
	outputs []output.Output
	stop    bool
}

func NewInputBox(input Input, outputs []output.Output) *InputBox {
	return &InputBox{
		input:   input,
		outputs: outputs,
		stop:    false,
	}
}

func (box *InputBox) Beat() {
	var (
		event map[string]interface{}
	)

	for !box.stop {
		event = box.input.readOneEvent()
		if event == nil {
			glog.Info("receive nil message. shutdown...")
			break
		}
		box.input.GotoNext(event)
	}

	box.shutdown()
}

func (box *InputBox) Shutdown() {
	box.stop = true
}

func (box *InputBox) shutdown() {
	glog.Infof("try to shutdown input %T", box.input)
	box.input.Shutdown()
	for _, o := range box.outputs {
		glog.Infof("try to shutdown output %T", o)
		o.Shutdown()
	}
}
