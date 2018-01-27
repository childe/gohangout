package codec

import "time"

type PlainDecoder struct {
}

func (d *PlainDecoder) Decode(s string) map[string]interface{} {
	rst := make(map[string]interface{})
	rst["@timestamp"] = time.Now().UTC()
	rst["message"] = s
	return rst
}
