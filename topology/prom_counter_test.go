package topology

import "testing"

func TestGetPromCounter(t *testing.T) {
	type TestCase struct {
		config map[interface{}]interface{}
		want   bool
	}

	for _, c := range []TestCase{
		{
			config: nil,
			want:   false,
		},
		{
			config: map[interface{}]interface{}{"prometheus_counter": "test"},
			want:   false,
		},
		{
			config: map[interface{}]interface{}{"prometheus_counter": map[string]string{
				"name":      "gohangout_add_filter",
				"namespace": "rack_a",
				"help":      "rack_a gohangout add filter counter",
			}},
			want: true,
		},
		{
			config: map[interface{}]interface{}{"prometheus_counter": map[string]string{
				"name":      "gohangout_add_filter",
				"namespace": "rack_a",
				"help":      "rack_a gohangout add filter counter",
			}},
			want: true,
		},
		{
			config: map[interface{}]interface{}{"prometheus_counter": map[string]string{
				"name":      "gohangout_add_filter",
				"namespace": "rack_a",
				"help":      "xxxxxxxxxxx",
			}},
			want: true,
		},
		{
			config: map[interface{}]interface{}{"prometheus_counter": map[string]string{
				"name":      "gohangout_raname_filter",
				"namespace": "rack_a",
				"help":      "rack_a gohangout add filter counter",
			}},
			want: true,
		},
	} {
		counter := GetPromCounter(c.config)
		if (counter != nil) != c.want {
			t.Errorf("GetPromCounter(%v) = %v, want %v", c.config, counter != nil, c.want)
		}
	}

	if len(counterManager) != 2 {
		t.Errorf("len(counterManager) = %v, want %v", len(counterManager), 2)
	}
}
