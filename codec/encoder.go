package codec

import (
	"plugin"
	"strings"

	"github.com/childe/gohangout/simplejson"
	"k8s.io/klog/v2"
)

type Encoder interface {
	Encode(any) ([]byte, error)
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
			klog.Fatalf("format of `%s` is incorrect", t)
		}
		format := splited[1]
		return NewFormatEncoder(format)
	}

	// try plugin
	p, err := plugin.Open(t)
	if err != nil {
		klog.Fatalf("could not open %s: %s", t, err)
	}
	newFunc, err := p.Lookup("New")
	if err != nil {
		klog.Fatalf("could not find New function in %s: %s", t, err)
	}
	return newFunc.(func() any)().(Encoder)
}
