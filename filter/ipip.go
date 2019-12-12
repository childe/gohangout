package filter

import (
	"strconv"

	"github.com/childe/gohangout/value_render"
	"github.com/golang/glog"
	datx "github.com/ipipdotnet/datx-go"
)

type IPIPFilter struct {
	config    map[interface{}]interface{}
	src       string
	srcVR     value_render.ValueRender
	target    string
	database  string
	city      *datx.City
	overwrite bool
}

func (l *MethodLibrary) NewIPIPFilter(config map[interface{}]interface{}) *IPIPFilter {
	plugin := &IPIPFilter{
		config:    config,
		target:    "geoip",
		overwrite: true,
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

func (plugin *IPIPFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
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
		event["country_name"] = a[0]
		event["province_name"] = a[1]
		event["city_name"] = a[2]
		if len(a) >= 10 {
			latitude, _ := strconv.ParseFloat(a[5], 10)
			longitude, _ := strconv.ParseFloat(a[6], 10)
			event["latitude"] = latitude
			event["longitude"] = longitude
			event["location"] = []interface{}{longitude, latitude}
			event["country_code"] = a[11]
		}
	} else {
		target := make(map[string]interface{})
		target["country_name"] = a[0]
		target["province_name"] = a[1]
		target["city_name"] = a[2]
		if len(a) >= 10 {
			latitude, _ := strconv.ParseFloat(a[5], 10)
			longitude, _ := strconv.ParseFloat(a[6], 10)
			target["latitude"] = latitude
			target["longitude"] = longitude
			target["location"] = []interface{}{longitude, latitude}
			target["country_code"] = a[11]
		}
		event[plugin.target] = target
	}
	return event, true
}
