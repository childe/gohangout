package input

import (
	"bufio"
	"net"

	"github.com/childe/gohangout/codec"
	"github.com/golang/glog"
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

func readLine(c net.Conn, messages chan<- []byte) {
	scanner := bufio.NewScanner(c)
	for scanner.Scan() {
		messages <- scanner.Bytes()
	}

	if err := scanner.Err(); err != nil {
		glog.Errorf("read from %s->%s error: %s", c.RemoteAddr(), c.LocalAddr(), err)
	}
	c.Close()
}

func NewTCPInput(config map[interface{}]interface{}) *TCPInput {
	var codertype string = "plain"
	if v, ok := config["codec"]; ok {
		codertype = v.(string)
	}

	p := &TCPInput{
		config:   config,
		decoder:  codec.NewDecoder(codertype),
		messages: make(chan []byte, 10),
	}

	p.network = "tcp"
	if network, ok := config["network"]; ok {
		p.network = network.(string)
	}

	if addr, ok := config["address"]; ok {
		p.address = addr.(string)
	} else {
		glog.Fatal("address must be set in TCP input")
	}

	l, err := net.Listen(p.network, p.address)
	if err != nil {
		glog.Fatal(err)
	}
	p.l = l

	go func() {
		for !p.stop {
			conn, err := l.Accept()
			if err != nil {
				if p.stop {
					return
				}
				glog.Error(err)
			} else {
				go readLine(conn, p.messages)
			}
		}
	}()
	return p
}

func (p *TCPInput) readOneEvent() map[string]interface{} {
	text := <-p.messages
	if text == nil {
		return nil
	}
	return p.decoder.Decode(text)
}

func (p *TCPInput) Shutdown() {
	p.stop = true
	p.l.Close()
}
