package main

import "reflect"

type MultiLevelFieldSetter struct {
	preFields    []string
	lastField    string
	fieldsLength int
}

func NewMultiLevelFieldSetter(fields []string) *MultiLevelFieldSetter {
	fieldsLength := len(fields)
	preFields := make([]string, fieldsLength-1)
	for i := range preFields {
		preFields[i] = fields[i]
	}

	r := &MultiLevelFieldSetter{
		preFields:    preFields,
		fieldsLength: fieldsLength,
		lastField:    fields[fieldsLength-1],
	}
	return r
}

func (fs *MultiLevelFieldSetter) SetField(event map[string]interface{}, value interface{}, field string, overwrite bool) map[string]interface{} {
	current := event
	for _, field := range fs.preFields {
		if value, ok := current[field]; ok {
			if reflect.TypeOf(value).Kind() == reflect.Map {
				current = value.(map[string]interface{})
			}
		} else {
			a := make(map[string]interface{})
			current[field] = a
			current = a
		}
	}
	current[fs.lastField] = value
	return event
}
