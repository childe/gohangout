package value_render

import (
	"testing"
)

func TestJsonpathRender(t *testing.T) {
	var event map[string]interface{}
	var template string
	var vr ValueRender

	event = make(map[string]interface{})
	event["msg"] = "this is msg line"

	template = "$.msg"

	vr = GetValueRender(template)
	value, err := vr.Render(event)
	t.Log(value)

	if err != nil {
		t.Errorf("err != nil")
	}

	if value != "this is msg line" {
		t.Errorf("%q != %q", value, "this is msg line")
	}
}
