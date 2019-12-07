package value_render

// used for ES indexname template

import (
	"encoding/json"
	"errors"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/golang/glog"
)

func dateFormat(t interface{}, format string, location *time.Location) (string, error) {
	if t1, ok := t.(time.Time); ok {
		return t1.In(location).Format(format), nil
	}
	if reflect.TypeOf(t).String() == "json.Number" {
		t1, err := t.(json.Number).Int64()
		if err != nil {
			return format, err
		}
		return time.Unix(t1/1000, t1%1000*1000000).In(location).Format(format), nil
	}
	if reflect.TypeOf(t).Kind() == reflect.Int {
		t1 := int64(t.(int))
		return time.Unix(t1/1000, t1%1000*1000000).In(location).Format(format), nil
	}
	if reflect.TypeOf(t).Kind() == reflect.Int64 {
		t1 := t.(int64)
		return time.Unix(t1/1000, t1%1000*1000000).In(location).Format(format), nil
	}
	if reflect.TypeOf(t).Kind() == reflect.String {
		t1, e := time.Parse(time.RFC3339, t.(string))
		if e != nil {
			return format, e
		}
		return t1.In(location).Format(format), nil
	}
	return format, errors.New("could not tell the type timestamp field belongs to")
}

type field struct {
	literal bool
	date    bool
	value   string
}

type IndexRender struct {
	fields   []*field
	location *time.Location
}

func NewIndexRender(t string) *IndexRender {
	r, _ := regexp.Compile(`%{.*?}`)
	fields := make([]*field, 0)
	lastPos := 0
	for _, loc := range r.FindAllStringIndex(t, -1) {
		s, e := loc[0], loc[1]
		fields = append(fields, &field{
			literal: true,
			value:   t[lastPos:s],
		})

		if t[s+2] == '+' {
			fields = append(fields, &field{
				literal: false,
				date:    true,
				value:   t[s+3 : e-1],
			})
		} else {
			fields = append(fields, &field{
				literal: false,
				date:    false,
				value:   t[s+2 : e-1],
			})
		}

		lastPos = e
	}

	if lastPos < len(t) {
		fields = append(fields, &field{
			literal: true,
			value:   t[lastPos:len(t)],
		})
	}
	return &IndexRender{fields, time.UTC}
}

// SetTimeLocation parse `location` to time.Location ans set it as its member.
// use this location to format time string
func (r *IndexRender) SetTimeLocation(loc string) {
	location, err := time.LoadLocation(loc)
	if err != nil {
		glog.Fatalf("invalid localtion: %s", loc)
	}
	r.location = location
}

func (r *IndexRender) Render(event map[string]interface{}) interface{} {
	fields := make([]string, len(r.fields))
	for i, f := range r.fields {
		if f.literal {
			fields[i] = f.value
			continue
		}

		if f.date {
			if t, ok := event["@timestamp"]; ok {
				fields[i], _ = dateFormat(t, f.value, r.location)
			} else {
				fields[i], _ = dateFormat(time.Now(), f.value, r.location)
			}
		} else {
			if s, ok := event[f.value]; !ok {
				fields[i] = "null"
			} else {
				if fields[i], ok = s.(string); !ok {
					fields[i] = "null"
				}
			}
		}
	}
	return strings.Join(fields, "")
}
