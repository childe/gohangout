package codec

import (
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
	default:
		glog.Infof("no %s encoder, use json decoder", t)
		return &JsonEncoder{}
	}
}
