package input

import (
	"math/rand"
	"strconv"

	"github.com/childe/gohangout/codec"
	"github.com/golang/glog"
)

type RandomInput struct {
	config  map[interface{}]interface{}
	decoder codec.Decoder

	from int
	to   int

	maxMessages int
	count       int
}

func (l *MethodLibrary) NewRandomInput(config map[interface{}]interface{}) *RandomInput {
	var codertype string = "plain"

	p := &RandomInput{
		config:      config,
		decoder:     codec.NewDecoder(codertype),
		count:       0,
		maxMessages: -1,
	}

	if v, ok := config["from"]; ok {
		p.from = v.(int)
	} else {
		glog.Fatal("from must be configured in Random Input")
	}

	if v, ok := config["to"]; ok {
		p.to = v.(int)
	} else {
		glog.Fatal("to must be configured in Random Input")
	}

	if v, ok := config["max_messages"]; ok {
		p.maxMessages = v.(int)
	}

	return p
}

func (p *RandomInput) ReadOneEvent() map[string]interface{} {
	if p.maxMessages != -1 && p.count >= p.maxMessages {
		return nil
	}
	n := p.from + rand.Intn(1+p.to-p.from)
	p.count++
	return p.decoder.Decode([]byte(strconv.Itoa(n)))
}

func (p *RandomInput) Shutdown() {}
