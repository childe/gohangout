package output

import (
	"fmt"

	"github.com/childe/gohangout/codec"
	"github.com/golang/glog"
)

type StdoutOutput struct {
	BaseOutput
	config  map[interface{}]interface{}
	encodec codec.Encoder
}

func NewStdoutOutput(config map[interface{}]interface{}) *StdoutOutput {
	p := &StdoutOutput{
		BaseOutput: NewBaseOutput(config),
		config:     config,
	}

	if v, ok := config["codec"]; ok {
		p.encodec = codec.NewEncoder(v.(string))
	} else {
		p.encodec = codec.NewEncoder("json")
	}

	return p

}

func (p *StdoutOutput) Emit(event map[string]interface{}) {
	buf, err := p.encodec.Encode(event)
	if err != nil {
		glog.Errorf("marshal %v error:%s", event, err)
	}
	fmt.Println(string(buf))
}

func (p *StdoutOutput) Shutdown() {}
