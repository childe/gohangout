package input

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ConvertToJSONCompatible recursively converts map[any]any to map[string]any for JSON compatibility
func ConvertToJSONCompatible(input any) any {
	switch v := input.(type) {
	case map[any]any:
		result := make(map[string]any)
		for k, val := range v {
			if keyStr, ok := k.(string); ok {
				result[keyStr] = ConvertToJSONCompatible(val)
			} else {
				panic(fmt.Sprintf("config key '%v' is not a string", k))
			}
		}
		return result
	case []any:
		result := make([]any, len(v))
		for i, val := range v {
			result[i] = ConvertToJSONCompatible(val)
		}
		return result
	default:
		return v
	}
}

// SafeDecodeConfig safely decodes input configuration using encoding/json
// This provides detailed error messages from the standard library
func SafeDecodeConfig(inputType string, config map[any]any, result any) {
	// Convert config to JSON-serializable format
	jsonConfig := ConvertToJSONCompatible(config)

	// Convert map to JSON and then unmarshal to struct
	jsonBytes, err := json.Marshal(jsonConfig)
	if err != nil {
		panic(fmt.Sprintf("%s input: failed to marshal config to JSON: %v", inputType, err))
	}

	// Use a decoder with UseNumber to preserve number precision and allow type flexibility
	decoder := json.NewDecoder(strings.NewReader(string(jsonBytes)))
	decoder.UseNumber()
	
	if err := decoder.Decode(result); err != nil {
		panic(fmt.Sprintf("%s input configuration error: %v", inputType, err))
	}
}

// ValidateRequiredFields checks that required fields are present in the decoded config
func ValidateRequiredFields(inputType string, fields map[string]any) {
	for fieldName, fieldValue := range fields {
		if fieldValue == nil || (fmt.Sprintf("%v", fieldValue) == "") {
			panic(fmt.Sprintf("%s input: '%s' is required", inputType, fieldName))
		}
	}
}