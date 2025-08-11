package output

import (
	"fmt"

	"github.com/childe/gohangout/topology"
)

type DotOutput struct {
	config map[any]any
}

func newDotOutput(config map[any]any) topology.Output {
	return &DotOutput{
		config: config,
	}
}

func init() {
	Register("Dot", newDotOutput)
}

func (outputPlugin *DotOutput) Emit(event map[string]any) {
	fmt.Print(".")
}

func (outputPlugin *DotOutput) Shutdown() {}
