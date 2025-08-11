package codec

import (
	"testing"
)

func TestFormatEncoder(t *testing.T) {
	cases := []struct {
		codec    string
		event    map[string]any
		expected string
	}{
		{
			codec:    "format:[msg]",
			event:    map[string]any{"msg": "this is a line"},
			expected: "this is a line",
		},
		{
			codec:    "format:msg",
			event:    map[string]any{"msg": "this is a line"},
			expected: "msg",
		},
		{
			codec:    "format:my name is %{name}",
			event:    map[string]any{"name": "childe"},
			expected: "my name is childe",
		},
		{
			codec:    "format:my name is $.name.firstname",
			event:    map[string]any{"name": map[string]string{"firstname": "jia"}},
			expected: "my name is $.name.firstname",
		},
		{
			codec:    "format:$.name.firstname",
			event:    map[string]any{"name": map[string]string{"firstname": "jia"}},
			expected: "jia",
		},
		{
			codec:    "format:my name is {{.name.firstname}}",
			event:    map[string]any{"name": map[string]string{"firstname": "jia"}},
			expected: "my name is jia",
		},
	}

	for _, c := range cases {
		t.Logf("test %v", c.codec)
		encoder := NewEncoder(c.codec)
		got, err := encoder.Encode(c.event)
		if err != nil {
			t.Errorf("get error from `%s`: %v", c.codec, err)
			continue
		}
		if string(got) != c.expected {
			t.Errorf("expected `%s`, got `%s`", c.expected, got)
		}
	}
}
