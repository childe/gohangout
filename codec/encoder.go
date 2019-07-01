package codec

import (
	"github.com/childe/gohangout/simplejson"
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
	panic(t + " encoder not supported")
	return nil
}
