package main

type Input interface {
	readOneEvent() map[string]interface{}
}

func getInput(inputType string, config map[interface{}]interface{}) Input {
	switch inputType {
	case "Stdin":
		//t := reflect.TypeOf(StdinInput{})
		//pt := reflect.New(t)
		//b := pt.Interface().(Input)
		//return b
		return NewStdinInput(config)
	case "Kafka":
		return NewKafkaInput(config)
	}
	return nil
}
