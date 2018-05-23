package output

import "github.com/childe/gohangout/condition_filter"

type Output interface {
	Emit(map[string]interface{})
	Pass(map[string]interface{}) bool
	Shutdown()
}

func GetOutput(outputType string, config map[interface{}]interface{}) Output {
	switch outputType {
	case "Stdout":
		return NewStdoutOutput(config)
	case "Elasticsearch":
		return NewElasticsearchOutput(config)
	}
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
