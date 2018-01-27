package codec

import (
	"strings"
	"time"

	"github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type JsonDecoder struct {
}

func (jd *JsonDecoder) Decode(s string) map[string]interface{} {
	rst := make(map[string]interface{})
	rst["@timestamp"] = time.Now().UTC()
	d := json.NewDecoder(strings.NewReader(s))
	d.UseNumber()
	err := d.Decode(&rst)
	if err != nil {
		rst["message"] = s
	}
	return rst
}
