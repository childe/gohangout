package output

import "fmt"

type DotOutput struct {
	BaseOutput
	config map[interface{}]interface{}
}

func NewDotOutput(config map[interface{}]interface{}) *DotOutput {
	return &DotOutput{
		BaseOutput: NewBaseOutput(config),
		config:     config,
	}
}

func (outputPlugin *DotOutput) Emit(event map[string]interface{}) {
	fmt.Print(".")
}

func (outputPlugin *DotOutput) Shutdown() {}
