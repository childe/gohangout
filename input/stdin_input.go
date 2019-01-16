package input

import (
	"bufio"
	"os"

	"github.com/childe/gohangout/codec"
	"github.com/golang/glog"
)

type StdinInput struct {
	config  map[interface{}]interface{}
	reader  *bufio.Reader
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
		reader:   bufio.NewReader(os.Stdin),
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
		close(p.messages)
	}()
	return p
}

func (p *StdinInput) readOneEvent() map[string]interface{} {
	text := <-p.messages
	if text == nil {
		return nil
	}
	return p.decoder.Decode(text)
}

func (p *StdinInput) Shutdown() {}
