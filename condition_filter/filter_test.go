package condition_filter

import (
	"testing"
	"time"
)

func TestParseCondition(t *testing.T) {
	var (
		condition string
		root      *OPNode
		err       error
		event     map[string]interface{}
		pass      bool
	)

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

	// successive !
	condition = `EQ(name,first,"jia") && !!EQ(name,last,"liu")`
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

	event = make(map[string]interface{})
	event["name"] = map[string]interface{}{"first": "jia", "last": "XXX"}
	pass = root.Pass(event)
	if pass {
		t.Errorf("`%s` %#v", condition, event)
	}

	// parse error

	// successive condition (no && || between them)
	condition = `EQ(name,first,"jia") EQ(name,last,"liu")`
	_, err = parseBoolTree(condition)
	if err == nil {
		t.Errorf("parse %s should has error: %s", condition)
	}

	// single &
	condition = `EQ(name,first,"jia") & EQ(name,last,"liu")`
	_, err = parseBoolTree(condition)
	if err == nil {
		t.Errorf("parse %s should has error: %s", condition)
	}

	// 3 &
	condition = `EQ(name,first,"jia") &&& EQ(name,last,"liu")`
	_, err = parseBoolTree(condition)
	if err == nil {
		t.Errorf("parse %s should has error: %s", condition)
	}

	// unclose ()
	condition = `EQ(name,first,"jia" && EQ(name,last,"liu")`
	_, err = parseBoolTree(condition)
	if err == nil {
		t.Errorf("parse %s should has error: %s", condition)
	}

	// unclose ""
	condition = `EQ(name,first,"jia") && EQ(name,last,"liu)`
	_, err = parseBoolTree(condition)
	if err == nil {
		t.Errorf("parse %s should has error: %s", condition)
	}

	// ! before &&
	condition = `EQ(name,first,"jia") ! && EQ(name,last,"liu")`
	_, err = parseBoolTree(condition)
	if err == nil {
		t.Errorf("parse %s should has error: %s", condition)
	}

	// successive &&
	condition = `EQ(name,first,"jia") && && EQ(name,last,"liu")`
	_, err = parseBoolTree(condition)
	if err == nil {
		t.Errorf("parse %s should has error: %s", condition)
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
	condition = `!Exist(via) || !EQ(via,"akamai")`
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
	event["XXX"] = "akamai"
	pass = root.Pass(event)
	if !pass {
		t.Errorf("`%s` %#v", condition, event)
	}

	event = make(map[string]interface{})
	event["via"] = "akamai"
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

}
