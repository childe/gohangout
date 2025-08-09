package topology

import (
	"github.com/childe/gohangout/condition_filter"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog/v2"
)

type Output interface {
	Emit(map[string]any)
	Shutdown()
}

type OutputBox struct {
	Output
	*condition_filter.ConditionFilter
	promCounter prometheus.Counter
}

type buildOutputFunc func(outputType string, config map[any]any) *OutputBox

func BuildOutputs(config map[string]any, buildOutput buildOutputFunc) []*OutputBox {
	rst := make([]*OutputBox, 0)

	for _, outputs := range config["outputs"].([]any) {
		for outputType, outputConfig := range outputs.(map[any]any) {
			outputType := outputType.(string)
			klog.Infof("output type: %s", outputType)
			outputConfig := outputConfig.(map[any]any)
			output := buildOutput(outputType, outputConfig)

			output.promCounter = GetPromCounter(outputConfig)

			rst = append(rst, output)
		}
	}
	return rst
}

// Process implement Processor interface
func (p *OutputBox) Process(event map[string]any) map[string]any {
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
func (p OutputsProcessor) Process(event map[string]any) map[string]any {
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
