package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"

	"github.com/childe/gohangout/input"
	"github.com/childe/gohangout/topology"
	"github.com/golang/glog"
)

var options = &struct {
	config     string
	autoReload bool // 配置文件更新自动重启
	pprof      bool
	pprofAddr  string
	cpuprofile string
	memprofile string

	exitWhenNil bool
}{}

var (
	worker = flag.Int("worker", 1, "worker thread count")
)

type gohangoutInputs []*input.InputBox

var inputs gohangoutInputs

var mainThreadExitChan chan struct{} = make(chan struct{}, 0)

func (inputs gohangoutInputs) start() {
	boxes := ([]*input.InputBox)(inputs)
	var wg sync.WaitGroup
	wg.Add(len(boxes))

	for i := range boxes {
		go func(i int) {
			defer wg.Done()
			boxes[i].Beat(*worker)
		}(i)
	}

	wg.Wait()
}

func (inputs gohangoutInputs) stop() {
	boxes := ([]*input.InputBox)(inputs)
	for _, box := range boxes {
		box.Shutdown()
	}
}

func init() {
	flag.StringVar(&options.config, "config", options.config, "path to configuration file or directory")
	flag.BoolVar(&options.autoReload, "reload", options.autoReload, "if auto reload while config file changed")

	flag.BoolVar(&options.pprof, "pprof", false, "pprof or not")
	flag.StringVar(&options.pprofAddr, "pprof-address", "127.0.0.1:8899", "default: 127.0.0.1:8899")
	flag.StringVar(&options.cpuprofile, "cpuprofile", "", "write cpu profile to `file`")
	flag.StringVar(&options.memprofile, "memprofile", "", "write mem profile to `file`")

	flag.BoolVar(&options.exitWhenNil, "exit-when-nil", false, "triger gohangout to exit when receive a nil event")

	flag.Parse()
}

func buildPluginLink(config map[string]interface{}) (boxes []*input.InputBox, err error) {
	boxes = make([]*input.InputBox, 0)

	for inputIdx, inputI := range config["inputs"].([]interface{}) {
		var inputPlugin topology.Input

		i := inputI.(map[interface{}]interface{})
		glog.Infof("input[%d] %v", inputIdx+1, i)

		// len(i) is 1
		for inputTypeI, inputConfigI := range i {
			inputType := inputTypeI.(string)
			inputConfig := inputConfigI.(map[interface{}]interface{})

			inputPlugin = input.GetInput(inputType, inputConfig)
			if inputPlugin == nil {
				err = fmt.Errorf("invalid input plugin")
				return
			}

			box := input.NewInputBox(inputPlugin, inputConfig, config, mainThreadExitChan)
			if box == nil {
				err = fmt.Errorf("new input box fail")
				return
			}
			box.SetShutdownWhenNil(options.exitWhenNil)
			boxes = append(boxes, box)
		}
	}

	return
}

func main() {
	printVersion()
	defer glog.Flush()

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
		glog.Fatalf("could not parse config: %v", err)
	}
	boxes, err := buildPluginLink(config)
	if err != nil {
		glog.Fatalf("build plugin link error: %v", err)
	}
	inputs = gohangoutInputs(boxes)
	go inputs.start()

	go func() {
		for cfg := range configChannel {
			inputs.stop()
			boxes, err := buildPluginLink(cfg)
			if err == nil {
				inputs = gohangoutInputs(boxes)
				go inputs.start()
			} else {
				glog.Errorf("build plugin link error: %v", err)
				exit()
			}
		}
	}()

	if options.autoReload {
		if err := watchConfig(options.config, configChannel); err != nil {
			glog.Fatalf("watch config fail: %s", err)
		}
	}

	go listenSignal()

	<-mainThreadExitChan
	inputs.stop()
}

var configChannel = make(chan map[string]interface{})

func exit() {
	mainThreadExitChan <- struct{}{}
}
