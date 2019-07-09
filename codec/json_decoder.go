package codec

import (
	"bytes"
	"time"
)

type JsonDecoder struct {
	useNumber bool
}

func (jd *JsonDecoder) Decode(value []byte) map[string]interface{} {
	rst := make(map[string]interface{})
	rst["@timestamp"] = time.Now()
	d := json.NewDecoder(bytes.NewReader(value))

	if jd.useNumber {
		d.UseNumber()
	}
	err := d.Decode(&rst)
	if err != nil || d.More() {
		return map[string]interface{}{
			"@timestamp": time.Now(),
			"message":    string(value),
		}
	}

	return rst
}
