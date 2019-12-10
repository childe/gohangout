package output

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
	var output Output
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
		glog.Fatalf("could not build %s output plugin", outputType)
		output = nil
	}

	return &OutputBox{
		Output:          output,
		ConditionFilter: condition_filter.NewConditionFilter(config),
	}
}

type BaseOutput struct {
	conditionFilter *condition_filter.ConditionFilter
}

func NewBaseOutput(config map[interface{}]interface{}) BaseOutput {
	return BaseOutput{
		conditionFilter: condition_filter.NewConditionFilter(config),
	}
}

func (f BaseOutput) Pass(event map[string]interface{}) bool {
	return f.conditionFilter.Pass(event)
}
