package filter

import (
	"encoding/json"
	"testing"
)

func TestIntConverter(t *testing.T) {
	type testCase struct {
		v    interface{}
		want interface{}
		err  bool
	}

	convert := &IntConverter{}

	cases := []testCase{
		{
			json.Number("1"), int64(1), false,
		},
		{
			"1", int64(1), false,
		},
		{
			1, int64(1), false,
		},
		{
			-1, int64(-1), false,
		},
		{
			"-1", int64(-1), false,
		},
		{
			"12345678901234567890", int64(0), true,
		},
	}

	for _, c := range cases {
		ans, err := convert.convert(c.v)
		if ans != c.want {
			t.Errorf("convert %v: want %v, got %v", c.v, c.want, ans)
		}

		if c.err != (err != nil) {
			t.Errorf("convert %v: want %v, got %v", c.v, c.err, err)
		}
	}
}

func TestUIntConverter(t *testing.T) {
	type testCase struct {
		v    interface{}
		want interface{}
		err  bool
	}

	convert := &UIntConverter{}

	cases := []testCase{
		{
			json.Number("1"), uint64(1), false,
		},
		{
			"1", uint64(1), false,
		},
		{
			1, uint64(1), false,
		},
		{
			-1, uint64(0), true,
		},
		{
			"-1", uint64(0), true,
		},
		{
			"12345678901234567890", uint64(12345678901234567890), false,
		},
	}

	for _, c := range cases {
		ans, err := convert.convert(c.v)
		if ans != c.want {
			t.Errorf("convert %v: want %v, got %v", c.v, c.want, ans)
		}

		if c.err != (err != nil) {
			t.Errorf("convert %v: want %v, got %v", c.v, c.err, err)
		}
	}
}

func TestFloatConverter(t *testing.T) {
	type testCase struct {
		v    interface{}
		want interface{}
		err  bool
	}

	convert := &FloatConverter{}

	cases := []testCase{
		{
			json.Number("1.1"), float64(1.1), false,
		},
		{
			"1.2", float64(1.2), false,
		},
		{
			1.3, float64(1.3), false,
		},
		{
			-1.4, float64(-1.4), false,
		},
		{
			"-1.5", float64(-1.5), false,
		},
		{
			"abcd", 0.0, true,
		},
		{
			"", 0.0, true,
		},
	}

	for _, c := range cases {
		ans, err := convert.convert(c.v)
		if ans != c.want {
			t.Errorf("convert %v: want %v, got %v", c.v, c.want, ans)
		}

		if c.err != (err != nil) {
			t.Errorf("convert %v: want %v, got %v", c.v, c.err, err)
		}
	}
}

func TestBoolConverter(t *testing.T) {
	type testCase struct {
		v    interface{}
		want interface{}
		err  bool
	}

	convert := &BoolConverter{}

	cases := []testCase{
		{
			"abcd", nil, true,
		},
		{
			"True", true, false,
		},
		{
			"false", false, false,
		},
		{
			json.Number("1"), nil, true,
		},
		{
			1234, nil, true,
		},
	}

	for _, c := range cases {
		ans, err := convert.convert(c.v)
		if ans != c.want {
			t.Errorf("convert %v: want %v, got %v", c.v, c.want, ans)
		}

		if c.err != (err != nil) {
			t.Errorf("convert %v: want %v, got %v", c.v, c.err, err)
		}
	}
}

func TestSettoIfNil(t *testing.T) {
	config := make(map[interface{}]interface{})
	fields := make(map[interface{}]interface{})
	fields["timeTaken"] = map[interface{}]interface{}{
		"to":           "float",
		"setto_if_nil": 0.0,
	}
	config["fields"] = fields
	f := BuildFilter("Convert", config)
	event := map[string]interface{}{}

	event, ok := f.Filter(event)
	t.Log(event)

	if ok == false {
		t.Error("ConvertFilter fail")
	}

	if event["timeTaken"].(float64) != 0.0 {
		t.Error("timeTaken convert error")
	}
}

func TestConvertFilter(t *testing.T) {
	config := make(map[interface{}]interface{})
	fields := make(map[interface{}]interface{})
	fields["id"] = map[interface{}]interface{}{
		"to":            "uint",
		"setto_if_fail": 0,
	}
	fields["responseSize"] = map[interface{}]interface{}{
		"to":            "int",
		"setto_if_fail": 0,
	}
	fields["timeTaken"] = map[interface{}]interface{}{
		"to":             "float",
		"remove_if_fail": true,
	}
	// add to string test case
	fields["toString"] = map[interface{}]interface{}{
		"to":             "string",
		"remove_if_fail": true,
	}
	config["fields"] = fields
	f := BuildFilter("Convert", config)

	case1 := map[string]int{"a": 5, "b": 7}
	event := map[string]interface{}{
		"id":           "12345678901234567890",
		"responseSize": "10",
		"timeTaken":    "0.010",
		"toString":     case1,
	}
	t.Log(event)

	event, ok := f.Filter(event)
	t.Log(event)

	if ok == false {
		t.Error("ConvertFilter fail")
	}

	if event["id"].(uint64) != 12345678901234567890 {
		t.Error("id should be 12345678901234567890")
	}
	if event["responseSize"].(int64) != 10 {
		t.Error("responseSize should be 10")
	}
	if event["timeTaken"].(float64) != 0.01 {
		t.Error("timeTaken should be 0.01")
	}
	if event["toString"].(string) != "{\"a\":5,\"b\":7}" {
		t.Error("toString is unexpected")
	}
	event = map[string]interface{}{
		"responseSize": "10.1",
		"timeTaken":    "abcd",
		"toString":     "huangjacky",
	}
	t.Log(event)

	event, ok = f.Filter(event)
	t.Log(event)

	if ok == false {
		t.Error("ConvertFilter fail")
	}

	if event["responseSize"].(int) != 0 {
		t.Error("responseSize should be 0")
	}
	if event["timeTaken"] != nil {
		t.Error("timeTaken should be nil")
	}
	if event["toString"].(string) != "huangjacky" {
		t.Error("toString should be huangjacky")
	}
}
