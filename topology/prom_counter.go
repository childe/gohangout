package topology

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/mitchellh/mapstructure"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func GetPromCounter(config map[interface{}]interface{}, worker int) prometheus.Counter {
	if promConf, ok := config["prometheus_counter"]; ok {
		promConf := promConf.(map[interface{}]interface{})
		var opts prometheus.CounterOpts = prometheus.CounterOpts{}
		err := mapstructure.Decode(promConf, &opts)

		workerIdx := fmt.Sprintf("%d", worker)
		if opts.ConstLabels == nil {
			opts.ConstLabels = map[string]string{"worker": workerIdx}
		} else {
			opts.ConstLabels["worker"] = workerIdx
		}

		if err != nil {
			glog.Errorf("marshal prometheus counter config error: %v", err)
			return nil
		}
		return promauto.NewCounter(opts)
	}

	return nil
}
