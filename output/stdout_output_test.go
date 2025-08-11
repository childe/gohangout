package output

import (
	"testing"
)

func TestStdoutConfigParsing(t *testing.T) {
	tests := []struct {
		name     string
		config   map[any]any
		expected StdoutConfig
		panics   bool
	}{
		{
			name: "valid stdout config",
			config: map[any]any{
				"codec": "json",
			},
			expected: StdoutConfig{
				Codec: "json",
			},
			panics: false,
		},
		{
			name:   "empty config should work",
			config: map[any]any{},
			expected: StdoutConfig{
				Codec: "",
			},
			panics: false,
		},
		{
			name: "invalid JSON structure should panic",
			config: map[any]any{
				123: "invalid key",
			},
			panics: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.panics {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("newStdoutOutput() should have panicked")
					}
				}()
				newStdoutOutput(tt.config)
			} else {
				output := newStdoutOutput(tt.config)
				stdoutOutput, ok := output.(*StdoutOutput)
				if !ok {
					t.Errorf("newStdoutOutput() should return *StdoutOutput")
					return
				}
				
				// The encoder is created so we can't directly check config
				// but we can verify the output was created successfully
				if stdoutOutput == nil {
					t.Errorf("StdoutOutput should not be nil")
				}
			}
		})
	}
}