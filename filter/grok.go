package filter

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/childe/gohangout/value_render"
	"github.com/golang/glog"
)

func (grok *Grok) loadPattern(filename string) {
	var r *bufio.Reader
	if strings.HasPrefix(filename, "http://") || strings.HasPrefix(filename, "https://") {
		resp, err := http.Get(filename)
		if err != nil {
			glog.Fatalf("load pattern error:%s", err)
		}
		defer resp.Body.Close()
		r = bufio.NewReader(resp.Body)
	} else {
		f, err := os.Open(filename)
		if err != nil {
			glog.Fatalf("load pattern error:%s", err)
		}
		r = bufio.NewReader(f)
	}
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
		grok.patterns[ss[0]] = ss[1]
	}
}

func (grok *Grok) loadPatterns() {
	for _, path := range grok.patternPaths {
		files, err := getFiles(path)
		if err != nil {
			glog.Fatalf("build grok filter error: %s", err)
		}
		for _, file := range files {
			grok.loadPattern(file)
		}
	}
	glog.V(5).Infof("patterns:%s", grok.patterns)
}

func getFiles(filepath string) ([]string, error) {
	fi, err := os.Stat(filepath)
	if err != nil {
		return nil, err
	}

	if !fi.IsDir() {
		return []string{filepath}, nil
	}

	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	list, err := f.Readdir(-1)
	f.Close()

	if err != nil {
		return nil, err
	}
	files := make([]string, 0)
	for _, l := range list {
		if l.Mode().IsRegular() {
			files = append(files, path.Join(filepath, l.Name()))
		}
	}
	return files, nil
}

func (grok *Grok) replaceFunc(s string) string {
	p, err := regexp.Compile(`%{(\w+?)(?::(\w+?))?}`)
	if err != nil {
		glog.Fatal(err)
	}
	rst := p.FindAllStringSubmatch(s, -1)
	if len(rst) != 1 {
		glog.Fatalf("sub match in `%s` != 1", s)
	}
	if pattern, ok := grok.patterns[rst[0][1]]; ok {
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

func (grok *Grok) translateMatchPattern(s string) string {
	p, err := regexp.Compile(`%{\w+?(:\w+?)?}`)
	if err != nil {
		glog.Fatal(err)
	}
	var r string = ""
	for {
		r = p.ReplaceAllStringFunc(s, grok.replaceFunc)
		if r == s {
			return r
		}
		s = r
	}
}

type Grok struct {
	p           *regexp.Regexp
	subexpNames []string
	ignoreBlank bool

	patterns     map[string]string
	patternPaths []string
}

func (grok *Grok) grok(input string) map[string]string {
	rst := make(map[string]string)
	for i, substring := range grok.p.FindStringSubmatch(input) {
		if grok.subexpNames[i] == "" {
			continue
		}
		if grok.ignoreBlank && substring == "" {
			continue
		}
		rst[grok.subexpNames[i]] = substring
	}
	return rst
}

func NewGrok(match string, patternPaths []string, ignoreBlank bool) *Grok {
	grok := &Grok{
		patternPaths: patternPaths,
		patterns:     make(map[string]string),
		ignoreBlank:  ignoreBlank,
	}
	grok.loadPatterns()

	finalPattern := grok.translateMatchPattern(match)
	glog.Infof("final pattern:%s", finalPattern)
	p, err := regexp.Compile(finalPattern)
	if err != nil {
		glog.Fatalf("could not build Grok:%s", err)
	}
	grok.p = p
	grok.subexpNames = p.SubexpNames()

	return grok
}

type GrokFilter struct {
	config    map[interface{}]interface{}
	overwrite bool
	groks     []*Grok
	target    string
	src       string
	vr        value_render.ValueRender
}

func (l *MethodLibrary) NewGrokFilter(config map[interface{}]interface{}) *GrokFilter {
	var patternPaths []string = make([]string, 0)
	if i, ok := config["pattern_paths"]; ok {
		for _, p := range i.([]interface{}) {
			patternPaths = append(patternPaths, p.(string))
		}
	}
	ignoreBlank := true
	if i, ok := config["ignore_blank"]; ok {
		ignoreBlank = i.(bool)
	}
	groks := make([]*Grok, 0)
	if matchValue, ok := config["match"]; ok {
		match := matchValue.([]interface{})
		for _, mValue := range match {
			groks = append(groks, NewGrok(mValue.(string), patternPaths, ignoreBlank))
		}
	} else {
		glog.Fatal("match must be set in grok filter")
	}

	gf := &GrokFilter{
		config:    config,
		groks:     groks,
		overwrite: true,
		target:    "",
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

	if target, ok := config["target"]; ok {
		gf.target = target.(string)
	}

	return gf
}

func (gf *GrokFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	var input string
	inputI := gf.vr.Render(event)
	if inputI == nil {
		return event, false
	} else {
		input = inputI.(string)
	}

	for _, grok := range gf.groks {
		rst := grok.grok(input)
		if len(rst) == 0 {
			continue
		}

		if gf.target == "" {
			for field, value := range rst {
				event[field] = value
			}
		} else {
			target := make(map[string]interface{})
			for field, value := range rst {
				target[field] = value
			}
			event[gf.target] = target
		}
		return event, true
	}
	return event, false
}
