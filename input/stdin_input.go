package input

import (
	"bufio"
	"os"
	"time"

	"github.com/childe/gohangout/codec"
	"github.com/childe/gohangout/topology"
	"github.com/golang/glog"
)

type StdinInput struct {
	config  map[interface{}]interface{}
	decoder codec.Decoder

	scanner  *bufio.Scanner
	messages chan []byte

	stop bool
}

func init() {
	Register("Stdin", newStdinInput)
}

func newStdinInput(config map[interface{}]interface{}) topology.Input {
	var codertype string = "plain"
	if v, ok := config["codec"]; ok {
		codertype = v.(string)
	}
	p := &StdinInput{

		config:   config,
		decoder:  codec.NewDecoder(codertype),
		scanner:  bufio.NewScanner(os.Stdin),
		messages: make(chan []byte, 10),
	}

	return p
}

func (p *StdinInput) ReadOneEvent() map[string]interface{} {
	if p.scanner.Scan() {
		t := p.scanner.Bytes()
		msg := make([]byte, len(t))
		copy(msg, t)
		return p.decoder.Decode(msg)
	}
	if err := p.scanner.Err(); err != nil {
		glog.Errorf("stdin scan error: %v", err)
	} else {
		// EOF here. when stdin is closed by C-D, cpu will raise up to 100% if not sleep
		time.Sleep(time.Millisecond * 1000)
	}
	return nil
}

func (p *StdinInput) Shutdown() {
	// what we need is to stop emit new event; close messages or not is not important
	p.stop = true
}
