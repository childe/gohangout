package codec

import (
	"encoding/json"
	"strings"
	"time"
)

type JsonDecoder struct {
}

func (jd *JsonDecoder) Decode(s string) map[string]interface{} {
	rst := make(map[string]interface{})
	rst["@timestamp"] = time.Now().UnixNano() / 1000000
	d := json.NewDecoder(strings.NewReader(s))
	d.UseNumber()
	err := d.Decode(&rst)
	if err != nil {
		rst["message"] = s
	}
	return rst
}
