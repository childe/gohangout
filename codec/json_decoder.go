package codec

import (
	"bytes"
	"time"
)

type JsonDecoder struct {
	useNumber bool
}

func (jd *JsonDecoder) Decode(value []byte) map[string]any {
	rst := make(map[string]any)
	rst["@timestamp"] = time.Now()
	d := json.NewDecoder(bytes.NewReader(value))

	if jd.useNumber {
		d.UseNumber()
	}
	err := d.Decode(&rst)
	if err != nil || d.More() {
		return map[string]any{
			"@timestamp": time.Now(),
			"message":    string(value),
		}
	}

	return rst
}
