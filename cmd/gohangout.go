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
}{}

func printVersion() {
	glog.Info("Current build version: ", gitCommit)
}

var (
	gitCommit       string
	worker          = flag.Int("worker", 1, "worker thread count")
	boxes           []*input.InputBox
	reloadBoxedLock sync.Mutex
)

func init() {
	flag.StringVar(&options.config, "config", options.config, "path to configuration file or directory")
	flag.BoolVar(&options.autoReload, "reload", true, "if auto reload while config file changed")

	flag.BoolVar(&options.pprof, "pprof", false, "if pprof")
	flag.StringVar(&options.pprofAddr, "pprof-address", "127.0.0.1:8899", "default: 127.0.0.1:8899")
	flag.StringVar(&options.cpuprofile, "cpuprofile", "", "write cpu profile to `file`")
	flag.StringVar(&options.memprofile, "memprofile", "", "write mem profile to `file`")

	flag.Parse()
}

func init() {
	printVersion()
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

			box := input.NewInputBox(inputPlugin, inputConfig, config)
			if box == nil {
				err = fmt.Errorf("new input box fail")
				return
			}
			boxes = append(boxes, box)
		}
	}

	return
}

func main() {
	// flush保证停止的时候日志都写入了文件中
	defer glog.Flush()
	defer stopBoxesBeat()

	if options.pprof {
		go func() {
			http.ListenAndServe(options.pprofAddr, nil)
		}()
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

	// 监听配置文件更新
	configChannel := make(chan map[string]interface{})
	go func() {
		for cfg := range configChannel {
			ReloadBoxes(cfg)
		}
	}()

	// 初始化配置文件
	if options.autoReload {
		if err := watchConfig(options.config, configChannel); err != nil {
			glog.Fatalf("watch config fail: %s", err)
		}
	} else {
		config, err := parseConfig(options.config)
		if err != nil {
			glog.Fatalf("could not parse config:%s", err)
		}
		ReloadBoxes(config)
	}

	listenSignal()
}

// ReloadBoxes stop current boxes and start new ones.
// it will do nothing if config is not valid
func ReloadBoxes(config map[string]interface{}) {
	reloadBoxedLock.Lock()
	defer reloadBoxedLock.Unlock()

	glog.Infof("config:\n%s", removeSensitiveInfo(config))

	newBoxes, err := buildPluginLink(config)
	if err != nil {
		glog.Errorf("could not build plugins from config: %s", err)
		return
	}
	stopBoxesBeat()
	startBoxesBeat(newBoxes)
}

func startBoxesBeat(newBoxes []*input.InputBox) {
	for i := range newBoxes {
		go func(i int) {
			newBoxes[i].Beat(*worker)
		}(i)
	}

	boxes = newBoxes
}

func stopBoxesBeat() {
	for _, box := range boxes {
		box.Shutdown()
	}
}
