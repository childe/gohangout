package output

import (
	"bufio"
	"net"

	"github.com/childe/gohangout/simplejson"
	"github.com/golang/glog"
)

type TCPOutput struct {
	BaseOutput
	config  map[interface{}]interface{}
	address string

	writer *bufio.Writer
	conn   net.Conn
}

func NewTCPOutput(config map[interface{}]interface{}) *TCPOutput {
	p := &TCPOutput{
		BaseOutput: NewBaseOutput(config),
		config:     config,
	}

	var address string
	if addr, ok := config["address"]; ok {
		address, ok = addr.(string)
		if !ok {
			glog.Fatal("address must be string")
		}
	} else {
		glog.Fatal("address must be set in TCP output")
	}
	p.address = address

	conn, err := net.Dial("tcp", address)
	if err != nil {
		glog.Fatal(err)
	}
	p.conn = conn
	p.writer = bufio.NewWriter(conn)

	return p
}

func (p *TCPOutput) Emit(event map[string]interface{}) {
	d := &simplejson.SimpleJsonDecoder{}
	buf, err := d.Encode(event)
	if err != nil {
		glog.Errorf("marshal %v error:%s", event, err)
	}

	buf = append(buf, '\n')
	n, err := p.writer.Write(buf)
	if n != len(buf) {
		glog.Errorf("write to %s error: %s", p.address, err)
	}
	p.writer.Flush()
}

func (p *TCPOutput) Shutdown() {
	p.writer.Flush()
	p.conn.Close()
}
