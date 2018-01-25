package filter

import (
	"bytes"
	"reflect"
	"strconv"
	"text/template"

	"github.com/golang/glog"
)

type Filter interface {
	Pass(map[string]interface{}) bool
	Process(map[string]interface{}) (map[string]interface{}, bool)
	PostProcess(map[string]interface{}, bool) map[string]interface{}
}

func GetFilter(filterType string, config map[interface{}]interface{}) Filter {
	switch filterType {
	case "Add":
		return NewAddFilter(config)
	case "Grok":
		return NewGrokFilter(config)
	case "Date":
		return NewDateFilter(config)
	case "Drop":
		return NewDropFilter(config)
	}
	glog.Fatalf("could build %s filter plugin", filterType)
	return nil
}

type BaseFilter struct {
	config       map[interface{}]interface{}
	ifConditions []*template.Template
	ifResult     string
}

func NewBaseFilter(config map[interface{}]interface{}) BaseFilter {
	f := BaseFilter{
		config: config,
	}
	if v, ok := config["if"]; ok {
		f.ifConditions = make([]*template.Template, 0)
		for i, c := range v.([]interface{}) {
			t, err := template.New(strconv.Itoa(i)).Parse(c.(string))
			if err != nil {
				glog.Fatalf("could NOT build template from %s:%s", c, err)
			}
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

func (f *BaseFilter) Pass(event map[string]interface{}) bool {
	if f.ifConditions == nil {
		return true
	}
	for _, c := range f.ifConditions {
		b := bytes.NewBuffer(nil)
		c.Execute(b, event)
		if b.String() != f.ifResult {
			return false
		}
	}
	return true
}
func (f *BaseFilter) Process(event map[string]interface{}) (map[string]interface{}, bool) {
	return event, true
}
func (f *BaseFilter) PostProcess(event map[string]interface{}, success bool) map[string]interface{} {
	if success {
		if remove_fields, ok := f.config["remove_fields"]; ok {
			for _, field := range remove_fields.([]interface{}) {
				delete(event, field.(string))
			}
		}
	} else {
		if v, ok := f.config["failTag"]; ok {
			failTag := v.(string)
			if tags, ok := event["tags"]; ok {
				if reflect.TypeOf(tags).Kind() == reflect.String {
					event["tags"] = []string{tags.(string), failTag}
				} else if reflect.TypeOf(tags).Kind() == reflect.Array {
					event["tags"] = append(tags.([]interface{}), failTag)
				} else {
				}
			} else {
				event["tags"] = failTag
			}
		}
	}
	return event
}
