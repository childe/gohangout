package input

import (
	"reflect"
	"sync"

	"github.com/childe/gohangout/filter"
	"github.com/childe/gohangout/output"
	"github.com/childe/gohangout/topology"
	"github.com/golang/glog"
)

type InputBox struct {
	config             map[string]interface{} // whole config
	input              topology.Input
	outputsInAllWorker [][]*topology.OutputBox
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
	var firstNode *topology.ProcessorNode = box.buildTopology(workerIdx)

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

func (box *InputBox) buildTopology(workerIdx int) *topology.ProcessorNode {
	outputs := topology.BuildOutputs(box.config, output.BuildOutput)
	box.outputsInAllWorker[workerIdx] = outputs

	var outputProcessor topology.Processor
	if len(outputs) == 1 {
		outputProcessor = outputs[0]
	} else {
		outputProcessor = (topology.OutputsProcessor)(outputs)
	}

	filterBoxes := topology.BuildFilterBoxes(box.config, filter.BuildFilter)

	// Set BelongTo
	for i, b := range filterBoxes {
		v := reflect.ValueOf(b.Filter)
		f := v.MethodByName("SetBelongTo")
		if f.IsValid() {
			if i == len(filterBoxes)-1 {
				f.Call([]reflect.Value{reflect.ValueOf(outputProcessor)})
			} else {
				f.Call([]reflect.Value{reflect.ValueOf(filterBoxes[i+1])})
			}
		}
	}

	var firstNode *topology.ProcessorNode
	for _, b := range filterBoxes {
		firstNode = topology.AppendProcessorsToLink(firstNode, b)
	}
	firstNode = topology.AppendProcessorsToLink(firstNode, outputProcessor)
	return firstNode
}

func (box *InputBox) Beat(worker int) {
	box.outputsInAllWorker = make([][]*topology.OutputBox, worker)
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
