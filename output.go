package main

type Output interface {
	emit(map[string]interface{})
}

func getOutput(outputType string, config map[interface{}]interface{}) Output {
	switch outputType {
	case "Stdout":
		return NewStdoutOutput(config)
	case "Elasticsearch":
		return NewElasticsearchOutput(config)
	}
	return nil
}
