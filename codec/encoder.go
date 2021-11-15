package codec

import (
	"plugin"
	"strings"

	"github.com/childe/gohangout/simplejson"
	"github.com/golang/glog"
)

type Encoder interface {
	Encode(interface{}) ([]byte, error)
}

func NewEncoder(t string) Encoder {
	switch t {
	case "json":
		return &JsonEncoder{}
	case "simplejson":
		return &simplejson.SimpleJsonDecoder{}
	}

	// FormatEncoder
	if strings.HasPrefix(t, "format:") {
		splited := strings.SplitN(t, ":", 2)
		if len(splited) != 2 {
			glog.Fatalf("format of `%s` is incorrect", t)
		}
		format := splited[1]
		return NewFormatEncoder(format)
	}

	// try plugin
	p, err := plugin.Open(t)
	if err != nil {
		glog.Fatalf("could not open %s: %s", t, err)
	}
	newFunc, err := p.Lookup("New")
	if err != nil {
		glog.Fatalf("could not find New function in %s: %s", t, err)
	}
	return newFunc.(func() interface{})().(Encoder)
}
