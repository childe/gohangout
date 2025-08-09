package field_setter

import "reflect"

type MultiLevelFieldSetter struct {
	preFields []string
	lastField string
}

func NewMultiLevelFieldSetter(fields []string) *MultiLevelFieldSetter {
	fieldsLength := len(fields)
	preFields := make([]string, fieldsLength-1)
	for i := range preFields {
		preFields[i] = fields[i]
	}

	return &MultiLevelFieldSetter{
		preFields: preFields,
		lastField: fields[fieldsLength-1],
	}
}

func (fs *MultiLevelFieldSetter) SetField(event map[string]any, value any, field string, overwrite bool) map[string]any {
	current := event
	for _, field := range fs.preFields {
		if value, ok := current[field]; ok {
			if reflect.TypeOf(value).Kind() == reflect.Map {
				current = value.(map[string]any)
			}
		} else {
			a := make(map[string]any)
			current[field] = a
			current = a
		}
	}
	current[fs.lastField] = value
	return event
}
