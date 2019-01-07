package input

import (
	"sync"

	"github.com/childe/gohangout/output"
	"github.com/golang/glog"
)

type InputBox struct {
	input   Input
	outputs []output.Output
	stop    bool
	wg      sync.WaitGroup
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

	defer box.wg.Wait()
	for !box.stop {
		event = box.input.readOneEvent()
		if event == nil {
			glog.Info("receive nil message. shutdown...")
			box.shutdown()
			return
		}
		box.input.GotoNext(event)
	}
}

func (box *InputBox) Shutdown() {
	box.wg.Add(1)
	defer box.wg.Done()
	box.stop = true
	box.shutdown()
}

func (box *InputBox) shutdown() {
	glog.Infof("try to shutdown input %T", box.input)
	box.input.Shutdown()
	for _, o := range box.outputs {
		glog.Infof("try to shutdown output %T", o)
		o.Shutdown()
	}
}
