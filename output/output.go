package output

type Output interface {
	Emit(map[string]interface{})
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
