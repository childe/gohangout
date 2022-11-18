package topology

import (
	"encoding/json"
	"sync"

	"github.com/golang/glog"
	"github.com/mitchellh/mapstructure"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var lock = sync.Mutex{}
var counterManager map[string]prometheus.Counter = make(map[string]prometheus.Counter)

func hashValue(opts prometheus.CounterOpts) string {
	opts.Help = ""
	b, _ := json.Marshal(opts)
	return string(b)
}

// GetPromCounter creates a prometheus.Counter from config.
// if same config exsits before, GetPromCounter would return the counter created before. Because tow counters with the same config leads to panic.
// Better practice maybe to let it panic, so owner can fix the config when program fails to start.
// But if user use multi workers to run gohangout, panic are bound to happen, this is bad. So we use a manager to return one counter with save config.
// Better way is to add {worker: idx} to ConstLabels, but it is too hard to implement it by code.
func GetPromCounter(config map[interface{}]interface{}) prometheus.Counter {
	lock.Lock()
	defer lock.Unlock()
	if promConf, ok := config["prometheus_counter"]; ok {
		// promConf := promConf.(map[interface{}]interface{})

		var opts prometheus.CounterOpts = prometheus.CounterOpts{}
		err := mapstructure.Decode(promConf, &opts)
		if err != nil {
			glog.Errorf("marshal prometheus counter config error: %v", err)
			return nil
		}

		key := hashValue(opts)

		if v, ok := counterManager[key]; ok {
			return v
		}
		c := promauto.NewCounter(opts)
		counterManager[key] = c
		return c
	}

	return nil
}
