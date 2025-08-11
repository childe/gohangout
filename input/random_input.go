package input

import (
	"math/rand"
	"strconv"

	"github.com/childe/gohangout/codec"
	"github.com/childe/gohangout/topology"
	"k8s.io/klog/v2"
)

type RandomInput struct {
	config  map[any]any
	decoder codec.Decoder

	from int
	to   int

	maxMessages int
	count       int
}

func init() {
	Register("Random", newRandomInput)
}

func newRandomInput(config map[any]any) topology.Input {
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
		klog.Fatal("from must be configured in Random Input")
	}

	if v, ok := config["to"]; ok {
		p.to = v.(int)
	} else {
		klog.Fatal("to must be configured in Random Input")
	}

	if v, ok := config["max_messages"]; ok {
		p.maxMessages = v.(int)
	}

	return p
}

func (p *RandomInput) ReadOneEvent() map[string]any {
	if p.maxMessages != -1 && p.count >= p.maxMessages {
		return nil
	}
	n := p.from + rand.Intn(1+p.to-p.from)
	p.count++
	return p.decoder.Decode([]byte(strconv.Itoa(n)))
}

func (p *RandomInput) Shutdown() {}
