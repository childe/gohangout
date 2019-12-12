package output

import (
	"fmt"

	"github.com/childe/gohangout/codec"
	"github.com/golang/glog"
)

type StdoutOutput struct {
	config  map[interface{}]interface{}
	encoder codec.Encoder
}

func (l *MethodLibrary) NewStdoutOutput(config map[interface{}]interface{}) *StdoutOutput {
	p := &StdoutOutput{
		config: config,
	}

	if v, ok := config["codec"]; ok {
		p.encoder = codec.NewEncoder(v.(string))
	} else {
		p.encoder = codec.NewEncoder("json")
	}

	return p

}

func (p *StdoutOutput) Emit(event map[string]interface{}) {
	buf, err := p.encoder.Encode(event)
	if err != nil {
		glog.Errorf("marshal %v error:%s", event, err)
	}
	fmt.Println(string(buf))
}

func (p *StdoutOutput) Shutdown() {}
