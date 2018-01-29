package filter

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/childe/gohangout/value_render"
	"github.com/golang/glog"
)

func getAllPatternsFromFile(filename string) map[string]string {
	var patterns map[string]string = make(map[string]string)
	f, err := os.Open(filename)
	if err != nil {
		glog.Fatal(err)
	}
	r := bufio.NewReader(f)
	for {
		line, isPrefix, err := r.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			glog.Fatalf("read pattenrs error:%s", err)
		}
		if isPrefix == true {
			glog.Fatal("readline prefix")
		}
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		ss := strings.SplitN(string(line), " ", 2)
		if len(ss) != 2 {
			glog.Fatalf("splited `%s` length !=2", string(line))
		}
		patterns[ss[0]] = ss[1]
	}
	return patterns
}

func replaceFunc(s string) string {
	patterns := getAllPatternsFromFile("patterns")
	p, err := regexp.Compile(`%{(\w+?)(?::(\w+?))?}`)
	if err != nil {
		glog.Fatal(err)
	}
	rst := p.FindAllStringSubmatch(s, -1)
	if len(rst) != 1 {
		glog.Fatal("!=1")
	}
	if pattern, ok := patterns[rst[0][1]]; ok {
		if rst[0][2] == "" {
			return fmt.Sprintf("(%s)", pattern)
		} else {
			return fmt.Sprintf("(?P<%s>%s)", rst[0][2], pattern)
		}
	} else {
		glog.Fatalf("`%s` could not be found", rst[0][1])
		return ""
	}
}

func translateMatchPattern(s string) string {
	p, err := regexp.Compile(`%{\w+?(:\w+?)?}`)
	if err != nil {
		glog.Fatal(err)
	}
	var r string = ""
	for {
		r = p.ReplaceAllStringFunc(s, replaceFunc)
		if r == s {
			return r
		}
		s = r
	}
	return r
}

type Grok struct {
	p           *regexp.Regexp
	subexpNames []string

	patterns     map[string]string
	patternPaths []string
}

func (grok *Grok) grok(input string) map[string]string {
	glog.V(5).Infof("grok `%s` match", grok.p)
	rst := make(map[string]string)
	for _, substrings := range grok.p.FindAllStringSubmatch(input, -1) {
		glog.Info(substrings)
		for i, substring := range substrings {
			glog.Info(substring)
			if grok.subexpNames[i] == "" {
				continue
			}
			rst[grok.subexpNames[i]] = substring
		}
	}
	return rst
}

func NewGrok(match string) *Grok {
	//finalPattern := translateMatchPattern(match)
	//glog.Infof("final pattern:%s", finalPattern)
	p, err := regexp.Compile(match)
	glog.Info(p.SubexpNames())
	glog.Info(len(p.SubexpNames()))
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
		BaseFilter: NewBaseFilter(config),
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
		gf.src = "message"
	}
	gf.vr = value_render.GetValueRender2(gf.src)

	return gf
}

func (gf *GrokFilter) Process(event map[string]interface{}) (map[string]interface{}, bool) {
	var input string
	inputI := gf.vr.Render(event)
	if inputI == nil {
		glog.V(5).Infof("(%s) not in event", gf.src)
		return event, false
	} else {
		input = inputI.(string)
		glog.V(100).Infof("input: (%s)", input)
	}

	success := false
	for _, grok := range gf.groks {
		rst := grok.grok(input)
		if len(rst) == 0 {
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
