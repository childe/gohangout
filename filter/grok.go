package filter

import (
	"regexp"

	"github.com/childe/gohangout/value_render"
	"github.com/golang/glog"
)

type Grok struct {
	p           *regexp.Regexp
	subexpNames []string
}

func (grok *Grok) grok(input string) map[string]string {
	if grok.p.MatchString(input) {
		glog.V(5).Infof("grok `%s` match", grok.p)
		rst := make(map[string]string)
		for _, substrings := range grok.p.FindAllStringSubmatch(input, -1) {
			for i, substring := range substrings {
				if i == 0 {
					continue
				}
				rst[grok.subexpNames[i]] = substring
			}
		}
		return rst
	}
	glog.V(5).Infof("grok not `%s` match", grok.p)
	return nil
}

func NewGrok(match string) *Grok {
	p, err := regexp.Compile(match)
	if err != nil {
		glog.Fatalf("could not build Grok:%s", err)
	}
	return &Grok{
		p:           p,
		subexpNames: p.SubexpNames(),
	}
}

type GrokFilter struct {
	BaseFilter

	config    map[interface{}]interface{}
	overwrite bool
	groks     []*Grok
	src       string
	vr        value_render.ValueRender
}

func NewGrokFilter(config map[interface{}]interface{}) *GrokFilter {
	groks := make([]*Grok, 0)
	if matchValue, ok := config["match"]; ok {
		match := matchValue.([]interface{})
		for _, mValue := range match {
			groks = append(groks, NewGrok(mValue.(string)))
		}
	} else {
		glog.Fatal("match must be set in grok filter")
	}

	gf := &GrokFilter{
		BaseFilter: BaseFilter{config},
		config:     config,
		groks:      groks,
		overwrite:  true,
	}

	if overwrite, ok := config["overwrite"]; ok {
		gf.overwrite = overwrite.(bool)
	}

	if srcValue, ok := config["src"]; ok {
		gf.src = srcValue.(string)
	} else {
		gf.src = "[message]"
	}
	gf.vr = value_render.GetValueRender(gf.src)

	return gf
}

func (gf *GrokFilter) Process(event map[string]interface{}) (map[string]interface{}, bool) {
	var input string
	inputI := gf.vr.Render(event)
	if inputI == nil {
		glog.V(5).Infof("(%s) not in event", gf.src)
		return event, true
	} else {
		input = inputI.(string)
		glog.V(100).Infof("input: (%s)", input)
	}

	success := false
	for _, grok := range gf.groks {
		rst := grok.grok(input)
		if rst == nil {
			continue
		}

		for field, value := range rst {
			event[field] = value
		}
		success = true
		break
	}
	return event, success
}
