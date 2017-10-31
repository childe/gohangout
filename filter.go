package main

type Filter interface {
	process(map[string]interface{}) map[string]interface{}
}

func getFilter(filterType string, config map[interface{}]interface{}) Filter {
	switch filterType {
	case "Add":
		return NewAddFilter(config)
	case "Grok":
		return NewGrokFilter(config)
	}
	return nil
}
