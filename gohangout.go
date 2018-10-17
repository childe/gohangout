package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/childe/gohangout/filter"
	"github.com/childe/gohangout/input"
	"github.com/childe/gohangout/output"
	"github.com/golang/glog"
	"github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary
var options = &struct {
	config    string
	pprof     bool
	pprofAddr string
}{}
var boxes []*input.InputBox

func init() {
	flag.StringVar(&options.config, "config", options.config, "path to configuration file or directory")

	flag.BoolVar(&options.pprof, "pprof", false, "if pprof")
	flag.StringVar(&options.pprofAddr, "pprof-address", "127.0.0.1:8899", "default: 127.0.0.1:8899")

	flag.Parse()
}

func buildPluginLink(config map[string]interface{}) []input.Input {
	outputs := output.BuildOutputs(config)
	filters := filter.BuildFilters(config, outputs)

	inputs := make([]input.Input, 0)

	var inputPlugin input.Input
	for input_idx, inputI := range config["inputs"].([]interface{}) {
		i := inputI.(map[interface{}]interface{})
		glog.Infof("input[%d] %v", input_idx+1, i)

		// len(i) is 1
		for inputTypeI, inputConfigI := range i {
			inputType := inputTypeI.(string)
			inputConfig := inputConfigI.(map[interface{}]interface{})

			if len(filters) > 0 {
				inputPlugin = input.GetInput(inputType, inputConfig, filters[0], nil)
			} else {
				inputPlugin = input.GetInput(inputType, inputConfig, nil, outputs)
			}

			inputs = append(inputs, inputPlugin)
		}
	}
	return inputs
}

func main() {
	if options.pprof {
		go func() {
			http.ListenAndServe(options.pprofAddr, nil)
		}()
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for {
			<-c
			signal.Stop(c)
			for _, box := range boxes {
				box.Shutdown()
			}
			//os.Exit(0) // leave the program leave itself
		}
	}()

	config, err := parseConfig(options.config)
	if err != nil {
		glog.Fatalf("could not parse config:%s", err)
	}
	glog.Infof("%v", config)

	inputs := buildPluginLink(config)

	var wg sync.WaitGroup
	wg.Add(len(inputs))
	defer wg.Wait()

	boxes = make([]*input.InputBox, len(inputs))
	for input_idx, inputPlugin := range inputs {
		box := input.NewInputBox(inputPlugin)
		boxes[input_idx] = &box

		go func() {
			defer wg.Done()
			box.Beat()
		}()
	}

}
