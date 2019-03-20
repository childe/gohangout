package input

import (
	"bufio"
	"os"
	"sync"
	"time"

	"github.com/childe/gohangout/codec"
	"github.com/golang/glog"
)

type StdinInput struct {
	config  map[interface{}]interface{}
	decoder codec.Decoder

	scanner  *bufio.Scanner
	messages chan []byte

	once sync.Once
	stop bool
	done chan bool
}

func (p *StdinInput) closeMessagesChan() {
	p.once.Do(func() { close(p.messages) })
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
		done:     make(chan bool, 0),
	}

	go func() {
		defer func() { p.done <- true }()
		for !p.stop && p.scanner.Scan() {
			t := p.scanner.Text()
			p.messages <- []byte(t)
		}
		if err := p.scanner.Err(); err != nil {
			glog.Errorf("%s", err)
		}

		// trigger shutdown
		p.closeMessagesChan()
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
	p.stop = true
	select {
	case <-p.done:
	case <-time.After(time.Second * 3):
	}
	p.closeMessagesChan()
}
