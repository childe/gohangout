package condition_filter

import (
	"testing"
	"time"
)

func TestEQJsonpathSyntaxError(t *testing.T) {
	condition := `EQ($.name.first,jia) && EQ($.name.last,liu)`
	_, err := parseBoolTree(condition)
	if err == nil {
		t.Errorf("%s should have error", condition)
	}
}

func TestEQJsonpathSingleCondition(t *testing.T) {
	condition := `EQ($.name.first,"jia")`
	root, err := parseBoolTree(condition)
	if err != nil {
		t.Errorf("parse %s error", condition)
	}

	event := make(map[string]interface{})
	event["name"] = map[string]interface{}{"first": "jia", "last": "liu"}
	pass := root.Pass(event)
	if !pass {
		t.Errorf("pass failed. `%s` %#v", condition, event)
	}
}

func TestEQJsonpath(t *testing.T) {
	condition := `EQ($.name.first,"jia") && EQ($.name.last,"liu")`
	root, err := parseBoolTree(condition)
	if err != nil {
		t.Errorf("parse %s error", condition)
	}

	event := make(map[string]interface{})
	event["name"] = map[string]interface{}{"first": "jia", "last": "liu"}
	pass := root.Pass(event)
	if !pass {
		t.Errorf("pass failed. `%s` %#v", condition, event)
	}
}

func TestHasPrefixJsonpath(t *testing.T) {
	condition := `HasPrefix($.name.first,"jia") || HasPrefix($.name.last,"liu")`
	root, err := parseBoolTree(condition)
	if err != nil {
		t.Errorf("parse %s error", condition)
	}

	event := make(map[string]interface{})
	event["name"] = map[string]interface{}{"first": "ji", "last": "liuu"}
	pass := root.Pass(event)
	if !pass {
		t.Errorf("pass failed. `%s` %#v", condition, event)
	}
}

func TestHasSuffixJsonpath(t *testing.T) {
	condition := `HasSuffix($.name.first,"jia") || HasSuffix($.name.last,"liu")`
	root, err := parseBoolTree(condition)
	if err != nil {
		t.Errorf("parse %s error", condition)
	}

	event := make(map[string]interface{})
	event["name"] = map[string]interface{}{"first": "ji", "last": "uliu"}
	pass := root.Pass(event)
	if !pass {
		t.Errorf("pass failed. `%s` %#v", condition, event)
	}
}

func TestMatchJsonpath(t *testing.T) {
	condition := `Match($.name.first,"^jia$") && Match($.fullname,"^liu,jia$")`
	root, err := parseBoolTree(condition)
	if err != nil {
		t.Errorf("parse %s error", condition)
	}

	event := make(map[string]interface{})
	event["name"] = map[string]interface{}{"first": "jia", "last": "liu"}
	event["fullname"] = "liu,jia"
	pass := root.Pass(event)
	if !pass {
		t.Errorf("pass failed. `%s` %#v", condition, event)
	}
}

func TestContainsJsonpath(t *testing.T) {
	condition := `Contains($.name.first,"jia") || Contains($.name.last,"liu")`
	root, err := parseBoolTree(condition)
	if err != nil {
		t.Errorf("parse %s error", condition)
	}

	event := make(map[string]interface{})
	event["name"] = map[string]interface{}{"first": "ji", "last": "uliu"}
	pass := root.Pass(event)
	if !pass {
		t.Errorf("pass failed. `%s` %#v", condition, event)
	}
}

func TestNotBeforeAnd(t *testing.T) {
	condition := `EQ(name,first,"jia") ! && EQ(name,last,"liu")`
	_, err := parseBoolTree(condition)
	if err == nil {
		t.Errorf("parse %s should has error", condition)
	}
}

func TestSuccessiveAnd(t *testing.T) {
	condition := `EQ(name,first,"jia") && && EQ(name,last,"liu")`
	_, err := parseBoolTree(condition)
	if err == nil {
		t.Errorf("parse %s should has error", condition)
	}
}

