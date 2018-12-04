package output

import (
	"fmt"

	"github.com/childe/gohangout/codec"
	"github.com/golang/glog"
)

type StdoutOutput struct {
	BaseOutput
	config map[interface{}]interface{}
}

func NewStdoutOutput(config map[interface{}]interface{}) *StdoutOutput {
	return &StdoutOutput{
		BaseOutput: NewBaseOutput(config),
		config:     config,
	}
}

func (outputPlugin *StdoutOutput) Emit(event map[string]interface{}) {
	d := &codec.SimpleJsonDecoder{}
	buf, err := d.Encode(event)
	if err != nil {
		glog.Errorf("marshal %v error:%s", event, err)
	}
	fmt.Println(string(buf))
}

func (outputPlugin *StdoutOutput) Shutdown() {}
