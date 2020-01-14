package condition_filter

import (
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/oliveagle/jsonpath"

	"github.com/childe/gohangout/value_render"
	"github.com/golang/glog"
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
	paths []string
}

func NewExistCondition(paths []string) *ExistCondition {
	return &ExistCondition{paths}
}

func (c *ExistCondition) Pass(event map[string]interface{}) bool {
	var (
		o      map[string]interface{} = event
		length int                    = len(c.paths)
	)
	for _, path := range c.paths[:length-1] {
		if v, ok := o[path]; ok && v != nil {
			if reflect.TypeOf(v).Kind() == reflect.Map {
				o = v.(map[string]interface{})
			} else {
				return false
			}
		} else {
			return false
		}
	}

	if _, ok := o[c.paths[length-1]]; ok {
		return true
	}
	return false
}

type EQCondition struct {
	pat   *jsonpath.Compiled
	paths []string
	value interface{}
	fn    int
}

func NewEQCondition(c string) (*EQCondition, error) {
	var (
		pat   *jsonpath.Compiled
		paths []string
		value string
		err   error
	)

	if strings.HasPrefix(c, `EQ($.`) {
		p := regexp.MustCompile(`^EQ\((\$\..*),(.*)\)$`)
		r := p.FindStringSubmatch(c)
		if len(r) != 3 {
			return nil, fmt.Errorf("split jsonpath pattern/value error in `%s`", c)
		}

		if pat, err = jsonpath.Compile(r[1]); err != nil {
			return nil, err
		}

		value = r[2]
	} else {
		paths = make([]string, 0)
		c = strings.TrimSuffix(strings.TrimPrefix(c, "EQ("), ")")
		for _, p := range strings.Split(c, ",") {
			paths = append(paths, strings.Trim(p, " "))
		}
		value = paths[len(paths)-1]
		paths = paths[:len(paths)-1]
	}

	if value[0] == '"' && value[len(value)-1] == '"' {
		value = value[1 : len(value)-1]
		return &EQCondition{pat, paths, value, len(paths)}, nil
	}
	if strings.Contains(value, ".") {
		if s, err := strconv.ParseFloat(value, 64); err == nil {
			return &EQCondition{pat, paths, s, len(paths)}, nil
		} else {
			return nil, err
		}
	}
	if s, err := strconv.ParseInt(value, 0, 32); err == nil {
		return &EQCondition{pat, paths, int(s), len(paths)}, nil
	} else {
		return nil, err
	}
	return &EQCondition{pat, paths, value, len(paths)}, nil
}

func (c *EQCondition) Pass(event map[string]interface{}) bool {
	if c.pat != nil {
		v, err := c.pat.Lookup(event)
		return err == nil && v == c.value
	}

	var (
		o map[string]interface{} = event
	)

	for _, path := range c.paths[:c.fn-1] {
		if v, ok := o[path]; ok && v != nil {
			if o, ok = v.(map[string]interface{}); !ok {
				return false
			}
		} else {
			return false
		}
	}

	if v, ok := o[c.paths[c.fn-1]]; ok {
		return v == c.value
	}
	return false
}

type HasPrefixCondition struct {
	pat    *jsonpath.Compiled
	paths  []string
	prefix string
}

func NewHasPrefixCondition(c string) (*HasPrefixCondition, error) {
	if strings.HasPrefix(c, `HasPrefix($.`) {
		p := regexp.MustCompile(`^HasPrefix\((\$\..*),"(.*)"\)$`)
		r := p.FindStringSubmatch(c)
		if len(r) != 3 {
			return nil, fmt.Errorf("split jsonpath pattern/value error in `%s`", c)
		}

		value := r[2]
		pat, err := jsonpath.Compile(r[1])
		if err != nil {
			return nil, err
		}

		return &HasPrefixCondition{pat, nil, value}, nil
	}

	paths := make([]string, 0)
	c = strings.TrimSuffix(strings.TrimPrefix(c, "HasPrefix("), ")")
	for _, p := range strings.Split(c, ",") {
		paths = append(paths, strings.Trim(p, " "))
	}
	value := paths[len(paths)-1]
	paths = paths[:len(paths)-1]
	return &HasPrefixCondition{nil, paths, value}, nil
}

