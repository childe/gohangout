package value_render

import (
	"regexp"
)

type ValueRender interface {
	Render(map[string]interface{}) interface{}
}

// GetValueRender("message") return NewLiteralValueRender("message")
func GetValueRender(template string) ValueRender {
	matchp, _ := regexp.Compile(`^(\[.*?\])+$`)
	matchGoTemp, _ := regexp.Compile(`{{.*}}`)
	matchESIndex, _ := regexp.Compile(`%{.*?}`) //%{+YYYY.MM.dd}

	if matchp.Match([]byte(template)) {
		findp, _ := regexp.Compile(`(\[(.*?)\])`)
		fields := make([]string, 0)
		for _, v := range findp.FindAllStringSubmatch(template, -1) {
			fields = append(fields, v[2])
		}

		if len(fields) == 1 {
			return NewOneLevelValueRender(fields[0])
		}
		return NewMultiLevelValueRender(fields)
	}
	if matchGoTemp.Match([]byte(template)) {
		return NewTemplateValueRender(template)
	}
	if matchESIndex.Match([]byte(template)) {
		return NewIndexRender(template)
	}

	return NewLiteralValueRender(template)
}

// GetValueRender2("message") return OneLevelValueRender("message")
func GetValueRender2(template string) ValueRender {
	matchp, _ := regexp.Compile(`^(\[.*?\])+$`)
	matchGoTemp, _ := regexp.Compile(`{{.*}}`)
	matchESIndex, _ := regexp.Compile(`%{.*?}`) //%{+YYYY.MM.dd}

	if matchp.Match([]byte(template)) {
		findp, _ := regexp.Compile(`(\[(.*?)\])`)
		fields := make([]string, 0)
		for _, v := range findp.FindAllStringSubmatch(template, -1) {
			fields = append(fields, v[2])
		}

		if len(fields) == 1 {
			return NewOneLevelValueRender(fields[0])
		}
		return NewMultiLevelValueRender(fields)
	}
	if matchGoTemp.Match([]byte(template)) {
		return NewTemplateValueRender(template)
	}
	if matchESIndex.Match([]byte(template)) {
		return NewIndexRender(template)
	}

	return NewOneLevelValueRender(template)
}
