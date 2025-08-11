package codec

import (
	"plugin"

	"k8s.io/klog/v2"
)

type Decoder interface {
	Decode([]byte) map[string]any
}

func NewDecoder(t string) Decoder {
	switch t {
	case "plain":
		return &PlainDecoder{}
	case "json":
		return &JsonDecoder{useNumber: true}
	case "json:not_usenumber":
		return &JsonDecoder{useNumber: false}
	default:
		p, err := plugin.Open(t)
		if err != nil {
			klog.Fatalf("could not open %s: %s", t, err)
		}
		newFunc, err := p.Lookup("New")
		if err != nil {
			klog.Fatalf("could not find New function in %s: %s", t, err)
		}
		return newFunc.(func() any)().(Decoder)
	}
}
