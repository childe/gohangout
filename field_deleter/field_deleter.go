package field_deleter

import "regexp"

type FieldDeleter interface {
	Delete(map[string]interface{})
}

func NewFieldDeleter(template string) FieldDeleter {
	matchp, _ := regexp.Compile(`(\[.*?\])+`)
	findp, _ := regexp.Compile(`(\[(.*?)\])`)
	if matchp.Match([]byte(template)) {
		fields := make([]string, 0)
		for _, v := range findp.FindAllStringSubmatch(template, -1) {
			fields = append(fields, v[2])
		}
		return NewMultiLevelFieldDeleter(fields)
	} else {
		return NewOneLevelFieldDeleter(template)
	}
}
