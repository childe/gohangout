package input

import (
	"bufio"
	"net"

	"github.com/childe/gohangout/codec"
	"github.com/childe/gohangout/topology"
	"github.com/golang/glog"
)

type UDPInput struct {
	config  map[interface{}]interface{}
	network string
	address string

	decoder codec.Decoder

	conn     *net.UDPConn
	messages chan []byte
	stop     bool
}

func init() {
	Register("UDP", newUDPInput)
}

func newUDPInput(config map[interface{}]interface{}) topology.Input {
	var codertype string = "plain"
	if v, ok := config["codec"]; ok {
		codertype = v.(string)
	}

	p := &UDPInput{
		config:   config,
		decoder:  codec.NewDecoder(codertype),
		messages: make(chan []byte, 10),
	}

	if v, ok := config["max_length"]; ok {
		if max, ok := v.(int); ok {
			if max <= 0 {
				glog.Fatal("max_length must be bigger than zero")
			}
		} else {
			glog.Fatal("max_length must be int")
		}
	}

	p.network = "udp"
	if network, ok := config["network"]; ok {
		p.network = network.(string)
	}

	if addr, ok := config["address"]; ok {
		p.address = addr.(string)
	} else {
		glog.Fatal("address must be set in UDP input")
	}

	udpAddr, err := net.ResolveUDPAddr(p.network, p.address)
	if err != nil {
		glog.Fatalf("resolve udp addr error: %v", err)
	}

	conn, err := net.ListenUDP(p.network, udpAddr)
	if err != nil {
		glog.Fatalf("listen udp error: %v", err)
	}
	p.conn = conn

	scanner := bufio.NewScanner(conn)
	if v, ok := config["max_length"]; ok {
		max := v.(int)
		scanner.Buffer(make([]byte, 0, max), max)
	}
	go func() {
		for {
			for scanner.Scan() {
				t := scanner.Bytes()
				buf := make([]byte, len(t))
				copy(buf, t)
				p.messages <- buf
			}

			if p.stop {
				return
			}

			if err := scanner.Err(); err != nil {
				glog.Errorf("read from %v->%v error: %v", p.conn.RemoteAddr(), p.conn.LocalAddr(), err)
			}
		}
	}()
	return p
}

func (p *UDPInput) ReadOneEvent() map[string]interface{} {
	text, more := <-p.messages
	if !more || text == nil {
		return nil
	}
	return p.decoder.Decode(text)
}

func (p *UDPInput) Shutdown() {
	p.stop = true
	p.conn.Close()
	close(p.messages)
}
