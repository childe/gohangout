package filter

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ConfigError represents a detailed configuration error with context
type ConfigError struct {
	PluginType string
	Field      string
	Message    string
	Cause      error
}

func (e *ConfigError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("[%s filter] %s: %s", e.PluginType, e.Field, e.Message)
	}
	return fmt.Sprintf("[%s filter] %s", e.PluginType, e.Message)
}

func (e *ConfigError) Unwrap() error {
	return e.Cause
}

// ConvertToJSONCompatible recursively converts map[any]any to map[string]any for JSON compatibility
// Returns an error if a config key is not a string
func ConvertToJSONCompatible(input any) (any, error) {
	switch v := input.(type) {
	case map[any]any:
		result := make(map[string]any)
		for k, val := range v {
			if keyStr, ok := k.(string); ok {
				convertedVal, err := ConvertToJSONCompatible(val)
				if err != nil {
					return nil, err
				}
				result[keyStr] = convertedVal
			} else {
				return nil, fmt.Errorf("config key '%v' is not a string", k)
			}
		}
		return result, nil
	case []any:
		result := make([]any, len(v))
		for i, val := range v {
			convertedVal, err := ConvertToJSONCompatible(val)
			if err != nil {
				return nil, err
			}
			result[i] = convertedVal
		}
		return result, nil
	default:
		return v, nil
	}
}

// MustConvertToJSONCompatible is the legacy version that panics on error
// Deprecated: Use ConvertToJSONCompatible instead
func MustConvertToJSONCompatible(input any) any {
	result, err := ConvertToJSONCompatible(input)
	if err != nil {
		panic(err)
	}
	return result
}

// DecodeConfig decodes filter configuration using encoding/json
// Returns detailed error messages with plugin context
func DecodeConfig(filterType string, config map[any]any, result any) error {
	// Convert config to JSON-serializable format
	jsonConfig, err := ConvertToJSONCompatible(config)
	if err != nil {
		return &ConfigError{
			PluginType: filterType,
			Message:    fmt.Sprintf("failed to convert config: %v", err),
			Cause:      err,
		}
	}

	// Convert map to JSON and then unmarshal to struct
	jsonBytes, err := json.Marshal(jsonConfig)
	if err != nil {
		return &ConfigError{
			PluginType: filterType,
			Message:    fmt.Sprintf("failed to marshal config to JSON: %v", err),
			Cause:      err,
		}
	}

	// Use a decoder with UseNumber to preserve number precision and allow type flexibility
	decoder := json.NewDecoder(strings.NewReader(string(jsonBytes)))
	decoder.UseNumber()

	if err := decoder.Decode(result); err != nil {
		return &ConfigError{
			PluginType: filterType,
			Message:    fmt.Sprintf("configuration parsing error: %v", err),
			Cause:      err,
		}
	}

	return nil
}

// SafeDecodeConfig safely decodes filter configuration using encoding/json
// This provides detailed error messages from the standard library
// Deprecated: Use DecodeConfig which returns error instead of panicking
func SafeDecodeConfig(filterType string, config map[any]any, result any) {
	if err := DecodeConfig(filterType, config, result); err != nil {
		panic(err.Error())
	}
}

// ValidateRequiredFields checks that required fields are present in the decoded config
// Returns an error if any required field is missing or empty
func ValidateRequiredFields(filterType string, fields map[string]any) error {
	for fieldName, fieldValue := range fields {
		if fieldValue == nil || (fmt.Sprintf("%v", fieldValue) == "") {
			return &ConfigError{
				PluginType: filterType,
				Field:      fieldName,
				Message:    "field is required",
			}
		}
	}
	return nil
}

// MustValidateRequiredFields is the legacy version that panics on error
// Deprecated: Use ValidateRequiredFields instead
func MustValidateRequiredFields(filterType string, fields map[string]any) {
	if err := ValidateRequiredFields(filterType, fields); err != nil {
		panic(err.Error())
	}
}
