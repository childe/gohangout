package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"sync"

	"github.com/childe/gohangout/filter"
	"github.com/childe/gohangout/input"
	"github.com/childe/gohangout/output"
	"github.com/childe/glog"
	"github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary
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

func getOutputs(config map[string]interface{}) []output.Output {
	if outputValue, ok := config["outputs"]; ok {
		rst := make([]output.Output, 0)
		outputs := outputValue.([]interface{})
		for _, outputValue := range outputs {
			o := outputValue.(map[interface{}]interface{})
			for k, v := range o {
				outputType := k.(string)
				glog.Infof("output type:%s", outputType)
				outputConfig := v.(map[interface{}]interface{})
				glog.Infof("output config:%v", outputConfig)
				outputPlugin := output.GetOutput(outputType, outputConfig)
				if outputPlugin == nil {
					glog.Fatalf("could build output plugin from type (%s)", outputType)
				}
				rst = append(rst, outputPlugin)
			}
		}
		return rst
	} else {
		return nil
	}
}

func getFilters(config map[string]interface{}) []filter.Filter {
	if filterValue, ok := config["filters"]; ok {
		rst := make([]filter.Filter, 0)
		filters := filterValue.([]interface{})
		for _, filterValue := range filters {
			filters := filterValue.(map[interface{}]interface{})
			for k, v := range filters {
				filterType := k.(string)
				glog.Infof("filter type:%s", filterType)
				filterConfig := v.(map[interface{}]interface{})
				glog.Infof("filter config:%v", filterConfig)
				filterPlugin := filter.GetFilter(filterType, filterConfig)
				if filterPlugin == nil {
					glog.Fatalf("could build filter plugin from type (%s)", filterType)
				}
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
		glog.Fatalf("could not parse config:%s", err)
	}
	glog.Infof("%v", config)

	if inputValue, ok := config["inputs"]; ok {
		//glog.Info(inputValue)
		var wg sync.WaitGroup
		inputs := inputValue.([]interface{})
		wg.Add(len(inputs))
		for _, inputValue := range inputs {
			i := inputValue.(map[interface{}]interface{})
			glog.Info(i)
			for k, v := range i {
				inputType := k.(string)
				glog.Info(inputType)
				inputConfig := v.(map[interface{}]interface{})
				glog.Info(inputConfig)

				inputPlugin := input.GetInput(inputType, inputConfig)
				//inputPlugin.config = inputPlugin
				box := input.NewInputBox(inputPlugin, getFilters(config), getOutputs(config), inputConfig)

				go func() {
					defer wg.Done()
					box.Beat()
				}()
			}
		}
		wg.Wait()
	} else {
		glog.Fatal("could not find inputs in config file")
	}
}
