package filter

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestRenameFilter(t *testing.T) {
	type testcase struct {
		config   map[any]any
		event    map[string]any
		expected map[string]any
	}

	testcases := []testcase{
		{
			config: map[any]any{
				"fields": map[any]any{
					"name1": "n1",
					"name2": "n2",
				},
			},
			event: map[string]any{
				"name1": "liu",
				"name2": "dehua",
			},
			expected: map[string]any{
				"n1": "liu",
				"n2": "dehua",
			},
		},
		{
			config: map[any]any{
				"fields": map[any]any{
					"[name][last]": "[name][first]",
				},
			},
			event: map[string]any{
				"name": map[string]any{
					"last": "liu",
				},
			},
			expected: map[string]any{
				"name": map[string]any{
					"first": "liu",
				},
			},
		},
		{
			config: map[any]any{
				"fields": map[any]any{
					"[name][last]": "[name][first]",
				},
			},
			event: map[string]any{
				"name": map[string]any{
					"last": nil,
				},
			},
			expected: map[string]any{
				"name": map[string]any{
					"first": nil,
				},
			},
		},
		{
			config: map[any]any{
				"fields": map[any]any{
					"[name][last]": "[name][first]",
				},
			},
			event: map[string]any{
				"name": map[string]any{
					"full": "dehua liu",
				},
			},
			expected: map[string]any{
				"name": map[string]any{
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
