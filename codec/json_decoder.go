package codec

import (
	"bytes"
	stdJson "encoding/json"
	"time"
)

type JsonDecoder struct {
}

func (jd *JsonDecoder) Decode(value []byte) map[string]interface{} {
	rst := make(map[string]interface{})
	rst["@timestamp"] = time.Now()
	d := json.NewDecoder(bytes.NewReader(value))
	d.UseNumber()
	err := d.Decode(&rst)
	if err != nil || d.More() {
		return map[string]interface{}{
			"@timestamp": time.Now(),
			"message":    string(value),
		}
	}

	convertNumberType(rst)
	return rst
}

func convertNumberType(rst map[string]interface{}) {
	for k, v := range rst {
		if nv, ok := v.(stdJson.Number); ok {
			if rv, err := nv.Int64(); err == nil {
				rst[k] = rv
			} else if rv, err := nv.Float64(); err == nil {
				rst[k] = rv
			} else {
				rst[k] = nv.String
			}
		}
	}
}
