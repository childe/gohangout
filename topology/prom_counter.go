package topology

import (
	"github.com/golang/glog"
	"github.com/mitchellh/mapstructure"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func GetPromCounter(config map[interface{}]interface{}) prometheus.Counter {
	if promConf, ok := config["prometheus_counter"]; ok {
		promConf := promConf.(map[interface{}]interface{})
		var opts prometheus.CounterOpts = prometheus.CounterOpts{}
		err := mapstructure.Decode(promConf, &opts)
		if err != nil {
			glog.Errorf("marshal prometheus counter config error: %v", err)
			return nil
		}
		return promauto.NewCounter(opts)
	}

	return nil
}
