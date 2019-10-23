package input

import (
	"context"
	"sync"

	"github.com/childe/gohangout/filter"
	"github.com/childe/gohangout/output"
	"github.com/golang/glog"
)

type InputBox struct {
	config             map[string]interface{} // whole config
	input              Input
	outputsInAllWorker [][]output.Output
	once               sync.Once
	wg                 sync.WaitGroup
	shutdownChan       chan bool
}

func NewInputBox(input Input, config map[string]interface{}) *InputBox {
	return &InputBox{
		input:        input,
		config:       config,
		shutdownChan: make(chan bool, 1),
	}
}

func (box *InputBox) beat(ctx context.Context, workerIdx int) {
	var outputNexter filter.Nexter
	outputs := output.BuildOutputs(box.config)
	if len(outputs) == 1 {
		outputNexter = &filter.OutputNexter{outputs[0]}
	} else {
		outputNexter = &filter.OutputsNexter{outputs}
	}
	filterBoxes := filter.BuildFilterBoxes(box.config, outputNexter)
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

	for {
		select {
		case <-ctx.Done():
			return
		default:
			event = box.input.readOneEvent()
			if event == nil {
				glog.Info("receive nil message. shutdown...")
				box.shutdown()
				return
			}
			nexter.Process(event)
		}
	}
}

func (box *InputBox) Beat(worker int) {
	ctx, cancel := context.WithCancel(context.Background())
	box.outputsInAllWorker = make([][]output.Output, worker)
	box.wg.Add(worker)
	for i := 0; i < worker; i++ {
		go func(i int) {
			defer box.wg.Done()
			box.beat(ctx, i)
		}(i)
	}

	<-box.shutdownChan
	cancel()
}

func (box *InputBox) shutdown() {
	box.once.Do(func() {
		glog.Infof("try to shutdown input %T", box.input)
		box.input.Shutdown()
		// close beat
		close(box.shutdownChan)
		// wait beat close
		box.wg.Wait()

		for i, outputs := range box.outputsInAllWorker {
			for _, o := range outputs {
				glog.Infof("try to shutdown output %T in worker %d", o, i)
				o.Shutdown()
			}
		}
	})

}

func (box *InputBox) Shutdown() {
	box.shutdown()
}
