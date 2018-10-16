package input

import "github.com/golang/glog"

type InputBox struct {
	input Input

	stopped bool
}

func NewInputBox(input Input) InputBox {
	return InputBox{
		//config: config,

		input: input,

		stopped: false,
	}
}

func (box *InputBox) Beat() {
	var (
		event map[string]interface{}
	)

	for box.stopped == false {
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

	box.shutdown()
}

func (box *InputBox) Shutdown() {
	box.stopped = true
}

func (box *InputBox) shutdown() {
	glog.Infof("try to shutdown input %T", box.input)
	//box.input.Shutdown()
	//for _, o := range box.outputs {
	//glog.Infof("try to shutdown output %T", o)
	//o.Shutdown()
	//}
}
