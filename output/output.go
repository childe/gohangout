package output

import (
	"plugin"

	"github.com/childe/gohangout/condition_filter"
	"github.com/childe/gohangout/topology"
	"github.com/golang/glog"
)

type OutputBox struct {
	topology.Output
	*condition_filter.ConditionFilter
}

func BuildOutputs(config map[string]interface{}) []*OutputBox {
	rst := make([]*OutputBox, 0)

	for _, outputI := range config["outputs"].([]interface{}) {
		// len(outputI) is 1
		for outputTypeI, outputConfigI := range outputI.(map[interface{}]interface{}) {
			outputType := outputTypeI.(string)
			glog.Infof("output type: %s", outputType)
			outputConfig := outputConfigI.(map[interface{}]interface{})
			glog.Infof("output config: %v", outputConfig)
			outputPlugin := BuildOutput(outputType, outputConfig)
			rst = append(rst, outputPlugin)
		}
	}
	return rst
}

func BuildOutput(outputType string, config map[interface{}]interface{}) *OutputBox {
	var output topology.Output
	switch outputType {
	case "Dot":
		output = NewDotOutput(config)
	case "Stdout":
		output = NewStdoutOutput(config)
	case "Kafka":
		output = NewKafkaOutput(config)
	case "Elasticsearch":
		output = NewElasticsearchOutput(config)
	case "Influxdb":
		output = NewInfluxdbOutput(config)
	case "Clickhouse":
		output = NewClickhouseOutput(config)
	case "TCP":
		output = NewTCPOutput(config)
	default:
		p, err := plugin.Open(outputType)
		if err != nil {
			glog.Fatalf("could not open %s: %s", outputType, err)
		}
		newFunc, err := p.Lookup("New")
		if err != nil {
			glog.Fatalf("could not find New function in %s: %s", outputType, err)
		}
		output = newFunc.(func(map[interface{}]interface{}) interface{})(config).(topology.Output)
	}

	return &OutputBox{
		output,
		condition_filter.NewConditionFilter(config),
	}
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
