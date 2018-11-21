package condition_filter

import (
	"testing"
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
		t.Errorf("[name][first] should be jia")
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
		t.Errorf("[name][first] should not be jia")
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
		t.Errorf("")
	}

	// parse error

	condition = `EQ(name,first,"jia") & EQ(name,last,"liu")`
	_, err = parseBoolTree(condition)
	if err == nil {
		t.Errorf("parse %s shoule has error: %s", condition, err)
	}

	condition = `EQ(name,first,"jia" && EQ(name,last,"liu")`
	_, err = parseBoolTree(condition)
	if err == nil {
		t.Errorf("parse %s shoule has error: %s", condition, err)
	}

	condition = `EQ(name,first,"jia") && EQ(name,last,"liu)`
	_, err = parseBoolTree(condition)
	if err == nil {
		t.Errorf("parse %s shoule has error: %s", condition, err)
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
		t.Errorf("")
	}

	event = make(map[string]interface{})
	event["name"] = map[string]interface{}{"first": "XXX", "last": "liu"}
	pass = root.Pass(event)
	if !pass {
		t.Errorf("")
	}

	// complex condition
	condition = `!Exist(via) || !EQ(via,akamai)`
	root, err = parseBoolTree(condition)
	if err != nil {
		t.Errorf("parse %s error: %s", condition, err)
	}

	event = make(map[string]interface{})
	event["via"] = "abc"
	pass = root.Pass(event)
	if pass {
		t.Errorf("")
	}

}
