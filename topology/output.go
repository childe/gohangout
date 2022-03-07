package topology

import (
	"github.com/childe/gohangout/condition_filter"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
)

type Output interface {
	Emit(map[string]interface{})
	Shutdown()
}

type OutputBox struct {
	Output
	*condition_filter.ConditionFilter
	promCounter prometheus.Counter
}

type buildOutputFunc func(outputType string, config map[interface{}]interface{}) *OutputBox

func BuildOutputs(config map[string]interface{}, buildOutput buildOutputFunc) []*OutputBox {
	rst := make([]*OutputBox, 0)

	for _, outputs := range config["outputs"].([]interface{}) {
		for outputType, outputConfig := range outputs.(map[interface{}]interface{}) {
			outputType := outputType.(string)
			glog.Infof("output type: %s", outputType)
			outputConfig := outputConfig.(map[interface{}]interface{})
			output := buildOutput(outputType, outputConfig)

			output.promCounter = GetPromCounter(outputConfig)

			rst = append(rst, output)
		}
	}
	return rst
}

// Process implement Processor interface
func (p *OutputBox) Process(event map[string]interface{}) map[string]interface{} {
	if p.Pass(event) {
		if p.promCounter != nil {
			p.promCounter.Inc()
		}
		p.Emit(event)
	}
	return nil
}

type OutputsProcessor []*OutputBox

// Process implement Processor interface
func (p OutputsProcessor) Process(event map[string]interface{}) map[string]interface{} {
	for _, o := range ([]*OutputBox)(p) {
		if o.Pass(event) {
			if o.promCounter != nil {
				o.promCounter.Inc()
			}
			o.Emit(event)
		}
	}
	return nil
}
