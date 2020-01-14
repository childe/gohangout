package value_render

import (
	"fmt"
	"regexp"

	"github.com/oliveagle/jsonpath"
)

var matchp, matchGoTemp, matchESIndex, jsonPath *regexp.Regexp

func init() {
	matchp, _ = regexp.Compile(`^(\[.*?\])+$`)
	matchGoTemp, _ = regexp.Compile(`{{.*}}`)
	matchESIndex, _ = regexp.Compile(`%{.*?}`) //%{+YYYY.MM.dd}
	jsonPath, _ = regexp.Compile(`^\$\.`)
}

type ValueRender interface {
	Render(map[string]interface{}) interface{}
}

// getValueRender matches all regexp pattern and return a ValueRender
// return nil if no pattern matched
func getValueRender(template string) ValueRender {
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
	if jsonPath.Match([]byte(template)) {
		pat, err := jsonpath.Compile(template)
		if err != nil {
			panic(fmt.Sprintf("json path compile `%s` error: %s", template, err))
		}
		return &JsonpathRender{pat}
	}

	return nil
}

// GetValueRender return a ValueRender, and return LiteralValueRender if no pattern matched
func GetValueRender(template string) ValueRender {
	r := getValueRender(template)
	if r != nil {
		return r
	}
	return NewLiteralValueRender(template)
}

// GetValueRender2 return a ValueRender, and return OneLevelValueRender("message") if no pattern matched
func GetValueRender2(template string) ValueRender {
	r := getValueRender(template)
	if r != nil {
		return r
	}
	return NewOneLevelValueRender(template)
}
