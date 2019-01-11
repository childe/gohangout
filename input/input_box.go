package input

import (
	"sync"

	"github.com/childe/gohangout/output"
	"github.com/golang/glog"
)

type InputBox struct {
	input       Input
	outputs     []output.Output
	stop        bool
	shutdownWG  sync.WaitGroup
	workerWG    sync.WaitGroup
	shutdownMux sync.Mutex
}

func NewInputBox(input Input, outputs []output.Output) *InputBox {
	return &InputBox{
		input:   input,
		outputs: outputs,
		stop:    false,
	}
}

func (box *InputBox) beat() {
	defer box.workerWG.Done()

	var (
		event map[string]interface{}
	)

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

func (box *InputBox) Beat(worker int) {
	defer box.workerWG.Wait()
	defer box.shutdownWG.Wait() // wait shutdown

	for i := 0; i < worker; i++ {
		box.workerWG.Add(1)
		go box.beat()
	}
}

func (box *InputBox) shutdown() {
	box.shutdownMux.Lock()
	defer box.shutdownMux.Unlock()

	if box.stop {
		return
	}

	box.stop = true
	glog.Infof("try to shutdown input %T", box.input)
	box.input.Shutdown()
	for _, o := range box.outputs {
		glog.Infof("try to shutdown output %T", o)
		o.Shutdown()
	}
}

func (box *InputBox) Shutdown() {
	box.shutdownWG.Add(1)
	defer box.shutdownWG.Done()
	box.shutdown()
}
