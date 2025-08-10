package topology

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"k8s.io/klog/v2"
)

var lock = sync.Mutex{}
var counterManager map[string]prometheus.Counter = make(map[string]prometheus.Counter)

// convertToJSONCompatible recursively converts map[any]any to map[string]any for JSON compatibility
func convertToJSONCompatible(input any) any {
	switch v := input.(type) {
	case map[any]any:
		result := make(map[string]any)
		for k, val := range v {
			if keyStr, ok := k.(string); ok {
				result[keyStr] = convertToJSONCompatible(val)
			} else {
				return nil // Skip non-string keys
			}
		}
		return result
	case []any:
		result := make([]any, len(v))
		for i, val := range v {
			result[i] = convertToJSONCompatible(val)
		}
		return result
	default:
		return v
	}
}

// decodePromConfig safely decodes configuration using encoding/json
func decodePromConfig(config any, result any) error {
	// Convert config to JSON-serializable format
	jsonConfig := convertToJSONCompatible(config)

	// Convert map to JSON and then unmarshal to struct
	jsonBytes, err := json.Marshal(jsonConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %v", err)
	}

	if err := json.Unmarshal(jsonBytes, result); err != nil {
		return fmt.Errorf("configuration error: %v", err)
	}
	return nil
}

func hashValue(opts prometheus.CounterOpts) string {
	opts.Help = ""
	b, _ := json.Marshal(opts)
	return string(b)
}

// GetPromCounter creates a prometheus.Counter from config.
// if same config exits before, GetPromCounter would return the counter created before. Because tow counters with the same config leads to panic.
// Better practice maybe to let it panic, so owner can fix the config when program fails to start.
// But if user use multi workers to run gohangout, panic are bound to happen, this is bad. So we use a manager to return one counter with save config.
// Better way is to add {worker: idx} to ConstLabels, but it is too hard to implement it by code.
func GetPromCounter(config map[any]any) prometheus.Counter {
	lock.Lock()
	defer lock.Unlock()
	if promConf, ok := config["prometheus_counter"]; ok {
		// promConf := promConf.(map[any]any)

		var opts prometheus.CounterOpts = prometheus.CounterOpts{}
		err := decodePromConfig(promConf, &opts)
		if err != nil {
			klog.Errorf("marshal prometheus counter config error: %v", err)
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
