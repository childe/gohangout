package main

import "reflect"

type Input interface {
	init(map[interface{}]interface{})
	readOneEvent() map[string]interface{}
}

func getInput(inputType string) Input {
	switch inputType {
	case "Stdin":
		t := reflect.TypeOf(StdinInput{})
		pt := reflect.New(t)
		b := pt.Interface().(Input)
		return b
	case "Kafka":
		return nil
	}
	return nil
}
