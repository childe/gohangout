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
}

func NewStdinInput(config map[interface{}]interface{}) *StdinInput {
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
		for p.scanner.Scan() {
			t := p.scanner.Text()
			p.messages <- []byte(t)
		}
		if err := p.scanner.Err(); err != nil {
			glog.Errorf("%s", err)
		}

		// shutdown
		p.messages <- nil
	}()
	return p
}

func (p *StdinInput) readOneEvent() map[string]interface{} {
	text, more := <-p.messages
	if !more || text == nil {
		return nil
	}
	return p.decoder.Decode(text)
}

func (p *StdinInput) Shutdown() {
	close(p.messages)
}
