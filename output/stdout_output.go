package output

import (
	"fmt"

	"github.com/childe/gohangout/codec"
	"github.com/childe/gohangout/topology"
	"k8s.io/klog/v2"
)

// StdoutConfig defines the configuration structure for Stdout output
type StdoutConfig struct {
	Codec string `json:"codec"`
}

func init() {
	Register("Stdout", newStdoutOutput)
}

type StdoutOutput struct {
	config  map[any]any
	encoder codec.Encoder
}

func newStdoutOutput(config map[any]any) topology.Output {
	// Parse configuration using SafeDecodeConfig helper
	var stdoutConfig StdoutConfig
	stdoutConfig.Codec = "json" // set default value

	SafeDecodeConfig("Stdout", config, &stdoutConfig)

	p := &StdoutOutput{
		config:  config,
		encoder: codec.NewEncoder(stdoutConfig.Codec),
	}

	return p
}

func (p *StdoutOutput) Emit(event map[string]any) {
	buf, err := p.encoder.Encode(event)
	if err != nil {
		klog.Errorf("marshal %v error:%s", event, err)
	}
	fmt.Println(string(buf))
}

func (p *StdoutOutput) Shutdown() {}
