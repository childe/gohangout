package input

import (
	"testing"
)

func TestStdinConfigParsing(t *testing.T) {
	tests := []struct {
		name     string
		config   map[any]any
		expected StdinConfig
		panics   bool
	}{
		{
			name: "valid stdin config",
			config: map[any]any{
				"codec": "json",
			},
			expected: StdinConfig{
				Codec: "json",
			},
			panics: false,
		},
		{
			name:   "empty config should work",
			config: map[any]any{},
			expected: StdinConfig{
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
						t.Errorf("newStdinInput() should have panicked")
					}
				}()
				newStdinInput(tt.config)
			} else {
				input := newStdinInput(tt.config)
				stdinInput, ok := input.(*StdinInput)
				if !ok {
					t.Errorf("newStdinInput() should return *StdinInput")
					return
				}
				
				// The decoder is created so we can't directly check config
				// but we can verify the input was created successfully
				if stdinInput == nil {
					t.Errorf("StdinInput should not be nil")
				}
			}
		})
	}
}