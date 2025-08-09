package codec

import "time"

type PlainDecoder struct {
}

func (d *PlainDecoder) Decode(value []byte) map[string]any {
	rst := make(map[string]any)
	rst["@timestamp"] = time.Now()
	rst["message"] = string(value)
	return rst
}
