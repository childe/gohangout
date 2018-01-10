package codec

import "github.com/golang/glog"

type Decoder interface {
	Decode(string) map[string]interface{}
}

func NewDecoder(t string) Decoder {
	if t == "json" {
		return &JsonDecoder{}
	}
	if t == "plain" {
		return &PlainDecoder{}
	}
	glog.Infof("no %s decoder, use plain decoder", t)
	return &PlainDecoder{}
}
