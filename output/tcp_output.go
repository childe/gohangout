package output

import (
	"bufio"
	"net"
	"time"

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

	var d net.Dialer

	if timeoutI, ok := config["dial.timeout"]; ok {
		timeout, ok := timeoutI.(int)
		if !ok {
			glog.Fatal("dial.timeout must be integer")
		}
		d.Timeout = time.Second * time.Duration(timeout)
	}

	if keepaliveI, ok := config["keepalive"]; ok {
		keepalive, ok := keepaliveI.(int)
		if !ok {
			glog.Fatal("keepalive must be integer")
		}
		d.KeepAlive = time.Second * time.Duration(keepalive)
	}

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

	for len(buf) > 0 {
		n, err := p.conn.Write(buf)
		if err != nil {
			glog.Errorf("write to %s[%s] error: %s", p.address, p.conn.RemoteAddr(), err)
		}
		buf = buf[n:]
	}

	buf = []byte{'\n'}
	for len(buf) > 0 {
		n, err := p.conn.Write(buf)
		if n == 1 {
			break
		}
		if err != nil {
			glog.Errorf("write to %s[%s] error: %s", p.address, p.conn.RemoteAddr(), err)
		}
	}

	//buf = append(buf, '\n')
	//n, err := p.writer.Write(buf)
	//if n != len(buf) {
	//glog.Errorf("write to %s[%s] error: %s", p.address, p.conn.RemoteAddr(), err)
	//}
	//p.writer.Flush()
}

func (p *TCPOutput) Shutdown() {
	p.writer.Flush()
	p.conn.Close()
}
