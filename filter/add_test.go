package filter

import (
	"strings"
	"testing"
	"time"
)

func TestAddFilter(t *testing.T) {
	config := make(map[any]any)
	fields := make(map[any]any)
	fields["name"] = `{{.first}} {{.last}}`
	fields["firstname"] = `$.first`
	config["fields"] = fields
	f := BuildFilter("Add", config)

	event := make(map[string]any)
	event["@timestamp"] = time.Now().Unix()
	event["first"] = "dehua"
	event["last"] = "liu"
	t.Log(event)

	event, ok := f.Filter(event)
	t.Log(event)

	if ok == false {
		t.Error("add filter fail")
	}

	name, ok := event["name"]
	if ok == false {
		t.Error("add filter should add `name` field")
	}
	if name != "dehua liu" {
		t.Error("name field should be `dehua liu`")
	}

	firstname, ok := event["firstname"]
	if ok == false {
		t.Error("add filter should add `firstname` field")
	}
	if firstname != "dehua" {
		t.Error("firstname field should be `dehua`")
	}
}

func TestAddConfigParsing(t *testing.T) {
	tests := []struct {
		name        string
		config      map[any]any
		expectError bool
		errorSubstr string
	}{
		{
			name: "valid config",
			config: map[any]any{
				"fields": map[any]any{
					"name": "test",
					"type": "web",
				},
				"overwrite": true,
			},
			expectError: false,
		},
		{
			name: "missing fields",
			config: map[any]any{
				"overwrite": true,
			},
			expectError: true,
			errorSubstr: "fields' is required",
		},
		{
			name: "wrong type for fields - string instead of map",
			config: map[any]any{
				"fields": "this should be a map",
			},
			expectError: true,
			errorSubstr: "cannot parse 'fields'",
		},
		{
			name: "wrong type for overwrite - string instead of bool",
			config: map[any]any{
				"fields": map[any]any{
					"name": "test",
				},
				"overwrite": "not a boolean",
			},
			expectError: true,
			errorSubstr: "cannot parse 'overwrite'",
		},
		{
			name: "wrong type for field value - number instead of string",
			config: map[any]any{
				"fields": map[any]any{
					"name": "test",
					"age":  123, // this should be string
				},
			},
			expectError: true,
			errorSubstr: "cannot parse 'Fields[age]'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.expectError {
						t.Errorf("Expected no panic but got: %v", r)
					} else {
						errorStr := r.(string)
						if !strings.Contains(errorStr, tt.errorSubstr) {
							t.Errorf("Expected error to contain '%s', but got: %s", tt.errorSubstr, errorStr)
						}
					}
				} else {
					if tt.expectError {
						t.Errorf("Expected error but got none")
					}
				}
			}()

			newAddFilter(tt.config)
		})
	}
}