func TestSuccessiveNot(t *testing.T) {
	condition := `EQ(name,first,"jia") && !!EQ(name,last,"liu")`
	root, err := parseBoolTree(condition)
	if err != nil {
		t.Errorf("parse %s error: %s", condition, err)
	}

	event := make(map[string]interface{})
	event["name"] = map[string]interface{}{"first": "jia", "last": "liu"}
	pass := root.Pass(event)
	if !pass {
		t.Errorf("`%s` %#v", condition, event)
	}

	event = make(map[string]interface{})
	event["name"] = map[string]interface{}{"first": "jia", "last": "XXX"}
	pass = root.Pass(event)
	if pass {
		t.Errorf("`%s` %#v", condition, event)
	}
}

func TestComplexCondition(t *testing.T) {
	var (
		condition string
		root      *OPNode
		err       error
		event     map[string]interface{}
		pass      bool
	)
	condition = `(EQ(namespace,"elasticsearch") && EQ(kubernetes.container.name,"nginx")) || (EQ(namespace,"kibana") && EQ(kubernetes.container.name,"nginx-100014379"))`
	root, err = parseBoolTree(condition)
	if err != nil || root == nil {
		t.Error("parse error")
	}

	// pass
	event = make(map[string]interface{})
	event["namespace"] = "elasticsearch"
	event["kubernetes.container.name"] = "nginx"

	pass = root.Pass(event)
	if !pass {
		t.Error("")
	}

	// not pass
	event = make(map[string]interface{})
	event["namespace"] = "elasticsearch"
	event["kubernetes.container.name"] = "nginx-100014379"

	pass = root.Pass(event)
	if pass {
		t.Error("")
	}
}

