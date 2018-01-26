package value_render

import (
	"encoding/json"
	"errors"
	"reflect"
	"text/template"
	"time"

	"github.com/golang/glog"
)

type TemplateValueRender struct {
	tmpl *template.Template
}

var GOHANGOUT_TYPE_UNKNOWN_ERROR error = errors.New("field type unknown, it must be of json.Number|Int64|Int|int8")

var funcMap = template.FuncMap{}

func convertToInt(x interface{}) (int, error) {
	if reflect.TypeOf(x).String() == "json.Number" {
		b, _ := x.(json.Number).Int64()
		return int(b), nil
	} else if reflect.TypeOf(x).Kind() == reflect.Int64 {
		return int(x.(int64)), nil
	} else if reflect.TypeOf(x).Kind() == reflect.Int {
		return x.(int), nil
	} else if reflect.TypeOf(x).Kind() == reflect.Int8 {
		return int(x.(int8)), nil
	}
	return 0, GOHANGOUT_TYPE_UNKNOWN_ERROR
}

func init() {
	funcMap["now"] = func() int64 { return time.Now().UnixNano() / 1000000 }
	funcMap["timestamp"] = func(event map[string]interface{}) int64 {
		timestamp := event["@timestamp"]
		if timestamp == nil {
			//return time.Now().UnixNano() / 1000000
			return 0
		}
		if reflect.TypeOf(timestamp).String() == "time.Time" {
			return timestamp.(time.Time).UnixNano() / 1000000
		}
		return 0
	}

	funcMap["plus"] = func(x, y interface{}) (int, error) {
		a, err := convertToInt(x)
		if err != nil {
			return 0, err
		}
		b, err := convertToInt(y)
		if err != nil {
			return 0, err
		}
		return a + b, nil
	}

	funcMap["plus"] = func(x, y interface{}) (int, error) {
		a, err := convertToInt(x)
		if err != nil {
			return 0, err
		}
		b, err := convertToInt(y)
		if err != nil {
			return 0, err
		}
		return a - b, nil
	}
	funcMap["multiply"] = func(x, y interface{}) (int, error) {
		a, err := convertToInt(x)
		if err != nil {
			return 0, err
		}
		b, err := convertToInt(y)
		if err != nil {
			return 0, err
		}
		return a * b, nil
	}
	funcMap["divide"] = func(x, y interface{}) (int, error) {
		a, err := convertToInt(x)
		if err != nil {
			return 0, err
		}
		b, err := convertToInt(y)
		if err != nil {
			return 0, err
		}
		return a / b, nil
	}
	funcMap["mod"] = func(x, y interface{}) (int, error) {
		a, err := convertToInt(x)
		if err != nil {
			return 0, err
		}
		b, err := convertToInt(y)
		if err != nil {
			return 0, err
		}
		return a % b, nil
	}
}

func NewTemplateValueRender(t string) *TemplateValueRender {
	tmpl, err := template.New(t).Funcs(funcMap).Parse(t)
	if err != nil {
		glog.Fatalf("could not parse template %s:%s", t, err)
	}
	return &TemplateValueRender{
		tmpl: tmpl,
	}
}

func (r *TemplateValueRender) Render(event map[string]interface{}) interface{} {
	return nil
}
