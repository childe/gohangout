package main

import (
	"encoding/json"
	"fmt"

	"github.com/golang/glog"
)

type StdoutOutput struct {
	config map[interface{}]interface{}
}

func NewStdoutOutput(config map[interface{}]interface{}) *StdoutOutput {
	return &StdoutOutput{config}
}

func (outputPlugin *StdoutOutput) emit(event map[string]interface{}) {
	buf, err := json.Marshal(event)
	if err != nil {
		glog.Errorf("marshal %v error:%s", event, err)
	}
	fmt.Println(string(buf))
}
