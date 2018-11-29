package codec

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"encoding/binary"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/golang/glog"
)

type HermesDecoder2 struct {
	MAGIC      []byte
	CRC_LENGTH int
}

func (hd *HermesDecoder2) Decode(value []byte) map[string]interface{} {
	offset := 0
	offset += 4

	//version := int(value[offset]) // version
	offset++

	//binary.BigEndian.Uint32(value[offset:]) // totalLength
	offset += 4

	headerLength := int(binary.BigEndian.Uint32(value[offset:]))
	offset += 4

	//bodyLength := binary.BigEndian.Uint32(value[offset:])
	offset += 4

	codecType := getCodecType(value[offset : offset+headerLength])

	offset += headerLength

	value = value[offset : len(value)-hd.CRC_LENGTH]

	codeAndCompress := strings.SplitN(codecType, ",", 2)
	if len(codeAndCompress) == 2 {
		if codeAndCompress[1] == "gzip" {
			reader, err := gzip.NewReader(bytes.NewReader(value))
			if err != nil {
				glog.Errorf("gzip decode hermes message error:%s", err)
				return nil
			}
			if value, err = ioutil.ReadAll(reader); err != nil && err != io.EOF {
				glog.Errorf("gzip decode hermes message error:%s", err)
				return nil
			}
		} else if strings.Contains(codeAndCompress[1], "deflater") {
			reader := flate.NewReader(bytes.NewReader(value))
			var err error
			if value, err = ioutil.ReadAll(reader); err != nil && err != io.EOF {
				glog.Errorf("gzip decode hermes message error:%s", err)
				return nil
			}
		} else {
			glog.Fatalf("%s unknown codec type", codecType)
		}
	}

	rst := map[string]interface{}{"@timestamp": time.Now()}

	// value is created by json.dumps(json.dumps(JosnEvent)). OMG
	var ss string
	err := json.Unmarshal(value, &ss)
	if err != nil {
		rst["message"] = string(value)
	} else {
		d := json.NewDecoder(strings.NewReader(ss))
		d.UseNumber()
		err := d.Decode(&rst)
		if err != nil {
			rst["message"] = string(value)
		}
	}
	return rst
}
