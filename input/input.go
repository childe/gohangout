package input

type Input interface {
	readOneEvent() map[string]interface{}
	Shutdown()
}

func GetInput(inputType string, config map[interface{}]interface{}) Input {
	switch inputType {
	case "Stdin":
		f := NewStdinInput(config)
		return f
	case "Kafka":
		f := NewKafkaInput(config)
		return f
	case "Random":
		f := NewRandomInput(config)
		return f
	}
	return nil
}
