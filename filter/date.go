package filter

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"time"

	"github.com/childe/gohangout/field_setter"
	"github.com/childe/gohangout/topology"
	"github.com/childe/gohangout/value_render"
	"github.com/relvacode/iso8601" // https://pkg.go.dev/github.com/relvacode/iso8601#section-readme
	"k8s.io/klog/v2"
)

type DateParser interface {
	Parse(interface{}) (time.Time, error)
}

type FormatParser struct {
	format   string
	location *time.Location
	addYear  bool
}

var MustStringTypeError = errors.New("timestamp field must be string")

func (dp *FormatParser) Parse(t interface{}) (time.Time, error) {
	var (
		rst time.Time
		err error
	)
	value, ok := t.(string)

	if !ok {
		return rst, MustStringTypeError
	}

	if dp.addYear {
		value = fmt.Sprintf("%d%s", time.Now().Year(), value)
	}
	if dp.location == nil {
		return time.Parse(dp.format, value)
	}
	rst, err = time.ParseInLocation(dp.format, value, dp.location)
	if err != nil {
		return rst, err
	}
	return rst.UTC(), nil
}

type UnixParser struct{}

func (p *UnixParser) Parse(t interface{}) (time.Time, error) {
	var (
		rst time.Time
	)
	if v, ok := t.(json.Number); ok {
		t1, err := v.Int64()
		if err != nil {
			return rst, err
		}
		return time.Unix(t1, 0), nil
	}

	if v, ok := t.(string); ok {
		t1, err := strconv.Atoi(v)
		if err != nil {
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return rst, err
			}
			t1 := math.Floor(f)
			return time.Unix(int64(t1), int64(1000000000*(f-t1))), nil
		}
		return time.Unix(int64(t1), 0), nil
	}

	if t1, ok := t.(int); ok {
		return time.Unix(int64(t1), 0), nil
	}
	if t1, ok := t.(int64); ok {
		return time.Unix(t1, 0), nil
	}
	return rst, fmt.Errorf("%s unknown type:%s", t, reflect.TypeOf(t).String())
}

type UnixMSParser struct{}

func (p *UnixMSParser) Parse(t interface{}) (time.Time, error) {
	var (
		rst time.Time
	)
	if v, ok := t.(json.Number); ok {
		t1, err := v.Int64()
		if err != nil {
			return rst, err
		}
		return time.Unix(t1/1000, t1%1000*1000000), nil
	}
	if v, ok := t.(string); ok {
		t1, err := strconv.Atoi(v)
		if err != nil {
			return rst, err
		}
		t2 := int64(t1)
		return time.Unix(t2/1000, t2%1000*1000000), nil
	}
	if v, ok := t.(int); ok {
		t1 := int64(v)
		return time.Unix(t1/1000, t1%1000*1000000), nil
	}
	if v, ok := t.(int64); ok {
		return time.Unix(v/1000, v%1000*1000000), nil
	}
	return rst, fmt.Errorf("%s unknown type:%s", t, reflect.TypeOf(t).String())
}

type ISO8601Parser struct {
	location *time.Location // If the input does not have timezone information, it will use the given location.
}

func (p *ISO8601Parser) Parse(t interface{}) (time.Time, error) {
	var (
		rst time.Time
	)
	if v, ok := t.(string); ok {
		if p.location == nil {
			return iso8601.ParseString(v)
		}
		return iso8601.ParseStringInLocation(v, p.location)
	}

	return rst, fmt.Errorf("%s unknown type:%s", t, reflect.TypeOf(t).String())
}

func getDateParser(format string, l *time.Location, addYear bool) DateParser {
	if format == "UNIX" {
		return &UnixParser{}
	}
	if format == "UNIX_MS" {
		return &UnixMSParser{}
	}
	if format == "RFC3339" {
		return &FormatParser{time.RFC3339, l, addYear}
	}
	if format == "ISO8601" {
		return &ISO8601Parser{l}
	}
	return &FormatParser{format, l, addYear}
}

type DateFilter struct {
	config      map[interface{}]interface{}
	dateParsers []DateParser
	overwrite   bool
	src         string
	srcVR       value_render.ValueRender
	target      string
	targetFS    field_setter.FieldSetter
}

func init() {
	Register("Date", newDateFilter)
}

func newDateFilter(config map[interface{}]interface{}) topology.Filter {
	plugin := &DateFilter{
		config:      config,
		overwrite:   true,
		dateParsers: make([]DateParser, 0),
	}

	if overwrite, ok := config["overwrite"]; ok {
		plugin.overwrite = overwrite.(bool)
	}

	if srcValue, ok := config["src"]; ok {
		plugin.src = srcValue.(string)
	} else {
		klog.Fatal("src must be set in date filter plugin")
	}
	plugin.srcVR = value_render.GetValueRender2(plugin.src)

	if targetI, ok := config["target"]; ok {
		plugin.target = targetI.(string)
	} else {
		plugin.target = "@timestamp"
	}
	plugin.targetFS = field_setter.NewFieldSetter(plugin.target)

	var (
		location *time.Location
		addYear  bool = false
		err      error
	)
	if locationI, ok := config["location"]; ok {
		location, err = time.LoadLocation(locationI.(string))
		if err != nil {
			klog.Fatalf("load location error:%s", err)
		}
	} else {
		location = nil
	}
	if addYearI, ok := config["add_year"]; ok {
		addYear = addYearI.(bool)
	}
	if formats, ok := config["formats"]; ok {
		for _, formatI := range formats.([]interface{}) {
			plugin.dateParsers = append(plugin.dateParsers, getDateParser(formatI.(string), location, addYear))
		}
	} else {
		klog.Fatal("formats must be set in date filter plugin")
	}

	return plugin
}

func (plugin *DateFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	inputI := plugin.srcVR.Render(event)
	if inputI == nil {
		return event, false
	}

	for _, dp := range plugin.dateParsers {
		t, err := dp.Parse(inputI)
		if err == nil {
			event = plugin.targetFS.SetField(event, t, "", plugin.overwrite)
			return event, true
		}
	}
	return event, false
}
