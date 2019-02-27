package output

import (
	"net"
	"strings"
	"sync"
	"time"

	"github.com/childe/gohangout/simplejson"
	"github.com/golang/glog"
)

type TCPOutput struct {
	BaseOutput
	config    map[interface{}]interface{}
	network   string
	address   string
	timeout   time.Duration
	keepalive time.Duration

	//writer *bufio.Writer
	conn net.Conn

	dialLock sync.Mutex
}

func NewTCPOutput(config map[interface{}]interface{}) *TCPOutput {
	p := &TCPOutput{
		BaseOutput: NewBaseOutput(config),
		config:     config,
	}

	p.network = "tcp"
	if network, ok := config["network"]; ok {
		p.network = network.(string)
	}

	if addr, ok := config["address"]; ok {
		p.address, ok = addr.(string)
	} else {
		glog.Fatal("address must be set in TCP output")
	}

	if timeoutI, ok := config["dial.timeout"]; ok {
		timeout := timeoutI.(int)
		p.timeout = time.Second * time.Duration(timeout)
	}

	if keepaliveI, ok := config["keepalive"]; ok {
		keepalive, ok := keepaliveI.(int)
		if !ok {
			glog.Fatal("keepalive must be integer")
		}
		p.keepalive = time.Second * time.Duration(keepalive)
	}

	p.loopDial()

	return p
}

func (p *TCPOutput) loopDial() {
	p.dialLock.Lock()
	defer p.dialLock.Unlock()
	for {
		if err := p.dial(); err != nil {
			glog.Errorf("dial error: %s. sleep 10s", err)
			time.Sleep(10 * time.Second)
		} else {
			glog.Infof("conn built to %s", p.conn.RemoteAddr())
			return
		}
	}
}

func (p *TCPOutput) dial() error {
	var d net.Dialer
	d.Timeout = p.timeout
	d.KeepAlive = p.keepalive

	conn, err := net.Dial(p.network, p.address)
	if err != nil {
		return err
	}
	p.conn = conn
	//p.writer = bufio.NewWriter(conn)

	return nil
}
func (p *TCPOutput) write(buf []byte) error {
	for len(buf) > 0 {
		n, err := p.conn.Write(buf)
		if err != nil {
			glog.Errorf("write to %s[%s] error: %s", p.address, p.conn.RemoteAddr(), err)
			str := err.Error()
			switch {
			case strings.Contains(str, "use of closed network connection"):
				p.loopDial()
				return err
			case strings.Contains(str, "write: broken pipe"):
				p.conn.Close()
				p.loopDial()
				return err
			}
		}
		buf = buf[n:]
	}
	return nil
}

func (p *TCPOutput) Emit(event map[string]interface{}) {
	d := &simplejson.SimpleJsonDecoder{}
	buf, err := d.Encode(event)
	if err != nil {
		glog.Errorf("marshal %v error:%s", event, err)
		return
	}

	p.write(buf) // always write \n, no matter if error occures here

	buf = []byte{'\n'}
	if err := p.write(buf); err != nil {
		return
	}

	//buf = append(buf, '\n')
	//n, err := p.writer.Write(buf)
	//if n != len(buf) {
	//glog.Errorf("write to %s[%s] error: %s", p.address, p.conn.RemoteAddr(), err)
	//}
	//p.writer.Flush()
}

func (p *TCPOutput) Shutdown() {
	//p.writer.Flush()
	p.conn.Close()
}
