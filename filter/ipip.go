package filter

import (
	"strconv"
	"unsafe"

	"github.com/childe/gohangout/topology"
	"github.com/childe/gohangout/value_render"
	datx "github.com/ipipdotnet/datx-go"
	ipdb "github.com/ipipdotnet/ipdb-go"
	"k8s.io/klog/v2"
)

type IPIPFilter struct {
	config    map[interface{}]interface{}
	src       string
	srcVR     value_render.ValueRender
	target    string
	data_type string
	language  string
	database  string
	city      unsafe.Pointer
	overwrite bool
}

const bitSize int = 64

func init() {
	Register("IPIP", newIPIPFilter)
}

func newIPIPFilter(config map[interface{}]interface{}) topology.Filter {
	plugin := &IPIPFilter{
		config:    config,
		target:    "geoip",
		data_type: "datx",
		language:  "CN",
		overwrite: true,
	}

	if overwrite, ok := config["overwrite"]; ok {
		plugin.overwrite = overwrite.(bool)
	}
	if data_type, ok := config["type"]; ok {
		plugin.data_type = data_type.(string)
	}
	if language, ok := config["language"]; ok {
		plugin.language = language.(string)
	}
	if database, ok := config["database"]; ok {
		plugin.database = database.(string)
		var (
			c1  *datx.City
			c2  *ipdb.City
			err error
		)
		if plugin.data_type == "datx" {
			c1, err = datx.NewCity(plugin.database)
			plugin.city = unsafe.Pointer(c1)
		} else {
			c2, err = ipdb.NewCity(plugin.database)
			plugin.city = unsafe.Pointer(c2)
		}
		if err != nil {
			klog.Fatalf("could not load %s: %s", plugin.database, err)
		}
	} else {
		klog.Fatal("database must be set in IPIP filter plugin")
	}

	if src, ok := config["src"]; ok {
		plugin.src = src.(string)
		plugin.srcVR = value_render.GetValueRender2(plugin.src)
	} else {
		klog.Fatal("src must be set in IPIP filter plugin")
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
	var a []string
	var err error
	if plugin.data_type == "datx" {
		city := (*datx.City)(plugin.city)
		a, err = city.Find(inputI.(string))
	} else {
		city := (*ipdb.City)(plugin.city)
		a, err = city.Find(inputI.(string), plugin.language)
	}
	if err != nil {
		klog.V(10).Infof("failed to find %s: %s", inputI.(string), err)
		return event, false
	}
	if plugin.target == "" {
		event["country_name"] = a[0]
		event["province_name"] = a[1]
		event["city_name"] = a[2]
		if len(a) >= 5 {
			event["isp"] = a[4]
		}
		if len(a) >= 10 {
			latitude, _ := strconv.ParseFloat(a[5], bitSize)
			longitude, _ := strconv.ParseFloat(a[6], bitSize)
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
		if len(a) >= 5 {
			target["isp"] = a[4]
		}
		if len(a) >= 10 {
			latitude, _ := strconv.ParseFloat(a[5], bitSize)
			longitude, _ := strconv.ParseFloat(a[6], bitSize)
			target["latitude"] = latitude
			target["longitude"] = longitude
			target["location"] = []interface{}{longitude, latitude}
			target["country_code"] = a[11]
		}
		event[plugin.target] = target
	}
	return event, true
}
