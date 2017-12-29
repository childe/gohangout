package main

import "regexp"

type ValueRender interface {
	render(map[string]interface{}) interface{}
}

func getValueRender(template string) ValueRender {
	matchGoTemp, _ := regexp.Compile(`{{.*}}`)
	matchp, _ := regexp.Compile(`^(\[.*?\])+$`)
	findp, _ := regexp.Compile(`(\[(.*?)\])`)
	if matchp.Match([]byte(template)) {
		fields := make([]string, 0)
		for _, v := range findp.FindAllStringSubmatch(template, -1) {
			fields = append(fields, v[2])
		}
		return NewMultiLevelValueRender(fields)
	} else {
		if matchGoTemp.Match([]byte(template)) {
			return NewTemplateValueRender(template)
		}
		return NewLiteralValueRender(template)
	}
	return nil
}