func (c *HasPrefixCondition) Pass(event map[string]interface{}) bool {
	if c.pat != nil {
		v, err := c.pat.Lookup(event)
		return err == nil && strings.HasPrefix(v.(string), c.prefix)
	}

	var (
		o      map[string]interface{} = event
		length int                    = len(c.paths)
	)

	for _, path := range c.paths[:length-1] {
		if v, ok := o[path]; ok && v != nil {
			if reflect.TypeOf(v).Kind() == reflect.Map {
				o = v.(map[string]interface{})
			} else {
				return false
			}
		} else {
			return false
		}
	}

	if v, ok := o[c.paths[length-1]]; ok && v != nil {
		if reflect.TypeOf(v).Kind() == reflect.String {
			return strings.HasPrefix(v.(string), c.prefix)
		}
	}
	return false
}

type HasSuffixCondition struct {
	pat    *jsonpath.Compiled
	paths  []string
	suffix string
}

func NewHasSuffixCondition(c string) (*HasSuffixCondition, error) {
	if strings.HasPrefix(c, `HasSuffix($.`) {
		p := regexp.MustCompile(`^HasSuffix\((\$\..*),"(.*)"\)$`)
		r := p.FindStringSubmatch(c)
		if len(r) != 3 {
			return nil, fmt.Errorf("split jsonpath pattern/value error in `%s`", c)
		}

		value := r[2]
		pat, err := jsonpath.Compile(r[1])
		if err != nil {
			return nil, err
		}

		return &HasSuffixCondition{pat, nil, value}, nil
	}

	paths := make([]string, 0)
	c = strings.TrimSuffix(strings.TrimPrefix(c, "HasSuffix("), ")")
	for _, p := range strings.Split(c, ",") {
		paths = append(paths, strings.Trim(p, " "))
	}
	value := paths[len(paths)-1]
	paths = paths[:len(paths)-1]
	return &HasSuffixCondition{nil, paths, value}, nil
}

func (c *HasSuffixCondition) Pass(event map[string]interface{}) bool {
	if c.pat != nil {
		v, err := c.pat.Lookup(event)
		return err == nil && strings.HasSuffix(v.(string), c.suffix)
	}

	var (
		o      map[string]interface{} = event
		length int                    = len(c.paths)
	)

	for _, path := range c.paths[:length-1] {
		if v, ok := o[path]; ok && v != nil {
			if reflect.TypeOf(v).Kind() == reflect.Map {
				o = v.(map[string]interface{})
			} else {
				return false
			}
		} else {
			return false
		}
	}

	if v, ok := o[c.paths[length-1]]; ok && v != nil {
		if reflect.TypeOf(v).Kind() == reflect.String {
			return strings.HasSuffix(v.(string), c.suffix)
		}
	}
	return false
}

type ContainsCondition struct {
	pat       *jsonpath.Compiled
	paths     []string
	substring string
}

func NewContainsCondition(c string) (*ContainsCondition, error) {
	if strings.HasPrefix(c, `Contains($.`) {
		p := regexp.MustCompile(`^Contains\((\$\..*),"(.*)"\)$`)
		r := p.FindStringSubmatch(c)
		if len(r) != 3 {
			return nil, fmt.Errorf("split jsonpath pattern/value error in `%s`", c)
		}

		value := r[2]
		pat, err := jsonpath.Compile(r[1])
		if err != nil {
			return nil, err
		}

		return &ContainsCondition{pat, nil, value}, nil
	}
	paths := make([]string, 0)
	c = strings.TrimSuffix(strings.TrimPrefix(c, "Contains("), ")")
	for _, p := range strings.Split(c, ",") {
		paths = append(paths, strings.Trim(p, " "))
	}
	value := paths[len(paths)-1]
	paths = paths[:len(paths)-1]
	return &ContainsCondition{nil, paths, value}, nil
}

func (c *ContainsCondition) Pass(event map[string]interface{}) bool {
	if c.pat != nil {
		v, err := c.pat.Lookup(event)
		return err == nil && strings.Contains(v.(string), c.substring)
	}

	var (
		o      map[string]interface{} = event
		length int                    = len(c.paths)
	)

	for _, path := range c.paths[:length-1] {
		if v, ok := o[path]; ok && v != nil {
			if reflect.TypeOf(v).Kind() == reflect.Map {
				o = v.(map[string]interface{})
			} else {
				return false
			}
		} else {
			return false
		}
	}

	if v, ok := o[c.paths[length-1]]; ok && v != nil {
		if reflect.TypeOf(v).Kind() == reflect.String {
			return strings.Contains(v.(string), c.substring)
		}
	}
	return false
}

