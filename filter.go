package main

import "reflect"

type Filter interface {
	init(map[interface{}]interface{})
	process(map[string]interface{}) map[string]interface{}
}

func getFilter(filterType string, config map[interface{}]interface{}) Filter {
	switch filterType {
	case "Add":
		t := reflect.TypeOf(AddFilter{})
		pt := reflect.New(t)
		b := pt.Interface().(Filter)
		b.init(config)
		return b
	}
	return nil
}
