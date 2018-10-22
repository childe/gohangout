package codec

import "time"

type PlainDecoder struct {
}

func (d *PlainDecoder) Decode(value []byte) map[string]interface{} {
	rst := make(map[string]interface{})
	rst["@timestamp"] = time.Now()
	rst["message"] = string(value)
	return rst
}