type ContainsAnyCondition struct {
	paths     []string
	substring string
}

func NewContainsAnyCondition(paths []string, substring string) *ContainsAnyCondition {
	return &ContainsAnyCondition{paths, substring}
}

func (c *ContainsAnyCondition) Pass(event map[string]interface{}) bool {
	var (
		o      map[string]interface{} = event
		length int                    = len(c.paths)
	)

	for _, path := range c.paths[:length-1] {
		if v, ok := o[path]; ok && v != nil {
			if reflect.TypeOf(v).Kind() == reflect.Map {
				o = v.(map[string]interface{})
			} else {
				return false
			}
		} else {
			return false
		}
	}

	if v, ok := o[c.paths[length-1]]; ok && v != nil {
		if reflect.TypeOf(v).Kind() == reflect.String {
			return strings.ContainsAny(v.(string), c.substring)
		}
	}
	return false
}

type MatchCondition struct {
	pat    *jsonpath.Compiled
	paths  []string
	regexp *regexp.Regexp
}

func NewMatchCondition(c string) (*MatchCondition, error) {
	if strings.HasPrefix(c, `Match($.`) {
		p := regexp.MustCompile(`^Match\((\$\..*),"(.*)"\)$`)
		r := p.FindStringSubmatch(c)
		if len(r) != 3 {
			return nil, fmt.Errorf("split jsonpath pattern/value error in `%s`", c)
		}

		pat, err := jsonpath.Compile(r[1])
		if err != nil {
			return nil, err
		}

		value := r[2]
		regexp, err := regexp.Compile(value)
		if err != nil {
			return nil, err
		}

		return &MatchCondition{pat, nil, regexp}, nil
	}

	paths := make([]string, 0)
	c = strings.TrimSuffix(strings.TrimPrefix(c, "Match("), ")")
	for _, p := range strings.Split(c, ",") {
		paths = append(paths, strings.Trim(p, " "))
	}
	value := paths[len(paths)-1]
	paths = paths[:len(paths)-1]
	regexp, err := regexp.Compile(value)
	if err != nil {
		return nil, err
	}
	return &MatchCondition{nil, paths, regexp}, nil
}

func (c *MatchCondition) Pass(event map[string]interface{}) bool {
	if c.pat != nil {
		v, err := c.pat.Lookup(event)
		return err == nil && c.regexp.MatchString(v.(string))
	}
	var (
		o      map[string]interface{} = event
		length int                    = len(c.paths)
	)

	for _, path := range c.paths[:length-1] {
		if v, ok := o[path]; ok && v != nil {
			if reflect.TypeOf(v).Kind() == reflect.Map {
				o = v.(map[string]interface{})
			} else {
				return false
			}
		} else {
			return false
		}
	}

	if v, ok := o[c.paths[length-1]]; ok && v != nil {
		if reflect.TypeOf(v).Kind() == reflect.String {
			return c.regexp.MatchString(v.(string))
		}
	}
	return false
}

type RandomCondition struct {
	value int
}

func NewRandomCondition(value int) *RandomCondition {
	rand.Seed(time.Now().UnixNano())
	return &RandomCondition{value}
}

func (c *RandomCondition) Pass(event map[string]interface{}) bool {
	return rand.Intn(c.value) == 0
}

type BeforeCondition struct {
	d time.Duration
}

func NewBeforeCondition(value string) *BeforeCondition {
	d, err := time.ParseDuration(value)
	if err != nil {
		glog.Fatalf("could not parse %s to duration: %s", value, err)
	}
	return &BeforeCondition{d}
}

func (c *BeforeCondition) Pass(event map[string]interface{}) bool {
	timestamp := event["@timestamp"]
	if timestamp == nil || reflect.TypeOf(timestamp).String() != "time.Time" {
		return false
	}
	return timestamp.(time.Time).Before(time.Now().Add(c.d))
}

