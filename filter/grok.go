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

	"github.com/childe/gohangout/topology"
	"github.com/childe/gohangout/value_render"
	"github.com/mitchellh/mapstructure"
	"k8s.io/klog/v2"
)

func (grok *Grok) loadPattern(filename string) {
	var r *bufio.Reader
	if strings.HasPrefix(filename, "http://") || strings.HasPrefix(filename, "https://") {
		resp, err := http.Get(filename)
		if err != nil {
			klog.Fatalf("load pattern error:%s", err)
		}
		defer resp.Body.Close()
		r = bufio.NewReader(resp.Body)
	} else {
		f, err := os.Open(filename)
		if err != nil {
			klog.Fatalf("load pattern error:%s", err)
		}
		r = bufio.NewReader(f)
	}
	for {
		line, isPrefix, err := r.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			klog.Fatalf("read pattenrs error:%s", err)
		}
		if isPrefix {
			klog.Fatal("readline prefix")
		}
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		ss := strings.SplitN(string(line), " ", 2)
		if len(ss) != 2 {
			klog.Fatalf("splited `%s` length !=2", string(line))
		}
		grok.patterns[ss[0]] = ss[1]
	}
}

func (grok *Grok) loadPatterns() {
	for _, path := range grok.patternPaths {
		files, err := getFiles(path)
		if err != nil {
			klog.Fatalf("build grok filter error: %s", err)
		}
		for _, file := range files {
			grok.loadPattern(file)
		}
	}
	klog.V(5).Infof("patterns:%s", grok.patterns)
}

func getFiles(filepath string) ([]string, error) {
	if strings.HasPrefix(filepath, "http://") || strings.HasPrefix(filepath, "https://") {
		return []string{filepath}, nil
	}

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
		klog.Fatal(err)
	}
	rst := p.FindAllStringSubmatch(s, -1)
	if len(rst) != 1 {
		klog.Fatalf("sub match in `%s` != 1", s)
	}
	if pattern, ok := grok.patterns[rst[0][1]]; ok {
		if rst[0][2] == "" {
			return fmt.Sprintf("(%s)", pattern)
		} else {
			return fmt.Sprintf("(?P<%s>%s)", rst[0][2], pattern)
		}
	} else {
		klog.Fatalf("`%s` could not be found", rst[0][1])
		return ""
	}
}

func (grok *Grok) translateMatchPattern(s string) string {
	p, err := regexp.Compile(`%{\w+?(:\w+?)?}`)
	if err != nil {
		klog.Fatal(err)
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
	klog.Infof("final pattern:%s", finalPattern)
	p, err := regexp.Compile(finalPattern)
	if err != nil {
		klog.Fatalf("could not build Grok:%s", err)
	}
	grok.p = p
	grok.subexpNames = p.SubexpNames()

	return grok
}

// GrokConfig defines the configuration structure for Grok filter
type GrokConfig struct {
	Src          string   `mapstructure:"src"`
	Target       string   `mapstructure:"target"`
	Match        []string `mapstructure:"match"`
	PatternPaths []string `mapstructure:"pattern_paths"`
	IgnoreBlank  bool     `mapstructure:"ignore_blank"`
	Overwrite    bool     `mapstructure:"overwrite"`
}

type GrokFilter struct {
	config    map[any]any
	overwrite bool
	groks     []*Grok
	target    string
	src       string
	vr        value_render.ValueRender
}

func init() {
	Register("Grok", newGrokFilter)
}

func newGrokFilter(config map[any]any) topology.Filter {
	// Parse configuration using mapstructure
	var grokConfig GrokConfig
	// Set default values
	grokConfig.Src = "message"
	grokConfig.IgnoreBlank = true
	grokConfig.Overwrite = true

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &grokConfig,
		ErrorUnused:      false,
	})
	if err != nil {
		klog.Fatalf("Grok filter: failed to create config decoder: %v", err)
	}

	if err := decoder.Decode(config); err != nil {
		klog.Fatalf("Grok filter configuration error: %v", err)
	}

	// Validate required fields
	if grokConfig.Match == nil || len(grokConfig.Match) == 0 {
		klog.Fatal("Grok filter: 'match' is required and cannot be empty")
	}

	// Create Grok instances
	groks := make([]*Grok, 0)
	for _, pattern := range grokConfig.Match {
		groks = append(groks, NewGrok(pattern, grokConfig.PatternPaths, grokConfig.IgnoreBlank))
	}

	gf := &GrokFilter{
		config:    config,
		groks:     groks,
		overwrite: grokConfig.Overwrite,
		target:    grokConfig.Target,
		src:       grokConfig.Src,
	}
	gf.vr = value_render.GetValueRender2(gf.src)

	return gf
}

func (gf *GrokFilter) Filter(event map[string]any) (map[string]any, bool) {
	var input string
	inputI, err := gf.vr.Render(event)
	if err != nil || inputI == nil {
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
			if gf.overwrite {
				for field, value := range rst {
					event[field] = value
				}
			} else {
				for field, value := range rst {
					if _, exists := event[field]; !exists {
						event[field] = value
					}
				}
			}
		} else {
			target := make(map[string]any)
			for field, value := range rst {
				target[field] = value
			}
			if gf.overwrite {
				event[gf.target] = target
			} else {
				if _, exists := event[gf.target]; !exists {
					event[gf.target] = target
				}
			}
		}
		return event, true
	}
	return event, false
}
