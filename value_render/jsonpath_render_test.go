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
	value := vr.Render(event).(string)
	t.Log(value)

	if value != "this is msg line" {
		t.Errorf(value)
	}
}
