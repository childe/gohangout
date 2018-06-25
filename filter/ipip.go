package filter

import (
	"github.com/childe/gohangout/value_render"
	"github.com/golang/glog"
	datx "github.com/ipipdotnet/datx-go"
)

type IPIPFilter struct {
	BaseFilter

	config    map[interface{}]interface{}
	src       string
	srcVR     value_render.ValueRender
	target    string
	database  string
	city      *datx.City
	overwrite bool
}

func NewIPIPFilter(config map[interface{}]interface{}) *IPIPFilter {
	plugin := &IPIPFilter{
		BaseFilter: NewBaseFilter(config),
		config:     config,
		target:     "",
		overwrite:  true,
	}

	if overwrite, ok := config["overwrite"]; ok {
		plugin.overwrite = overwrite.(bool)
	}

	if database, ok := config["database"]; ok {
		plugin.database = database.(string)
		city, err := datx.NewCity(plugin.database)
		if err != nil {
			glog.Fatalf("could not load %s: %s", plugin.database, err)
		} else {
			plugin.city = city
		}
	} else {
		glog.Fatal("database must be set in IPIP filter plugin")
	}

	if src, ok := config["src"]; ok {
		plugin.src = src.(string)
		plugin.srcVR = value_render.GetValueRender2(plugin.src)
	} else {
		glog.Fatal("src must be set in IPIP filter plugin")
	}

	if target, ok := config["target"]; ok {
		plugin.target = target.(string)
	}
	return plugin
}

func (plugin *IPIPFilter) Process(event map[string]interface{}) (map[string]interface{}, bool) {
	inputI := plugin.srcVR.Render(event)
	if inputI == nil {
		return event, false
	}

	a, err := plugin.city.Find(inputI.(string))
	if err != nil {
		glog.V(10).Infof("failed to find %s: %s", inputI.(string), err)
		return event, false
	}
	if plugin.target == "" {
		event["country"] = a[0]
		event["province"] = a[1]
		event["city"] = a[2]
		if len(a) >= 10 {
			event["latitude"] = a[5]
			event["longitude"] = a[6]
			event["cityCode"] = a[9]
			event["countryCode"] = a[11]
		}
	} else {
		target := make(map[string]string)
		target["country"] = a[0]
		target["province"] = a[1]
		target["city"] = a[2]
		if len(a) >= 10 {
			target["latitude"] = a[5]
			target["longitude"] = a[6]
			target["cityCode"] = a[9]
			target["countryCode"] = a[11]
		}
		event[plugin.target] = target
	}
	return event, true
}
