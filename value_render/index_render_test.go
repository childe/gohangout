package value_render

import (
	"testing"
	"time"
)

func TestIndexRender(t *testing.T) {
	ts, _ := time.Parse("2006-01-02", "2022-03-04")
	for _, c := range []struct {
		event    map[string]interface{}
		template string
		want     string
	}{
		{
			event: map[string]interface{}{
				"@timestamp": ts,
			},
			template: "%{+2006.01.02}",
			want:     "2022.03.04",
		},
		{
			event: map[string]interface{}{
				"@timestamp": ts,
			},
			template: "app-%{+2006.01.02}",
			want:     "app-2022.03.04",
		},
		{
			event: map[string]interface{}{
				"@timestamp": ts,
			},
			template: "%{+2006.01.02}-log",
			want:     "2022.03.04-log",
		},
		{
			event: map[string]interface{}{
				"@timestamp": ts,
			},
			template: "app-%{+2006.01.02}-log",
			want:     "app-2022.03.04-log",
		},
		{
			event: map[string]interface{}{
				"topic":      "topic-one",
				"@timestamp": ts,
			},
			template: "app-%{topic}-%{+2006.01.02}-log",
			want:     "app-topic-one-2022.03.04-log",
		},
		{
			event: map[string]interface{}{
				"@timestamp": ts,
			},
			template: "app-%{topic}-%{+2006.01.02}-log",
			want:     "app-null-2022.03.04-log",
		},
		{
			event: map[string]interface{}{
				"@timestamp": ts,
			},
			template: "app-%{@metadata}{kafka}{topic}-%{+2006.01.02}-log",
			want:     "app-null-2022.03.04-log",
		},
		{
			event: map[string]interface{}{
				"@metadata":  nil,
				"@timestamp": ts,
			},
			template: "app-%{@metadata}{kafka}{topic}-%{+2006.01.02}-log",
			want:     "app-null-2022.03.04-log",
		},
		{
			event: map[string]interface{}{
				"@metadata": map[string]interface{}{
					"topic": "topic-one",
				},
				"@timestamp": ts,
			},
			template: "app-%{@metadata}{kafka}{topic}-%{+2006.01.02}-log",
			want:     "app-null-2022.03.04-log",
		},
		{
			event: map[string]interface{}{
				"@metadata": map[string]interface{}{
					"kafka": map[string]interface{}{
						"topic": "topic-one",
					},
				},
				"@timestamp": ts,
			},
			template: "app-%{@metadata}{kafka}{topic}-%{+2006.01.02}-log",
			want:     "app-topic-one-2022.03.04-log",
		},
	} {
		vr := NewIndexRender(c.template)
		got := vr.Render(c.event).(string)
		if c.want != got {
			t.Errorf("render %q, want %s, got %s", c.template, c.want, got)
		}
	}
	var event map[string]interface{}
	var template string
	var vr ValueRender
	var indexname string

	// timestamp exists, appid missing
	event = make(map[string]interface{})
	event["@timestamp"], _ = time.Parse("2006-01-02T15:04:05", "2019-03-04T14:21:00")

	template = "nginx-%{appid}-%{+2006.01.02}"

	vr = NewIndexRender(template)
	indexname = vr.Render(event).(string)
	t.Log(indexname)

	if indexname != "nginx-null-2019.03.04" {
		t.Errorf("%s != nginx-null-2019.03.04\n", indexname)
	}

}
