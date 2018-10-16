package condition_filter

import "github.com/childe/gohangout/value_render"

type ConditionFilter struct {
	ifConditions []value_render.ValueRender
	ifResult     string
}

func NewConditionFilter(config map[interface{}]interface{}) *ConditionFilter {
	f := &ConditionFilter{}

	if v, ok := config["if"]; ok {
		f.ifConditions = make([]value_render.ValueRender, 0)
		for _, c := range v.([]interface{}) {
			t := value_render.GetValueRender(c.(string))
			f.ifConditions = append(f.ifConditions, t)
		}
	} else {
		f.ifConditions = nil
	}

	if v, ok := config["ifResult"]; ok {
		f.ifResult = v.(string)
	} else {
		f.ifResult = "y"
	}
	return f
}

func (f *ConditionFilter) Pass(event map[string]interface{}) bool {
	if f.ifConditions == nil {
		return true
	}

	for _, c := range f.ifConditions {
		r := c.Render(event)
		if r == nil || r.(string) != f.ifResult {
			return false
		}
	}
	return true
}
