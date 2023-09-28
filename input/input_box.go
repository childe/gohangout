package input

import (
	"math/rand"
	"reflect"
	"sync"

	"github.com/childe/gohangout/field_setter"
	"github.com/childe/gohangout/filter"
	"github.com/childe/gohangout/output"
	"github.com/childe/gohangout/topology"
	"github.com/childe/gohangout/value_render"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog/v2"
)

type InputBox struct {
	config             map[string]interface{} // whole config
	input              topology.Input
	outputsInAllWorker [][]*topology.OutputBox
	stop               bool
	once               sync.Once
	shutdownChan       chan bool

	workerDispatchField    string
	eventQueueForAllWorker []chan map[string]interface{}

	promCounter prometheus.Counter

	shutdownWhenNil bool
	exit            func()

	addFields map[field_setter.FieldSetter]value_render.ValueRender
}

// SetShutdownWhenNil is used for benchmark.
// Gohangout main thread would exit when one input box receive a nil message, such as Ctrl-D in Stdin input
func (box *InputBox) SetShutdownWhenNil(shutdownWhenNil bool) {
	box.shutdownWhenNil = shutdownWhenNil
}

func NewInputBox(input topology.Input, inputConfig map[interface{}]interface{}, config map[string]interface{}, exit func()) *InputBox {
	b := &InputBox{
		input:        input,
		config:       config,
		stop:         false,
		shutdownChan: make(chan bool, 1),

		promCounter: topology.GetPromCounter(inputConfig),

		exit: exit,
	}
	if add_fields, ok := inputConfig["add_fields"]; ok {
		b.addFields = make(map[field_setter.FieldSetter]value_render.ValueRender)
		for k, v := range add_fields.(map[interface{}]interface{}) {
			fieldSetter := field_setter.NewFieldSetter(k.(string))
			if fieldSetter == nil {
				klog.Errorf("could build field setter from %s", k.(string))
				return nil
			}
			b.addFields[fieldSetter] = value_render.GetValueRender(v.(string))
		}
	} else {
		b.addFields = nil
	}
	if dispatch_field, ok := inputConfig["worker_dispatch_field"].(string); ok {
		b.workerDispatchField = dispatch_field
	} else {
		b.workerDispatchField = ""
	}

	return b
}

func (box *InputBox) beat(workerIdx int) {
	var firstNode *topology.ProcessorNode = box.buildTopology(workerIdx)

	var (
		event map[string]interface{}
	)

	for !box.stop {
		event = box.input.ReadOneEvent()
		if box.promCounter != nil {
			box.promCounter.Inc()
		}
		if event == nil {
			klog.V(5).Info("received nil message.")
			if box.stop {
				break
			}
			if box.shutdownWhenNil {
				klog.Info("received nil message. shutdown...")
				box.exit()
				break
			} else {
				continue
			}
		}
		for fs, v := range box.addFields {
			event = fs.SetField(event, v.Render(event), "", false)
		}
		firstNode.Process(event)
	}
}

func (box *InputBox) dispatchBeat() {
	var (
		event map[string]interface{}
	)

	for !box.stop {
		event = box.input.ReadOneEvent()
		if box.promCounter != nil {
			box.promCounter.Inc()
		}
		if event == nil {
			klog.V(5).Info("received nil message.")
			if box.stop {
				break
			}
			if box.shutdownWhenNil {
				klog.Info("received nil message. shutdown...")
				box.exit()
				break
			} else {
				continue
			}
		}

		var workerIdx int
		queueSize := len(box.eventQueueForAllWorker)
		v := value_render.GetValueRender(box.workerDispatchField)
		if dispatchKey, ok := v.Render(event).(int32); ok {
			workerIdx = int(dispatchKey) % queueSize
		} else {
			workerIdx = rand.Intn(queueSize)
			klog.Info("Invalid workerDispatchField, use random worker index ", workerIdx)
		}
		box.dispatchToWorker(event, workerIdx)
	}
}

func (box *InputBox) dispatchToWorker(event map[string]interface{}, workerIdx int) {
	for fs, v := range box.addFields {
		event = fs.SetField(event, v.Render(event), "", false)
	}

	box.eventQueueForAllWorker[workerIdx] <- event //TODO： worker超时
}

func (box *InputBox) workerBeat(workerIdx int) {
	var firstNode *topology.ProcessorNode = box.buildTopology(workerIdx)

	var (
		event map[string]interface{}
	)

	for !box.stop {
		event = <-box.eventQueueForAllWorker[workerIdx]
		klog.V(10).Infof("worker %d start processing one message %v", workerIdx, event)

		// prome.IncCounterWithLabel("worker_get_event_count", map[string]string{
		// 	"workerIdx": strconv.Itoa(workerIdx),
		// })

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

	var firstNode *topology.ProcessorNode
	for _, b := range filterBoxes {
		firstNode = topology.AppendProcessorsToLink(firstNode, b)
	}
	firstNode = topology.AppendProcessorsToLink(firstNode, outputProcessor)

	// Set BelongTo
	var node *topology.ProcessorNode
	node = firstNode
	for _, b := range filterBoxes {
		node = node.Next
		v := reflect.ValueOf(b.Filter)
		f := v.MethodByName("SetBelongTo")
		if f.IsValid() {
			f.Call([]reflect.Value{reflect.ValueOf(node)})
		}
	}

	return firstNode
}

// Beat starts the processors and wait until shutdown
// if workerDispatch enabled, events will be dispatch to all workers by the dispatch key
func (box *InputBox) Beat(worker int) {
	box.outputsInAllWorker = make([][]*topology.OutputBox, worker)

	if worker > 1 && len(box.workerDispatchField) != 0 {
		klog.Infof("Initializing %d workers", worker)
		box.eventQueueForAllWorker = make([]chan map[string]interface{}, worker)
		for i := 0; i < worker; i++ {
			box.eventQueueForAllWorker[i] = make(chan map[string]interface{}, 10)
		}
		go box.dispatchBeat()

		for i := 0; i < worker; i++ {
			go box.workerBeat(i)
		}
	} else {
		for i := 0; i < worker; i++ {
			go box.beat(i)
		}
	}

	<-box.shutdownChan
}

func (box *InputBox) shutdown() {
	box.once.Do(func() {
		// input must not shutdown before output
		// for the commit work in output may need input
		box.input.Pause()

		for i, outputs := range box.outputsInAllWorker {
			for _, o := range outputs {
				klog.Infof("try to shutdown output %T in worker %d", o, i)
				o.Output.Shutdown()
			}
		}

		klog.Infof("try to shutdown input %T", box.input)
		box.input.Shutdown()
	})

	box.shutdownChan <- true
}

// Shutdown shutdowns the inputs and outputs
func (box *InputBox) Shutdown() {
	box.shutdown()
	box.stop = true
}