type AfterCondition struct {
	d time.Duration
}

func NewAfterCondition(value string) *AfterCondition {
	d, err := time.ParseDuration(value)
	if err != nil {
		glog.Fatalf("could not parse %s to duration: %s", value, err)
	}
	return &AfterCondition{d}
}

func (c *AfterCondition) Pass(event map[string]interface{}) bool {
	timestamp := event["@timestamp"]
	if timestamp == nil || reflect.TypeOf(timestamp).String() != "time.Time" {
		return false
	}
	return timestamp.(time.Time).After(time.Now().Add(c.d))
}

func NewCondition(c string) Condition {
	original_c := c

	c = strings.Trim(c, " ")

	if matched, _ := regexp.MatchString(`^{{.*}}$`, c); matched {
		return NewTemplateConditionFilter(c)
	}

	if root, err := parseBoolTree(c); err != nil {
		glog.Errorf("could not build Condition from `%s` : %s", original_c, err)
		return nil
	} else {
		return root
	}
}

func NewSingleCondition(c string) (Condition, error) {
	original_c := c

	// Exist
	if matched, _ := regexp.MatchString(`^Exist\(.*\)$`, c); matched {
		c = strings.TrimSuffix(strings.TrimPrefix(c, "Exist("), ")")
		paths := make([]string, 0)
		for _, p := range strings.Split(c, ",") {
			paths = append(paths, strings.Trim(p, " "))
		}
		return NewExistCondition(paths), nil
	}

	// EQ
	if matched, _ := regexp.MatchString(`^EQ\(.*\)$`, c); matched {
		return NewEQCondition(c)
	}

	// HasPrefix
	if matched, _ := regexp.MatchString(`^HasPrefix\(.*\)$`, c); matched {
		return NewHasPrefixCondition(c)
	}

	// HasSuffix
	if matched, _ := regexp.MatchString(`^HasSuffix\(.*\)$`, c); matched {
		return NewHasSuffixCondition(c)
	}

	// Contains
	if matched, _ := regexp.MatchString(`^Contains\(.*\)$`, c); matched {
		return NewContainsCondition(c)
	}

	// ContainsAny
	if matched, _ := regexp.MatchString(`^ContainsAny\(.*\)$`, c); matched {
		paths := make([]string, 0)
		c = strings.TrimSuffix(strings.TrimPrefix(c, "ContainsAny("), ")")
		for _, p := range strings.Split(c, ",") {
			paths = append(paths, strings.Trim(p, " "))
		}
		value := paths[len(paths)-1]
		paths = paths[:len(paths)-1]
		return NewContainsAnyCondition(paths, value), nil
	}

	// Match
	if matched, _ := regexp.MatchString(`^Match\(.*\)$`, c); matched {
		return NewMatchCondition(c)
	}

	// Random
	if matched, _ := regexp.MatchString(`^Random\(.*\)$`, c); matched {
		c = strings.TrimSuffix(strings.TrimPrefix(c, "Random("), ")")
		if value, err := strconv.ParseInt(c, 0, 32); err != nil {
			return nil, err
		} else {
			return NewRandomCondition(int(value)), nil
		}
	}

	// Before
	if matched, _ := regexp.MatchString(`^Before\(.*\)$`, c); matched {
		c = strings.TrimSuffix(strings.TrimPrefix(c, "Before("), ")")
		return NewBeforeCondition(c), nil
	}

	// After
	if matched, _ := regexp.MatchString(`^After\(.*\)$`, c); matched {
		c = strings.TrimSuffix(strings.TrimPrefix(c, "After("), ")")
		return NewAfterCondition(c), nil
	}

	return nil, fmt.Errorf("could not build Condition from `%s`", original_c)
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

type OPNode struct {
	op        int
	left      *OPNode
	right     *OPNode
	condition Condition //leaf node has condition
	pos       int
}

func (root *OPNode) Pass(event map[string]interface{}) bool {
	if root.condition != nil {
		return root.condition.Pass(event)
	}

	if root.op == _op_and {
		return root.left.Pass(event) && root.right.Pass(event)
	}
	if root.op == _op_or {
		return root.left.Pass(event) || root.right.Pass(event)
	}
	if root.op == _op_not {
		return !root.right.Pass(event)
	}
	return false
}
