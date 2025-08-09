package filter

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"k8s.io/klog/v2"
)

// SafeDecodeConfig safely decodes filter configuration using mapstructure
// This provides much better error messages than manual type assertions
func SafeDecodeConfig(filterType string, config map[any]any, result any) {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           result,
		ErrorUnused:      false,
	})
	if err != nil {
		klog.Fatalf("%s filter: failed to create config decoder: %v", filterType, err)
	}

	if err := decoder.Decode(config); err != nil {
		klog.Fatalf("%s filter configuration error: %v", filterType, err)
	}
}

// ValidateRequiredFields checks that required fields are present in the decoded config
func ValidateRequiredFields(filterType string, fields map[string]any) {
	for fieldName, fieldValue := range fields {
		if fieldValue == nil || (fmt.Sprintf("%v", fieldValue) == "") {
			klog.Fatalf("%s filter: '%s' is required", filterType, fieldName)
		}
	}
}

