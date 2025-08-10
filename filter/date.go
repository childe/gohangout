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
)

type DateParser interface {
	Parse(any) (time.Time, error)
}

type FormatParser struct {
	format   string
	location *time.Location
	addYear  bool
}

var MustStringTypeError = errors.New("timestamp field must be string")

func (dp *FormatParser) Parse(t any) (time.Time, error) {
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

func (p *UnixParser) Parse(t any) (time.Time, error) {
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

func (p *UnixMSParser) Parse(t any) (time.Time, error) {
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

func (p *ISO8601Parser) Parse(t any) (time.Time, error) {
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

// DateConfig defines the configuration structure for Date filter
type DateConfig struct {
	Src       string   `json:"src"`
	Target    string   `json:"target"`
	Location  string   `json:"location"`
	AddYear   bool     `json:"add_year"`
	Overwrite bool     `json:"overwrite"`
	Formats   []string `json:"formats"`
}

type DateFilter struct {
	config      map[any]any
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

func newDateFilter(config map[any]any) topology.Filter {
	plugin := &DateFilter{
		config:      config,
		dateParsers: make([]DateParser, 0),
	}

	// Parse configuration using SafeDecodeConfig
	var dateConfig DateConfig
	// Set default values
	dateConfig.Target = "@timestamp"
	dateConfig.Overwrite = true
	dateConfig.AddYear = false

	SafeDecodeConfig("Date", config, &dateConfig)

	// Validate required fields
	if dateConfig.Src == "" {
		panic("Date filter: 'src' is required")
	}
	if dateConfig.Formats == nil || len(dateConfig.Formats) == 0 {
		panic("Date filter: 'formats' is required and cannot be empty")
	}

	plugin.overwrite = dateConfig.Overwrite
	plugin.src = dateConfig.Src
	plugin.srcVR = value_render.GetValueRender2(plugin.src)
	plugin.target = dateConfig.Target
	plugin.targetFS = field_setter.NewFieldSetter(plugin.target)

	// Parse location
	var location *time.Location
	if dateConfig.Location != "" {
		var err error
		location, err = time.LoadLocation(dateConfig.Location)
		if err != nil {
			panic(fmt.Sprintf("Date filter: load location error: %s", err))
		}
	}

	// Create date parsers
	for _, format := range dateConfig.Formats {
		plugin.dateParsers = append(plugin.dateParsers, getDateParser(format, location, dateConfig.AddYear))
	}

	return plugin
}

func (plugin *DateFilter) Filter(event map[string]any) (map[string]any, bool) {
	inputI, err := plugin.srcVR.Render(event)
	if err != nil || inputI == nil {
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
