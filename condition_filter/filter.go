package condition_filter

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/childe/gohangout/value_render"
)

type Condition interface {
	Pass(event map[string]interface{}) bool
}

type TemplateCondition struct {
	ifCondition value_render.ValueRender
	ifResult    string
}

func (s *TemplateCondition) Pass(event map[string]interface{}) bool {
	r := s.ifCondition.Render(event)
	if r == nil || r.(string) != s.ifResult {
		return false
	}
	return true
}

func NewTemplateConditionFilter(condition string) *TemplateCondition {
	return &TemplateCondition{
		ifCondition: value_render.GetValueRender(condition),
		ifResult:    "y",
	}
}

type ExistCondition struct {
	pathes []string
}

func NewExistCondition(pathes []string) *ExistCondition {
	return &ExistCondition{pathes}
}

func (c *ExistCondition) Pass(event map[string]interface{}) bool {
	var (
		o      map[string]interface{} = event
		length int                    = len(c.pathes)
	)
	for _, path := range c.pathes[:length-1] {
		if v, ok := o[path]; ok {
			if reflect.TypeOf(v).Kind() == reflect.Map {
				o = v.(map[string]interface{})
			} else {
				return false
			}
		} else {
			return false
		}
	}

	if _, ok := o[c.pathes[length-1]]; ok {
		return true
	}
	return false
}

func NewCondition(c string) Condition {
	if matched, _ := regexp.MatchString(`^{{.*}}$`, c); matched {
		return NewTemplateConditionFilter(c)
	}
	if matched, _ := regexp.MatchString(`^Exist\(.*\)$`, c); matched {
		pathes := make([]string, 0)
		for _, p := range strings.Split(c[6:len(c)-1], ",") {
			pathes = append(pathes, strings.Trim(p, " "))
		}
		return NewExistCondition(pathes)
	}
	return nil
}

type ConditionFilter struct {
	conditions []Condition
}

func NewConditionFilter(config map[interface{}]interface{}) *ConditionFilter {
	f := &ConditionFilter{}

	if v, ok := config["if"]; ok {
		f.conditions = make([]Condition, len(v.([]interface{})))
		for i, c := range v.([]interface{}) {
			f.conditions[i] = NewCondition(c.(string))
		}
	} else {
		f.conditions = nil
	}
	return f
}

func (f *ConditionFilter) Pass(event map[string]interface{}) bool {
	if f.conditions == nil {
		return true
	}

	for _, c := range f.conditions {
		if !c.Pass(event) {
			return false
		}
	}
	return true
}
