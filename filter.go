package main

type Filter interface {
	process(map[string]interface{}) map[string]interface{}
}

func getFilter(filterType string, config map[interface{}]interface{}) Filter {
	switch filterType {
	case "Add":
		//t := reflect.TypeOf(AddFilter{})
		//pt := reflect.New(t)
		//b := pt.Interface().(Filter)
		//b.init(config)
		//return b
		return NewAddFilter(config)
	case "Grok":
		return NewGrokFilter(config)
	}
	return nil
}
