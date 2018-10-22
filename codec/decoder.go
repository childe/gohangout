package codec

import "github.com/golang/glog"

type Decoder interface {
	Decode([]byte) map[string]interface{}
}

func NewDecoder(t string) Decoder {
	if t == "json" {
		return &JsonDecoder{}
	}
	if t == "plain" {
		return &PlainDecoder{}
	}
	if t == "hermes2" {
		return &HermesDecoder2{
			MAGIC:      []byte("hems"),
			CRC_LENGTH: 8,
		}
	}
	if t == "hermes" {
		return &HermesDecoder{
			MAGIC:      []byte("hems"),
			CRC_LENGTH: 8,
		}
	}
	glog.Infof("no %s decoder, use plain decoder", t)
	return &PlainDecoder{}
}
