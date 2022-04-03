package field_setter

import "regexp"

// FieldSetter is the interface that wraps the SetField method.
type FieldSetter interface {
	SetField(event map[string]interface{}, value interface{}, fieldName string, overwrite bool) map[string]interface{}
}

// NewFieldSetter creates a new FieldSetter.
// It returns OneLevelFieldSetter if [xxx] passed
// It returns MultiLevelFieldSetter if [xxx][yyy] passed
func NewFieldSetter(template string) FieldSetter {
	matchp, _ := regexp.Compile(`(\[.*?\])+`)
	findp, _ := regexp.Compile(`(\[(.*?)\])`)
	if matchp.Match([]byte(template)) {
		fields := make([]string, 0)
		for _, v := range findp.FindAllStringSubmatch(template, -1) {
			fields = append(fields, v[2])
		}
		if len(fields) == 1 {
			return NewOneLevelFieldSetter(fields[0])
		}
		return NewMultiLevelFieldSetter(fields)
	} else {
		return NewOneLevelFieldSetter(template)
	}
}
