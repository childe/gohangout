package output

import "fmt"

type DotOutput struct {
	config map[interface{}]interface{}
}

func (l *MethodLibrary) NewDotOutput(config map[interface{}]interface{}) *DotOutput {
	return &DotOutput{
		config: config,
	}
}

func (outputPlugin *DotOutput) Emit(event map[string]interface{}) {
	fmt.Print(".")
}

func (outputPlugin *DotOutput) Shutdown() {}
