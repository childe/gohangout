package filter

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/childe/gohangout/value_render"
	"github.com/golang/glog"
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

func (l *MethodLibrary) NewTranslateFilter(config map[interface{}]interface{}) *TranslateFilter {
	plugin := &TranslateFilter{
		config: config,
	}

	if source, ok := config["source"]; ok {
		plugin.source = source.(string)
	} else {
		glog.Fatal("source must be set in translate filter plugin")
	}
	plugin.sourceVR = value_render.GetValueRender2(plugin.source)

	if target, ok := config["target"]; ok {
		plugin.target = target.(string)
	} else {
		glog.Fatal("target must be set in translate filter plugin")
	}

	if dictionaryPath, ok := config["dictionary_path"]; ok {
		plugin.dictionaryPath = dictionaryPath.(string)
	} else {
		glog.Fatal("dictionary_path must be set in translate filter plugin")
	}

	if refreshInterval, ok := config["refresh_interval"]; ok {
		plugin.refreshInterval = refreshInterval.(int)
	} else {
		glog.Fatal("refresh_interval must be set in translate filter plugin")
	}

	err := plugin.parseDict()
	if err != nil {
		glog.Fatalf("could not parse %s:%s", plugin.dictionaryPath, err)
	}

	ticker := time.NewTicker(time.Second * time.Duration(plugin.refreshInterval))
	go func() {
		for range ticker.C {
			err := plugin.parseDict()
			if err != nil {
				glog.Errorf("could not parse %s:%s", plugin.dictionaryPath, err)
			}
		}
	}()

	return plugin
}

func (plugin *TranslateFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	o := plugin.sourceVR.Render(event)
	if o == nil {
		return event, false
	}
	if targetValue, ok := plugin.dict[o]; ok {
		event[plugin.target] = targetValue
		return event, true
	}
	return event, false
}
