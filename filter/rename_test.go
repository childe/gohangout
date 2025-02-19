package filter

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestRenameFilter(t *testing.T) {
	type testcase struct {
		config   map[interface{}]interface{}
		event    map[string]interface{}
		expected map[string]interface{}
	}

	testcases := []testcase{
		{
			config: map[interface{}]interface{}{
				"fields": map[interface{}]interface{}{
					"name1": "n1",
					"name2": "n2",
				},
			},
			event: map[string]interface{}{
				"name1": "liu",
				"name2": "dehua",
			},
			expected: map[string]interface{}{
				"n1": "liu",
				"n2": "dehua",
			},
		},
		{
			config: map[interface{}]interface{}{
				"fields": map[interface{}]interface{}{
					"[name][last]": "[name][first]",
				},
			},
			event: map[string]interface{}{
				"name": map[string]interface{}{
					"last": "liu",
				},
			},
			expected: map[string]interface{}{
				"name": map[string]interface{}{
					"first": "liu",
				},
			},
		},
		{
			config: map[interface{}]interface{}{
				"fields": map[interface{}]interface{}{
					"[name][last]": "[name][first]",
				},
			},
			event: map[string]interface{}{
				"name": map[string]interface{}{
					"last": nil,
				},
			},
			expected: map[string]interface{}{
				"name": map[string]interface{}{
					"first": nil,
				},
			},
		},
		{
			config: map[interface{}]interface{}{
				"fields": map[interface{}]interface{}{
					"[name][last]": "[name][first]",
				},
			},
			event: map[string]interface{}{
				"name": map[string]interface{}{
					"full": "dehua liu",
				},
			},
			expected: map[string]interface{}{
				"name": map[string]interface{}{
					"full": "dehua liu",
				},
			},
		},
	}
	convey.Convey("RenameFilter", t, func() {
		for _, tc := range testcases {
			f := BuildFilter("Rename", tc.config)
			event, ok := f.Filter(tc.event)
			if !ok {
				t.Error("RenameFilter error")
			}
			convey.So(event, convey.ShouldResemble, tc.expected)
		}
	})
}
