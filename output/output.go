package output

import (
	"github.com/childe/gohangout/condition_filter"
	"github.com/golang/glog"
)

type Output interface {
	Emit(map[string]interface{})
	Pass(map[string]interface{}) bool
	Shutdown()
}

func BuildOutputs(config map[string]interface{}) []Output {
	rst := make([]Output, 0)

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

func BuildOutput(outputType string, config map[interface{}]interface{}) Output {
	switch outputType {
	case "Stdout":
		return NewStdoutOutput(config)
	case "Kafka":
		return NewKafkaOutput(config)
	case "Elasticsearch":
		return NewElasticsearchOutput(config)
	case "Influxdb":
		return NewInfluxdbOutput(config)
	case "Clickhouse":
		return NewClickhouseOutput(config)
	}
	glog.Fatalf("could not build %s output plugin", outputType)
	return nil
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
