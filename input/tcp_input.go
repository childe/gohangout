package input

import (
	"bufio"
	"net"

	"github.com/childe/gohangout/codec"
	"github.com/childe/gohangout/topology"
	"k8s.io/klog/v2"
)

type TCPInput struct {
	config  map[interface{}]interface{}
	network string
	address string

	decoder codec.Decoder

	l        net.Listener
	messages chan []byte
	stop     bool
}

func readLine(scanner *bufio.Scanner, c net.Conn, messages chan<- []byte) {
	for scanner.Scan() {
		t := scanner.Bytes()
		buf := make([]byte, len(t))
		copy(buf, t)
		messages <- buf
	}

	if err := scanner.Err(); err != nil {
		klog.Errorf("read from %v->%v error: %v", c.RemoteAddr(), c.LocalAddr(), err)
	}
	c.Close()
}

func init() {
	Register("TCP", newTCPInput)
}
func newTCPInput(config map[interface{}]interface{}) topology.Input {
	var codertype string = "plain"
	if v, ok := config["codec"]; ok {
		codertype = v.(string)
	}

	p := &TCPInput{
		config:   config,
		decoder:  codec.NewDecoder(codertype),
		messages: make(chan []byte, 10),
	}

	if v, ok := config["max_length"]; ok {
		if max, ok := v.(int); ok {
			if max <= 0 {
				klog.Fatal("max_length must be bigger than zero")
			}
		} else {
			klog.Fatal("max_length must be int")
		}
	}

	p.network = "tcp"
	if network, ok := config["network"]; ok {
		p.network = network.(string)
	}

	if addr, ok := config["address"]; ok {
		p.address = addr.(string)
	} else {
		klog.Fatal("address must be set in TCP input")
	}

	l, err := net.Listen(p.network, p.address)
	if err != nil {
		klog.Fatal(err)
	}
	p.l = l

	go func() {
		for !p.stop {
			conn, err := l.Accept()
			if err != nil {
				if p.stop {
					return
				}
				klog.Error(err)
			} else {
				scanner := bufio.NewScanner(conn)
				if v, ok := config["max_length"]; ok {
					max := v.(int)
					scanner.Buffer(make([]byte, 0, max), max)
				}
				go readLine(scanner, conn, p.messages)
			}
		}
	}()
	return p
}

func (p *TCPInput) ReadOneEvent() map[string]interface{} {
	text, more := <-p.messages
	if !more || text == nil {
		return nil
	}
	return p.decoder.Decode(text)
}

func (p *TCPInput) Pause() {}

func (p *TCPInput) Shutdown() {
	p.stop = true
	p.l.Close()
	close(p.messages)
}