func TestParseCondition(t *testing.T) {
	var (
		condition string
		root      *OPNode
		err       error
		event     map[string]interface{}
		pass      bool
	)

	config := make(map[interface{}]interface{})
	conditions := make([]interface{}, 3)
	conditions[0] = "{{if .name}}y{{end}}"
	conditions[1] = "{{if .name.first}}y{{end}}"
	conditions[2] = `{{if eq .name.first "dehua"}}y{{end}}`
	config["if"] = conditions
	f := NewConditionFilter(config)

	// should drop
	event = make(map[string]interface{})
	event["@timestamp"] = time.Now().Unix()
	event["name"] = map[string]interface{}{"first": "dehua"}

	pass = f.Pass(event)

	if pass == false {
		t.Error("should pass the conditions")
	}

	// should not drop
	event = make(map[string]interface{})
	event["@timestamp"] = time.Now().Unix()
	event["name"] = map[string]interface{}{"last": "liu"}

	pass = f.Pass(event)

	if pass == true {
		t.Error("should not pass the conditions")
	}

	// Single Condition
	condition = `EQ(name,first,"jia")`
	root, err = parseBoolTree(condition)
	if err != nil {
		t.Errorf("parse %s error: %s", condition, err)
	}

	event = make(map[string]interface{})
	event["name"] = map[string]interface{}{"first": "jia"}
	pass = root.Pass(event)
	if !pass {
		t.Errorf("`%s` %#v", condition, event)
	}

	condition = `Match(user,name,^liu.*a$)`
	root, err = parseBoolTree(condition)
	if err != nil {
		t.Errorf("parse %s error: %s", condition, err)
	}

	event = make(map[string]interface{})
	event["user"] = map[string]interface{}{"name": "liujia"}
	pass = root.Pass(event)
	if !pass {
		t.Errorf("`%s` %#v", condition, event)
	}

	event = make(map[string]interface{})
	event["user"] = map[string]interface{}{"name": "lujia"}
	pass = root.Pass(event)
	if pass {
		t.Errorf("`%s` %#v", condition, event)
	}

	// nil value

	condition = `Contains(name,jia)`
	root, err = parseBoolTree(condition)
	if err != nil {
		t.Errorf("parse %s error: %s", condition, err)
	}

	event = make(map[string]interface{})
	event["name"] = "liujia"
	pass = root.Pass(event)
	if !pass {
		t.Errorf("`%s` %#v", condition, event)
	}

	event = make(map[string]interface{})
	event["name"] = nil
	pass = root.Pass(event)
	if pass {
		t.Errorf("`%s` %#v", condition, event)
	}

	condition = `Contains(name,first,jia)`
	root, err = parseBoolTree(condition)
	if err != nil {
		t.Errorf("parse %s error: %s", condition, err)
	}

	event = make(map[string]interface{})
	event["name"] = "liujia"
	pass = root.Pass(event)
	if pass {
		t.Errorf("`%s` %#v", condition, event)
	}

	// ! Condition
	condition = `!EQ(name,first,"jia")`
	root, err = parseBoolTree(condition)
	if err != nil {
		t.Errorf("parse %s error: %s", condition, err)
	}

	event = make(map[string]interface{})
	event["name"] = map[string]interface{}{"first": "XX"}
	pass = root.Pass(event)
	if !pass {
		t.Errorf("`%s` %#v", condition, event)
	}

	// && Condition
	condition = `EQ(name,first,"jia") && EQ(name,last,"liu")`
	root, err = parseBoolTree(condition)
	if err != nil {
		t.Errorf("parse %s error: %s", condition, err)
	}

	event = make(map[string]interface{})
	event["name"] = map[string]interface{}{"first": "jia", "last": "liu"}
	pass = root.Pass(event)
	if !pass {
		t.Errorf("`%s` %#v", condition, event)
	}

	// combinin conditions

	condition = `!Exist(source) && (EQ(path,"/var/log/secure") || EQ(path,"/var/log/messages"))`
	root, err = parseBoolTree(condition)
	if err != nil {
		t.Errorf("parse %s error: %s", condition, err)
	}

	event = make(map[string]interface{})
	event["path"] = "/var/log/messages"
	pass = root.Pass(event)
	if !pass {
		t.Errorf("`%s` %#v", condition, event)
	}

	// parse blank before !

	condition = `EQ(name,first,"jia") && !EQ(name,last,"liu")`
	root, err = parseBoolTree(condition)
	if err != nil {
		t.Errorf("parse %s error: %s", condition, err)
	}

	event = make(map[string]interface{})
	event["name"] = map[string]interface{}{"first": "jia", "last": "liu"}
	pass = root.Pass(event)
	if pass {
		t.Errorf("`%s` %#v", condition, event)
	}

	event = make(map[string]interface{})
	event["name"] = map[string]interface{}{"first": "jia", "last": "XXX"}
	pass = root.Pass(event)
	if !pass {
		t.Errorf("`%s` %#v", condition, event)
	}
	// parse error

	// successive condition (no && || between them)
	condition = `EQ(name,first,"jia") EQ(name,last,"liu")`
	_, err = parseBoolTree(condition)
	if err == nil {
		t.Errorf("parse %s should has error", condition)
	}

	// single &
	condition = `EQ(name,first,"jia") & EQ(name,last,"liu")`
	_, err = parseBoolTree(condition)
	if err == nil {
		t.Errorf("parse %s should has error", condition)
	}

	// 3 &
	condition = `EQ(name,first,"jia") &&& EQ(name,last,"liu")`
	_, err = parseBoolTree(condition)
	if err == nil {
		t.Errorf("parse %s should has error", condition)
	}

	// unclose ()
	condition = `EQ(name,first,"jia" && EQ(name,last,"liu")`
	_, err = parseBoolTree(condition)
	if err == nil {
		t.Errorf("parse %s should has error", condition)
	}

	// unclose ""
	condition = `EQ(name,first,"jia") && EQ(name,last,"liu)`
	_, err = parseBoolTree(condition)
	if err == nil {
		t.Errorf("parse %s should has error", condition)
	}

	// ( in "" this is correct
	condition = `EQ(name,first,"ji()a") && EQ(name,last,"liu")`
	_, err = parseBoolTree(condition)
	if err != nil {
		t.Errorf("parse %s error: %s", condition, err)
	}

	// || condition
	condition = `EQ(name,first,"jia") || EQ(name,last,"liu")`
	root, err = parseBoolTree(condition)
	if err != nil {
		t.Errorf("parse %s error: %s", condition, err)
	}

	event = make(map[string]interface{})
	event["name"] = map[string]interface{}{"first": "jia", "last": "XXX"}
	pass = root.Pass(event)
	if !pass {
		t.Errorf("`%s` %#v", condition, event)
	}

	event = make(map[string]interface{})
	event["name"] = map[string]interface{}{"first": "XXX", "last": "liu"}
	pass = root.Pass(event)
	if !pass {
		t.Errorf("`%s` %#v", condition, event)
	}

	// complex condition
	condition = `!Exist(via) || !EQ(via,"ak")`
	root, err = parseBoolTree(condition)
	if err != nil {
		t.Errorf("parse %s error: %s", condition, err)
	}

	event = make(map[string]interface{})
	event["via"] = "abc"
	pass = root.Pass(event)
	if !pass {
		t.Errorf("`%s` %#v", condition, event)
	}

	event = make(map[string]interface{})
	event["XXX"] = "ak"
	pass = root.Pass(event)
	if !pass {
		t.Errorf("`%s` %#v", condition, event)
	}

	event = make(map[string]interface{})
	event["via"] = "ak"
	pass = root.Pass(event)
	if pass {
		t.Errorf("`%s` %#v", condition, event)
	}

	// ()
	condition = `Exist(a) && (Exist(b) || Exist(c))`
	root, err = parseBoolTree(condition)
	if err != nil {
		t.Errorf("parse %s error: %s", condition, err)
	}

	event = map[string]interface{}{"a": "", "b": "", "c": ""}
	pass = root.Pass(event)
	if !pass {
		t.Errorf("`%s` %#v", condition, event)
	}

	event = map[string]interface{}{"a": "", "b": ""}
	pass = root.Pass(event)
	if !pass {
		t.Errorf("`%s` %#v", condition, event)
	}

	event = map[string]interface{}{"a": "", "c": ""}
	pass = root.Pass(event)
	if !pass {
		t.Errorf("`%s` %#v", condition, event)
	}

	event = map[string]interface{}{"b": "", "c": ""}
	pass = root.Pass(event)
	if pass {
		t.Errorf("`%s` %#v", condition, event)
	}

	event = map[string]interface{}{"a": ""}
	pass = root.Pass(event)
	if pass {
		t.Errorf("`%s` %#v", condition, event)
	}

	// outsides
	condition = `Before(-24h) || After(24h)`
	root, err = parseBoolTree(condition)
	if err != nil {
		t.Errorf("parse %s error: %s", condition, err)
	}

	event = make(map[string]interface{})
	event["@timestamp"] = time.Now()
	pass = root.Pass(event)
	if pass {
		t.Errorf("`%s` %#v", condition, event)
	}

	event = make(map[string]interface{})
	event["@timestamp"] = time.Now().Add(time.Duration(time.Second * 86500))
	pass = root.Pass(event)
	if !pass {
		t.Errorf("`%s` %#v", condition, event)
	}

	// between
	condition = `Before(24h) && After(-24h)`
	root, err = parseBoolTree(condition)
	if err != nil {
		t.Errorf("parse %s error: %s", condition, err)
	}

	event = make(map[string]interface{})
	event["@timestamp"] = time.Now()
	pass = root.Pass(event)
	if !pass {
		t.Errorf("`%s` %#v", condition, event)
	}
	event = make(map[string]interface{})
	event["@timestamp"] = time.Now().Add(time.Duration(time.Second * -86500))
	pass = root.Pass(event)
	if pass {
		t.Errorf("`%s` %#v", condition, event)
	}

	// """"
	condition = `!Exist(via) || !EQ(via,""ak"")`
	root, err = parseBoolTree(condition)
	if err != nil {
		t.Errorf("parse %s error: %s", condition, err)
	}

	event = make(map[string]interface{})
	event["via"] = `"ak"`
	pass = root.Pass(event)
	if pass {
		t.Errorf("`%s` %#v", condition, event)
	}
	event = make(map[string]interface{})
	event["via"] = `ak`
	pass = root.Pass(event)
	if !pass {
		t.Errorf("`%s` %#v", condition, event)
	}
}
