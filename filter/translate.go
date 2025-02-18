package filter

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/childe/gohangout/topology"
	"github.com/childe/gohangout/value_render"
	"k8s.io/klog/v2"
)

type TranslateFilter struct {
	config          map[interface{}]interface{}
	refreshInterval int
	source          string
	target          string
	sourceVR        value_render.ValueRender
	dictionaryPath  string

	// TODO put code to utils
	dict map[interface{}]interface{}
}

func (plugin *TranslateFilter) parseDict() error {
	var (
		buffer []byte
		err    error
	)
	if strings.HasPrefix(plugin.dictionaryPath, "http://") || strings.HasPrefix(plugin.dictionaryPath, "https://") {
		resp, err := http.Get(plugin.dictionaryPath)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		buffer, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
	} else {
		configFile, err := os.Open(plugin.dictionaryPath)
		if err != nil {
			return err
		}
		fi, _ := configFile.Stat()

		buffer = make([]byte, fi.Size())
		_, err = configFile.Read(buffer)
		if err != nil {
			return err
		}
	}

	dict := make(map[interface{}]interface{})
	err = yaml.Unmarshal(buffer, &dict)
	if err != nil {
		return err
	}
	plugin.dict = dict
	return nil
}

func init() {
	Register("Translate", newTranslateFilter)
}

func newTranslateFilter(config map[interface{}]interface{}) topology.Filter {
	plugin := &TranslateFilter{
		config: config,
	}

	if source, ok := config["source"]; ok {
		plugin.source = source.(string)
	} else {
		klog.Fatal("source must be set in translate filter plugin")
	}
	plugin.sourceVR = value_render.GetValueRender2(plugin.source)

	if target, ok := config["target"]; ok {
		plugin.target = target.(string)
	} else {
		klog.Fatal("target must be set in translate filter plugin")
	}

	if dictionaryPath, ok := config["dictionary_path"]; ok {
		plugin.dictionaryPath = dictionaryPath.(string)
	} else {
		klog.Fatal("dictionary_path must be set in translate filter plugin")
	}

	if refreshInterval, ok := config["refresh_interval"]; ok {
		plugin.refreshInterval = refreshInterval.(int)
	} else {
		klog.Fatal("refresh_interval must be set in translate filter plugin")
	}

	err := plugin.parseDict()
	if err != nil {
		klog.Fatalf("could not parse %s:%s", plugin.dictionaryPath, err)
	}

	ticker := time.NewTicker(time.Second * time.Duration(plugin.refreshInterval))
	go func() {
		for range ticker.C {
			err := plugin.parseDict()
			if err != nil {
				klog.Errorf("could not parse %s:%s", plugin.dictionaryPath, err)
			}
		}
	}()

	return plugin
}

func (plugin *TranslateFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	o, err := plugin.sourceVR.Render(event)
	if err != nil || o == nil {
		return event, false
	}
	if targetValue, ok := plugin.dict[o]; ok {
		event[plugin.target] = targetValue
		return event, true
	}
	return event, false
}
