package value_render

import (
	"reflect"
	"regexp"
)

type MultiLevelValueRender struct {
	preFields []string
	lastField string
}

func NewMultiLevelValueRender(template string) *MultiLevelValueRender {
	findp, _ := regexp.Compile(`(\[(.*?)\])`)
	fields := make([]string, 0)
	for _, v := range findp.FindAllStringSubmatch(template, -1) {
		fields = append(fields, v[2])
	}

	fieldsLength := len(fields)
	preFields := make([]string, fieldsLength-1)
	for i := range preFields {
		preFields[i] = fields[i]
	}

	return &MultiLevelValueRender{
		preFields: preFields,
		lastField: fields[fieldsLength-1],
	}
}

func (vr *MultiLevelValueRender) Render(event map[string]interface{}) interface{} {
	var current map[string]interface{}
	current = event
	for _, field := range vr.preFields {
		if value, ok := current[field]; !ok {
			return nil
		} else {
			if reflect.TypeOf(value).Kind() == reflect.Map {
				current = value.(map[string]interface{})
			} else {
				return nil
			}
		}
	}
	if value, ok := current[vr.lastField]; ok {
		return value
	}
	return nil
}
