package main

import (
	"regexp"

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
	config    map[interface{}]interface{}
	overwrite bool
	groks     []*Grok
	src       string
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

	filter := &GrokFilter{
		config:    config,
		overwrite: true,
		groks:     groks,
	}
	if overwrite, ok := config["overwrite"]; ok {
		filter.overwrite = overwrite.(bool)
	}

	if srcValue, ok := config["src"]; ok {
		filter.src = srcValue.(string)
	} else {
		filter.src = "message"
	}

	return filter
}

func (plugin *GrokFilter) process(event map[string]interface{}) map[string]interface{} {
	var input string
	if inputValue, ok := event[plugin.src]; !ok {
		glog.V(5).Infof("(%s) not in event", plugin.src)
		return event
	} else {
		input = inputValue.(string)
		glog.V(10).Infof("input: (%s)", input)
	}
	for _, grok := range plugin.groks {
		rst := grok.grok(input)
		if rst == nil {
			continue
		}

		for field, value := range rst {
			event[field] = value
		}
	}
	return event
}
