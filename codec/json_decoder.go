package codec

import (
	"bytes"
	"time"

	"github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type JsonDecoder struct {
}

func (jd *JsonDecoder) Decode(value []byte) map[string]interface{} {
	rst := make(map[string]interface{})
	rst["@timestamp"] = time.Now()
	d := json.NewDecoder(bytes.NewReader(value))
	d.UseNumber()
	err := d.Decode(&rst)
	if err != nil {
		return map[string]interface{}{
			"@timestamp": time.Now(),
			"message":    string(value),
		}
	}
	return rst
}
