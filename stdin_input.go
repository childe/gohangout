package main

import (
	"bufio"
	"os"
	"time"

	"github.com/childe/gohangout/codec"
	"github.com/golang/glog"
)

type StdinInput struct {
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
		config:  config,
		reader:  bufio.NewReader(os.Stdin),
		decoder: codec.NewDecoder(codertype),
	}
}

func (inputPlugin *StdinInput) readOneEvent() map[string]interface{} {
	text, isPrefix, err := inputPlugin.reader.ReadLine()
	if err != nil {
		glog.Errorf("readline error:%s", err)
		return nil
	}
	if isPrefix {
		glog.Errorf("readline got only prefix")
		return nil
	}
	rst := make(map[string]interface{})
	rst["message"] = string(text)
	rst["@timestamp"] = time.Now()
	return rst
}
