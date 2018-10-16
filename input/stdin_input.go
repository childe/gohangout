package input

import (
	"bufio"
	"io"
	"os"

	"github.com/childe/gohangout/codec"
	"github.com/golang/glog"
)

type StdinInput struct {
	BaseInput

	config  map[interface{}]interface{}
	reader  *bufio.Reader
	decoder codec.Decoder
}

func NewStdinInput(config map[interface{}]interface{}) *StdinInput {
	var codertype string = "plain"
	if v, ok := config["codec"]; ok {
		codertype = v.(string)
	}
	return &StdinInput{
		BaseInput: BaseInput{},

		config:  config,
		reader:  bufio.NewReader(os.Stdin),
		decoder: codec.NewDecoder(codertype),
	}
}

func (inputPlugin *StdinInput) readOneEvent() map[string]interface{} {
	var text []byte = nil
	for {
		line, isPrefix, err := inputPlugin.reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			glog.Errorf("readline error:%s", err)
			return nil
		}
		if text == nil {
			text = line
		} else {
			text = append(text, line...)
		}
		if !isPrefix {
			break
		}
	}
	return inputPlugin.decoder.Decode(string(text))
}

func (inputPlugin *StdinInput) Shutdown() {}
