package filter

import (
	"github.com/childe/gohangout/field_deleter"
	"github.com/golang/glog"
)

type RemoveFilter struct {
	config         map[interface{}]interface{}
	fieldsDeleters []field_deleter.FieldDeleter
}

func (l *MethodLibrary) NewRemoveFilter(config map[interface{}]interface{}) *RemoveFilter {
	plugin := &RemoveFilter{
		config:         config,
		fieldsDeleters: make([]field_deleter.FieldDeleter, 0),
	}

	if fieldsValue, ok := config["fields"]; ok {
		for _, field := range fieldsValue.([]interface{}) {
			plugin.fieldsDeleters = append(plugin.fieldsDeleters, field_deleter.NewFieldDeleter(field.(string)))
		}
	} else {
		glog.Fatal("fileds must be set in remove filter plugin")
	}
	return plugin
}

func (plugin *RemoveFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	for _, d := range plugin.fieldsDeleters {
		d.Delete(event)
	}
	return event, true
}
