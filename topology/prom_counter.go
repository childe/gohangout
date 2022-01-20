package topology

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func GetPromCounter(config map[interface{}]interface{}) prometheus.Counter {
	if promName, ok := config["prometheus_counter_name"]; ok {
		return promauto.NewCounter(prometheus.CounterOpts{
			Name: promName.(string),
		})
	}
	return nil
}
