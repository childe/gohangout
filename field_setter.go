package main

import "regexp"

type FieldSetter interface {
	SetField(map[string]interface{}, interface{}, string, bool) map[string]interface{}
}

func NewFieldSetter(template string) FieldSetter {
	fields := make([]string, 0)
	matchp, _ := regexp.Compile(`(\[.*?\])+`)
	findp, _ := regexp.Compile(`(\[(.*?)\])`)
	if matchp.Match([]byte(template)) {
		for _, v := range findp.FindAllStringSubmatch(template, -1) {
			fields = append(fields, v[2])
		}
		return NewMultiLevelFieldSetter(fields)
	} else {
		return NewOneLevelFieldSetter(template)
	}
	return nil
}
