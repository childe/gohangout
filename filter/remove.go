package filter

import (
	"github.com/childe/gohangout/field_deleter"
	"github.com/childe/gohangout/topology"
	"k8s.io/klog/v2"
)

// RemoveConfig defines the configuration structure for Remove filter
type RemoveConfig struct {
	Fields []string `json:"fields"`
}

type RemoveFilter struct {
	config         map[any]any
	fieldsDeleters []field_deleter.FieldDeleter
}

func init() {
	Register("Remove", newRemoveFilter)
}

func newRemoveFilter(config map[any]any) topology.Filter {
	// Parse configuration using SafeDecodeConfig helper
	var removeConfig RemoveConfig

	SafeDecodeConfig("Remove", config, &removeConfig)

	// Validate required fields
	ValidateRequiredFields("Remove", map[string]any{
		"fields": removeConfig.Fields,
	})
	if len(removeConfig.Fields) == 0 {
		klog.Fatal("Remove filter: 'fields' cannot be empty")
	}

	plugin := &RemoveFilter{
		config:         config,
		fieldsDeleters: make([]field_deleter.FieldDeleter, 0),
	}

	// Create field deleters
	for _, field := range removeConfig.Fields {
		plugin.fieldsDeleters = append(plugin.fieldsDeleters, field_deleter.NewFieldDeleter(field))
	}

	return plugin
}

func (plugin *RemoveFilter) Filter(event map[string]any) (map[string]any, bool) {
	for _, d := range plugin.fieldsDeleters {
		d.Delete(event)
	}
	return event, true
}
