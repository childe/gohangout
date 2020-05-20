package main

import (
	"flag"
	"fmt"
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
	autoReload bool // 配置文件更新自动重启
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
	flag.BoolVar(&options.autoReload, "reload", options.autoReload, "if auto reload while config file changed")

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
	defer func() {
		// 一些强制类型转换出错导致panic的保护
		if e := recover(); e != nil {
			err = fmt.Errorf("[PANIC] %v", e)
		}
	}()

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

	go func() {
		// 监听配置文件更新
		for cfg := range configChannel {
			// 停止所有的流水线
			StopBoxesBeat()
			// 重新根据配置文件加载流水线
			newBoxes, err := buildPluginLink(cfg)
			if err == nil {
				boxes = newBoxes
				// 如果没报错才会执行更新
				go StartBoxesBeat()
			} else {
				glog.Errorf("build plugin link fail: %s", err)
			}
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
		glog.Infof("config:\n%s", removeSensitiveInfo(config))
		configChannel <- config
	}

	ListenSignal()
}

var boxes []*input.InputBox
var configChannel = make(chan map[string]interface{})

func StartBoxesBeat() {
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

func StopBoxesBeat() {
	for _, box := range boxes {
		box.Shutdown()
	}

	boxes = make([]*input.InputBox, 0)
}

func ListenSignal() {
	c := make(chan os.Signal, 1)
	var stop bool
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)

	defer glog.Infof("listen signal stop, exit...")

	for sig := range c {
		glog.Infof("capture signal: %v", sig)
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			StopBoxesBeat()
			close(configChannel)
			stop = true
		case syscall.SIGUSR1:
			// `kill -USR1 pid`也会触发重新加载
			config, err := parseConfig(options.config)
			if err != nil {
				glog.Errorf("could not parse config:%s", err)
				continue
			}
			glog.Infof("config:\n%s", removeSensitiveInfo(config))
			configChannel <- config
		}

		if stop {
			break
		}
	}
}
