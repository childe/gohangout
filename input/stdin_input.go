package input

import (
	"bufio"
	"os"

	"github.com/childe/gohangout/codec"
	"github.com/golang/glog"
)

type StdinInput struct {
	config  map[interface{}]interface{}
	decoder codec.Decoder

	scanner  *bufio.Scanner
	messages chan []byte

	stop bool
}

func (l *MethodLibrary) NewStdinInput(config map[interface{}]interface{}) *StdinInput {
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

	go func() {
		for p.scanner.Scan() && !p.stop {
			t := p.scanner.Bytes()
			msg := make([]byte, len(t))
			copy(msg, t)
			p.messages <- msg
		}
		if err := p.scanner.Err(); err != nil {
			glog.Errorf("%s", err)
		}

		// trigger shutdown
		close(p.messages)
	}()
	return p
}

func (p *StdinInput) ReadOneEvent() map[string]interface{} {
	text, more := <-p.messages
	if !more || text == nil {
		return nil
	}
	return p.decoder.Decode(text)
}

func (p *StdinInput) Shutdown() {
	// what we need is to stop emit new event; close messages or not is not important
	p.stop = true
}
