package codec

import (
	"reflect"
	"testing"

	"github.com/childe/gohangout/simplejson"
)

func TestNewEncoder(t *testing.T) {
	cases := []struct {
		codec       string
		encoderType Encoder
	}{
		{
			codec:       "json",
			encoderType: &JsonEncoder{},
		},
		{
			codec:       "simplejson",
			encoderType: &simplejson.SimpleJsonDecoder{},
		},
		{
			codec:       "format:[msg]",
			encoderType: &FormatEncoder{},
		},
	}

	for _, c := range cases {
		t.Logf("test %v", c.codec)
		encoder := NewEncoder(c.codec)
		got := reflect.TypeOf(encoder).String()
		expectedEncoderType := reflect.TypeOf(c.encoderType).String()
		if got != expectedEncoderType {
			t.Errorf("expected `%s`, got `%s`", expectedEncoderType, got)
		}
	}
}
