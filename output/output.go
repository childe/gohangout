package output

import (
	"plugin"

	"github.com/childe/gohangout/condition_filter"
	"github.com/childe/gohangout/topology"
	"github.com/golang/glog"
)

func BuildOutput(outputType string, config map[interface{}]interface{}) *topology.OutputBox {
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

	return &topology.OutputBox{
		output,
		condition_filter.NewConditionFilter(config),
	}
}
