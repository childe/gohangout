package filter

import "testing"

func TestGrokFilter(t *testing.T) {
	config := make(map[interface{}]interface{})
	match := make([]interface{}, 2)
	match[0] = `(?P<logtime>\S+ \S+) \[(?P<level>\w+)\] (?P<msg>.*)$`
	match[1] = `(?P<logtime>\S+ \S+)`
	config["match"] = match
	config["src"] = "message"

	f := methodLibrary.NewGrokFilter(config)

	event := make(map[string]interface{})
	event["message"] = "2018-07-12T14:45:00 +0800 [info] message"

	event, ok := f.Filter(event)
	t.Log(event)

	if ok == false {
		t.Error("grok filter fail")
	}

	if v, ok := event["msg"]; !ok {
		t.Error("msg field should exist")
	} else {
		if v != "message" {
			t.Error("msg field do not match")
		}
	}
}
func TestTarget(t *testing.T) {
	config := make(map[interface{}]interface{})
	match := make([]interface{}, 2)
	match[0] = `(?P<logtime>\S+ \S+) \[(?P<level>\w+)\] (?P<msg>.*)$`
	match[1] = `(?P<logtime>\S+ \S+)`
	config["match"] = match
	config["src"] = "message"
	config["target"] = "grok"

	f := methodLibrary.NewGrokFilter(config)

	event := make(map[string]interface{})
	event["message"] = "2018-07-12T14:45:00 +0800 [info] message"

	event, ok := f.Filter(event)
	t.Log(event)

	if ok == false {
		t.Error("grok filter fail")
	}

	if grok, ok := event["grok"]; ok {
		if msg, ok := grok.(map[string]interface{})["msg"]; !ok || msg.(string) != "message" {
			t.Error("msg field do not match")
		}
	} else {
		t.Error("grok field should exist")
	}
}

func TestPattern(t *testing.T) {
	grok := &Grok{}
	grok.patterns = map[string]string{
		"USERNAME": "[a-zA-Z0-9._-]+",
		"USER":     "%{USERNAME}",
		"IPV6":     `((([0-9A-Fa-f]{1,4}:){7}([0-9A-Fa-f]{1,4}|:))|(([0-9A-Fa-f]{1,4}:){6}(:[0-9A-Fa-f]{1,4}|((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){5}(((:[0-9A-Fa-f]{1,4}){1,2})|:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){4}(((:[0-9A-Fa-f]{1,4}){1,3})|((:[0-9A-Fa-f]{1,4})?:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){3}(((:[0-9A-Fa-f]{1,4}){1,4})|((:[0-9A-Fa-f]{1,4}){0,2}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){2}(((:[0-9A-Fa-f]{1,4}){1,5})|((:[0-9A-Fa-f]{1,4}){0,3}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){1}(((:[0-9A-Fa-f]{1,4}){1,6})|((:[0-9A-Fa-f]{1,4}){0,4}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(:(((:[0-9A-Fa-f]{1,4}){1,7})|((:[0-9A-Fa-f]{1,4}){0,5}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:)))(%.+)?`,
		"IPV4":     `(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)`,
		"IP":       "(?:%{IPV6}|%{IPV4})",
	}

	p := grok.translateMatchPattern(`%{USER:user}`)
	if p != `(?P<user>([a-zA-Z0-9._-]+))` {
		t.Error(p)
	}

	config := make(map[interface{}]interface{})
	match := make([]interface{}, 1)
	match[0] = grok.translateMatchPattern(`^%{IP:ip} %{USER:user} \[(?P<loglevel>\w+)\] (?P<msg>.*)`)
	config["match"] = match
	config["src"] = "message"

	f := methodLibrary.NewGrokFilter(config)

	event := make(map[string]interface{})
	event["message"] = "10.10.10.255 childe [info] message"

	event, ok := f.Filter(event)
	t.Log(event)

	if ok == false {
		t.Error("grok filter fail")
	}

	if v, ok := event["loglevel"]; !ok {
		t.Error("loglevel field should exist")
	} else {
		if v != "info" {
			t.Error("loglevel field do not match")
		}
	}

	if v, ok := event["user"]; !ok {
		t.Error("user field should exist")
	} else {
		if v != "childe" {
			t.Error("user field do not match")
		}
	}

	if v, ok := event["ip"]; !ok {
		t.Error("ip field should exist")
	} else {
		if v != "10.10.10.255" {
			t.Error("ip field do not match")
		}
	}

	if v, ok := event["msg"]; !ok {
		t.Error("msg field should exist")
	} else {
		if v != "message" {
			t.Error("msg field do not match")
		}
	}
}
