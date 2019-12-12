package filter

import "github.com/golang/glog"

type RenameFilter struct {
	config map[interface{}]interface{}
	fields map[string]string
}

func (l *MethodLibrary) NewRenameFilter(config map[interface{}]interface{}) *RenameFilter {
	plugin := &RenameFilter{
		config: config,
		fields: make(map[string]string),
	}

	if fieldsValue, ok := config["fields"]; ok {
		for k, v := range fieldsValue.(map[interface{}]interface{}) {
			plugin.fields[k.(string)] = v.(string)
		}
	} else {
		glog.Fatal("fileds must be set in rename filter plugin")
	}
	return plugin
}

func (plugin *RenameFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	for source, target := range plugin.fields {
		if v, ok := event[source]; ok {
			event[target] = v
			delete(event, source)
		}
	}
	return event, true
}
