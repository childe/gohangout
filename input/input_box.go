package input

import (
	"sync"

	"github.com/childe/gohangout/filter"
	"github.com/childe/gohangout/output"
	"github.com/golang/glog"
)

type InputBox struct {
	config             map[string]interface{} // whole config
	input              Input
	outputsInAllWorker [][]output.Output
	stop               bool
	shutdownWG         sync.WaitGroup
	workerWG           sync.WaitGroup
	shutdownMux        sync.Mutex
}

func NewInputBox(input Input, config map[string]interface{}) *InputBox {
	return &InputBox{
		input:  input,
		config: config,
		stop:   false,
	}
}

func (box *InputBox) beat(workerIdx int) {
	defer box.workerWG.Done()

	outputs := output.BuildOutputs(box.config)
	filterBoxes := filter.BuildFilterBoxes(box.config, outputs)
	box.outputsInAllWorker[workerIdx] = outputs

	var nexter filter.Nexter
	if len(filterBoxes) > 0 {
		nexter = &filter.FilterNexter{filterBoxes[0]}
	} else {
		if len(outputs) == 1 {
			nexter = &filter.OutputNexter{outputs[0]}
		} else {
			nexter = &filter.OutputsNexter{outputs}
		}
	}

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
		nexter.Process(event)
	}
}

func (box *InputBox) Beat(worker int) {
	defer box.workerWG.Wait()
	defer box.shutdownWG.Wait() // wait shutdown

	box.outputsInAllWorker = make([][]output.Output, worker)
	for i := 0; i < worker; i++ {
		box.workerWG.Add(1)
		go box.beat(i)
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

	for i, outputs := range box.outputsInAllWorker {
		for _, o := range outputs {
			glog.Infof("try to shutdown output %T in worker %d", o, i)
			o.Shutdown()
		}
	}
}

func (box *InputBox) Shutdown() {
	box.shutdownWG.Add(1)
	defer box.shutdownWG.Done()
	box.shutdown()
}
