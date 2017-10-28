package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"sync"

	"github.com/golang/glog"
)

var options = &struct {
	config    string
	pprof     bool
	pprofAddr string
}{}

func init() {
	flag.StringVar(&options.config, "config", options.config, "path to configuration file or directory")

	flag.BoolVar(&options.pprof, "pprof", false, "if pprof")
	flag.StringVar(&options.pprofAddr, "pprof-address", "127.0.0.1:8899", "default: 127.0.0.1:8899")

	flag.Parse()
}

func getOutputs(config map[string]interface{}) []Output {
	if outputValue, ok := config["outputs"]; ok {
		rst := make([]Output, 0)
		outputs := outputValue.([]interface{})
		for _, outputValue := range outputs {
			output := outputValue.(map[interface{}]interface{})
			glog.Info(output)
			for k, v := range output {
				outputType := k.(string)
				glog.Info(outputType)
				outputConfig := v.(map[interface{}]interface{})
				glog.Info(outputConfig)
				outputPlugin := getOutput(outputType)
				rst = append(rst, outputPlugin)
			}
		}
		return rst
	} else {
		return nil
	}
}

func getFilters(config map[string]interface{}) []Filter {
	if filterValue, ok := config["filters"]; ok {
		rst := make([]Filter, 0)
		filters := filterValue.([]interface{})
		for _, filterValue := range filters {
			filter := filterValue.(map[interface{}]interface{})
			glog.Info(filter)
			for k, v := range filter {
				filterType := k.(string)
				glog.Info(filterType)
				filterConfig := v.(map[interface{}]interface{})
				glog.Info(filterConfig)
				filterPlugin := getFilter(filterType, filterConfig)
				rst = append(rst, filterPlugin)
			}
		}
		return rst
	} else {
		return nil
	}
}

func main() {
	if options.pprof {
		go func() {
			http.ListenAndServe(options.pprofAddr, nil)
		}()
	}

	config, err := parseConfig(options.config)
	if err != nil {
		glog.Fatal(err)
	}
	//glog.Infof("%v", config)

	if inputValue, ok := config["inputs"]; ok {
		//glog.Info(inputValue)
		var wg sync.WaitGroup
		inputs := inputValue.([]interface{})
		wg.Add(len(inputs))
		for _, inputValue := range inputs {
			input := inputValue.(map[interface{}]interface{})
			glog.Info(input)
			for k, v := range input {
				inputType := k.(string)
				glog.Info(inputType)
				inputConfig := v.(map[interface{}]interface{})
				glog.Info(inputConfig)

				inputPlugin := getInput(inputType)
				//inputPlugin.config = inputPlugin
				box := InputBox{
					input:   inputPlugin,
					filters: getFilters(config),
					outputs: getOutputs(config),
					config:  inputConfig,
				}

				go func() {
					defer wg.Done()
					box.beat()
				}()
			}
		}
		wg.Wait()
	} else {
		glog.Fatal("could not find inputs in config file")
	}
}
