package input

import (
	"net"

	"github.com/childe/gohangout/codec"
	"github.com/childe/gohangout/topology"
	"k8s.io/klog/v2"
)

type msg struct {
	message []byte
	addr    *net.UDPAddr
}
type UDPInput struct {
	config        map[interface{}]interface{}
	network       string
	address       string
	addRemoteAddr string

	decoder codec.Decoder

	conn     *net.UDPConn
	messages chan msg
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
		messages: make(chan msg, 10),
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

	p.network = "udp"
	if network, ok := config["network"]; ok {
		p.network = network.(string)
	}

	if addr, ok := config["address"]; ok {
		p.address = addr.(string)
	} else {
		klog.Fatal("address must be set in UDP input")
	}

	udpAddr, err := net.ResolveUDPAddr(p.network, p.address)
	if err != nil {
		klog.Fatalf("resolve udp addr error: %v", err)
	}

	conn, err := net.ListenUDP(p.network, udpAddr)
	if err != nil {
		klog.Fatalf("listen udp error: %v", err)
	}
	p.conn = conn

	if v, ok := config["add_remote_addr"]; ok {
		p.addRemoteAddr = v.(string)
	}

	var max int = 65535
	if v, ok := config["max_length"]; ok {
		max = v.(int)
	}

	go func() {
		for !p.stop {
			buf := make([]byte, max)
			n, addr, err := p.conn.ReadFromUDP(buf)
			if err != nil {
				if p.stop {
					return
				}
				klog.Errorf("read from UDP error: %v", err)
			}
			p.messages <- msg{
				message: buf[:n],
				addr:    addr,
			}
		}
	}()
	return p
}

func (p *UDPInput) ReadOneEvent() map[string]interface{} {
	msg, more := <-p.messages
	if !more {
		return nil
	}
	event := p.decoder.Decode(msg.message)

	if p.addRemoteAddr != "" && msg.addr != nil {
		event[p.addRemoteAddr] = msg.addr.IP.String()
	}

	return event
}

func (p *UDPInput) Shutdown() {
	p.stop = true
	p.conn.Close()
	close(p.messages)
}
