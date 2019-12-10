package input

import (
	"sync"

	"github.com/childe/gohangout/filter"
	"github.com/childe/gohangout/output"
	"github.com/childe/gohangout/topology"
	"github.com/golang/glog"
)

type InputBox struct {
	config             map[string]interface{} // whole config
	input              topology.Input
	outputsInAllWorker [][]*output.OutputBox
	stop               bool
	once               sync.Once
	shutdownChan       chan bool
}

func NewInputBox(input topology.Input, config map[string]interface{}) *InputBox {
	return &InputBox{
		input:        input,
		config:       config,
		stop:         false,
		shutdownChan: make(chan bool, 1),
	}
}

func (box *InputBox) beat(workerIdx int) {
	var outputProcessorInLink topology.ProcessorInLink
	outputs := output.BuildOutputs(box.config)
	if len(outputs) == 1 {
		outputProcessorInLink = (*output.OutputProcessorInLink)(outputs[0])
	} else {
		outputProcessorInLink = (output.OutputsProcessorInLink)(outputs)
	}
	filterBoxes := filter.BuildFilterBoxes(box.config, outputProcessorInLink)
	box.outputsInAllWorker[workerIdx] = outputs

	var firstNode topology.ProcessorInLink
	if len(filterBoxes) > 0 {
		firstNode = (*filter.FilterProcessorInLink)(filterBoxes[0])
	} else {
		firstNode = outputProcessorInLink
	}

	var (
		event map[string]interface{}
	)

	for !box.stop {
		event = box.input.ReadOneEvent()
		if event == nil {
			if !box.stop {
				glog.Info("receive nil message. shutdown...")
				box.shutdown()
			}
			return
		}
		firstNode.Process(event)
	}
}

func (box *InputBox) Beat(worker int) {
	box.outputsInAllWorker = make([][]*output.OutputBox, worker)
	for i := 0; i < worker; i++ {
		go box.beat(i)
	}

	<-box.shutdownChan
}

func (box *InputBox) shutdown() {
	box.once.Do(func() {

		glog.Infof("try to shutdown input %T", box.input)
		box.input.Shutdown()

		for i, outputs := range box.outputsInAllWorker {
			for _, o := range outputs {
				glog.Infof("try to shutdown output %T in worker %d", o, i)
				o.Output.Shutdown()
			}
		}
	})

	box.shutdownChan <- true
}

func (box *InputBox) Shutdown() {
	box.shutdown()
	box.stop = true
}
