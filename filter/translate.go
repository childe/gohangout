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

// TranslateConfig defines the configuration structure for Translate filter
type TranslateConfig struct {
	Source          string `json:"source"`
	Target          string `json:"target"`
	DictionaryPath  string `json:"dictionary_path"`
	RefreshInterval int    `json:"refresh_interval"`
}

type TranslateFilter struct {
	config          map[any]any
	refreshInterval int
	source          string
	target          string
	sourceVR        value_render.ValueRender
	dictionaryPath  string

	// TODO put code to utils
	dict map[any]any
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

	dict := make(map[any]any)
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

func newTranslateFilter(config map[any]any) topology.Filter {
	// Parse configuration using SafeDecodeConfig helper
	var translateConfig TranslateConfig
	
	SafeDecodeConfig("Translate", config, &translateConfig)
	
	// Validate required fields
	ValidateRequiredFields("Translate", map[string]any{
		"source":           translateConfig.Source,
		"target":           translateConfig.Target,
		"dictionary_path":  translateConfig.DictionaryPath,
		"refresh_interval": translateConfig.RefreshInterval,
	})

	plugin := &TranslateFilter{
		config:          config,
		source:          translateConfig.Source,
		target:          translateConfig.Target,
		dictionaryPath:  translateConfig.DictionaryPath,
		refreshInterval: translateConfig.RefreshInterval,
	}

	plugin.sourceVR = value_render.GetValueRender2(plugin.source)

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

func (plugin *TranslateFilter) Filter(event map[string]any) (map[string]any, bool) {
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
