package main

import "reflect"

type Output interface {
	emit(map[string]interface{})
}

func getOutput(outputType string) Output {
	switch outputType {
	case "Stdout":
		t := reflect.TypeOf(StdoutOutput{})
		pt := reflect.New(t)
		b := pt.Interface().(Output)
		return b
	}
	return nil
}
