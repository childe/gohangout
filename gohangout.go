package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"sync"
	"syscall"

	"github.com/childe/gohangout/input"
	"github.com/childe/gohangout/topology"
	"github.com/golang/glog"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary
var options = &struct {
	config     string
	pprof      bool
	pprofAddr  string
	cpuprofile string
	memprofile string
}{}

var gitCommit string

func printVersion() {
	glog.Info("Current build version: ", gitCommit)
}

var (
	worker = flag.Int("worker", 1, "worker thread count")
)

func init() {
	flag.StringVar(&options.config, "config", options.config, "path to configuration file or directory")

	flag.BoolVar(&options.pprof, "pprof", false, "if pprof")
	flag.StringVar(&options.pprofAddr, "pprof-address", "127.0.0.1:8899", "default: 127.0.0.1:8899")
	flag.StringVar(&options.cpuprofile, "cpuprofile", "", "write cpu profile to `file`")
	flag.StringVar(&options.memprofile, "memprofile", "", "write mem profile to `file`")

	flag.Parse()
}

func init() {
	printVersion()
}

func buildPluginLink(config map[string]interface{}) []*input.InputBox {
	boxes := make([]*input.InputBox, 0)

	for inputIdx, inputI := range config["inputs"].([]interface{}) {
		var inputPlugin topology.Input

		i := inputI.(map[interface{}]interface{})
		glog.Infof("input[%d] %v", inputIdx+1, i)

		// len(i) is 1
		for inputTypeI, inputConfigI := range i {
			inputType := inputTypeI.(string)
			inputConfig := inputConfigI.(map[interface{}]interface{})

			inputPlugin = input.GetInput(inputType, inputConfig)

			box := input.NewInputBox(inputPlugin, config)
			boxes = append(boxes, box)
		}
	}
	return boxes
}

func main() {
	if options.pprof {
		go func() {
			http.ListenAndServe(options.pprofAddr, nil)
		}()
	}
	if options.cpuprofile != "" {
		f, err := os.Create(options.cpuprofile)
		if err != nil {
			glog.Fatalf("could not create CPU profile: %s", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			glog.Fatalf("could not start CPU profile: %s", err)
		}
		defer pprof.StopCPUProfile()
	}

	if options.memprofile != "" {
		defer func() {
			f, err := os.Create(options.memprofile)
			if err != nil {
				glog.Fatalf("could not create memory profile: %s", err)
			}
			defer f.Close()
			runtime.GC() // get up-to-date statistics
			if err := pprof.WriteHeapProfile(f); err != nil {
				glog.Fatalf("could not write memory profile: %s", err)
			}
		}()
	}

	config, err := parseConfig(options.config)
	if err != nil {
		glog.Fatalf("could not parse config:%s", err)
	}
	glog.Infof("%v", config)

	boxes := buildPluginLink(config)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for {
			<-c
			signal.Stop(c)
			for _, box := range boxes {
				box.Shutdown()
			}
		}
	}()

	var wg sync.WaitGroup
	wg.Add(len(boxes))
	defer wg.Wait()

	for i, _ := range boxes {
		go func(i int) {
			defer wg.Done()
			boxes[i].Beat(*worker)
		}(i)
	}
}
