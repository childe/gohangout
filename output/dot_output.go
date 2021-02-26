package output

import (
	"fmt"

	"github.com/childe/gohangout/topology"
)

type DotOutput struct {
	config map[interface{}]interface{}
}

func newDotOutput(config map[interface{}]interface{}) topology.Output {
	return &DotOutput{
		config: config,
	}
}

func init() {
	Register("Dot", newDotOutput)
}

func (outputPlugin *DotOutput) Emit(event map[string]interface{}) {
	fmt.Print(".")
}

func (outputPlugin *DotOutput) Shutdown() {}
