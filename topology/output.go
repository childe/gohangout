package topology

import (
	"github.com/childe/gohangout/condition_filter"
	"github.com/golang/glog"
)

type Output interface {
	Emit(map[string]interface{})
	Shutdown()
}

type OutputBox struct {
	Output
	*condition_filter.ConditionFilter
}

type buildOutputFunc func(outputType string, config map[interface{}]interface{}) *OutputBox

func BuildOutputs(config map[string]interface{}, buildOutput buildOutputFunc) []*OutputBox {
	rst := make([]*OutputBox, 0)

	for _, outputI := range config["outputs"].([]interface{}) {
		// len(outputI) is 1
		for outputTypeI, outputConfigI := range outputI.(map[interface{}]interface{}) {
			outputType := outputTypeI.(string)
			glog.Infof("output type: %s", outputType)
			outputConfig := outputConfigI.(map[interface{}]interface{})
			glog.Infof("output config: %v", outputConfig)
			outputPlugin := buildOutput(outputType, outputConfig)
			rst = append(rst, outputPlugin)
		}
	}
	return rst
}

// Process implement Processor interface
func (p *OutputBox) Process(event map[string]interface{}) map[string]interface{} {
	if p.Pass(event) {
		p.Emit(event)
	}
	return nil
}

type OutputsProcessor []*OutputBox

// Process implement Processor interface
func (p OutputsProcessor) Process(event map[string]interface{}) map[string]interface{} {
	for _, o := range ([]*OutputBox)(p) {
		if o.Pass(event) {
			o.Emit(event)
		}
	}
	return nil
}
