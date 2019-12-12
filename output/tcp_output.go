package output

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/childe/gohangout/simplejson"
	"github.com/golang/glog"
)

type TCPOutput struct {
	config    map[interface{}]interface{}
	network   string
	address   string
	timeout   time.Duration
	keepalive time.Duration

	concurrent int
	messages   chan map[string]interface{}
	conn       []net.Conn
	//writer *bufio.Writer

	dialLock sync.Mutex
}

func (l *MethodLibrary) NewTCPOutput(config map[interface{}]interface{}) *TCPOutput {
	p := &TCPOutput{
		config:     config,
		concurrent: 1,
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

	if v, ok := config["concurrent"]; ok {
		p.concurrent = v.(int)
	}
	p.messages = make(chan map[string]interface{}, p.concurrent)
	p.conn = make([]net.Conn, p.concurrent)

	for i := 0; i < p.concurrent; i++ {
		go func(i int) {
			p.conn[i] = p.loopDial()
			for {
				event := <-p.messages
				d := &simplejson.SimpleJsonDecoder{}
				buf, err := d.Encode(event)
				if err != nil {
					glog.Errorf("marshal %v error:%s", event, err)
					return
				}

				buf = append(buf, '\n')
				for {
					if err = write(p.conn[i], buf); err != nil {
						glog.Error(err)
						p.conn[i].Close()
						p.conn[i] = p.loopDial()
					} else {
						break
					}
				}
			}
		}(i)
	}

	return p
}

func (p *TCPOutput) loopDial() net.Conn {
	for {
		if conn, err := p.dial(); err != nil {
			glog.Errorf("dial error: %s. sleep 1s", err)
			time.Sleep(1 * time.Second)
		} else {
			glog.Infof("conn built to %s", conn.RemoteAddr())
			return conn
		}
	}
}

func (p *TCPOutput) dial() (net.Conn, error) {
	var d net.Dialer
	d.Timeout = p.timeout
	d.KeepAlive = p.keepalive

	conn, err := net.Dial(p.network, p.address)
	if err != nil {
		return conn, err
	}
	// *TcpConn is net.Conn interface, so we can pass conn instead of &conn
	go probe(conn)
	//p.writer = bufio.NewWriter(conn)

	return conn, nil
}

func probe(conn net.Conn) {
	var b = make([]byte, 1)

	conn.SetDeadline(time.Time{})
	conn.SetReadDeadline(time.Time{})
	_, err := conn.Read(b) // should block here
	if err != nil && err == io.EOF {
		glog.Infof("conn [%s] is closed by the server, close the conn.", conn.RemoteAddr())
		conn.Close()
	}
}

func (p *TCPOutput) Emit(event map[string]interface{}) {
	p.messages <- event
	//buf = append(buf, '\n')
	//n, err := p.writer.Write(buf)
	//if n != len(buf) {
	//glog.Errorf("write to %s[%s] error: %s", p.address, p.conn.RemoteAddr(), err)
	//}
	//p.writer.Flush()
}

func write(conn net.Conn, buf []byte) error {
	for len(buf) > 0 {
		n, err := conn.Write(buf)
		if err != nil {
			return err
			//glog.Errorf("write to %s[%s] error: %s", p.address, conn.RemoteAddr(), err)
			//switch {
			//case strings.Contains(str, "use of closed network connection"):
			//conn = loopDial()
			//return err
			//case strings.Contains(str, "write: broken pipe"):
			//conn.Close()
			//conn = loopDial()
			//return err
			//}
		}
		buf = buf[n:]
	}
	return nil
}

func (p *TCPOutput) Shutdown() {
	//p.writer.Flush()
	//p.conn.Close()
}
