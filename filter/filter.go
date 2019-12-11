package filter

import (
	"plugin"

	"github.com/childe/gohangout/topology"
	"github.com/golang/glog"
)

func BuildFilter(filterType string, config map[interface{}]interface{}) topology.Filter {
	switch filterType {
	case "Add":
		f := NewAddFilter(config)
		return f
	case "Remove":
		f := NewRemoveFilter(config)
		return f
	case "Rename":
		f := NewRenameFilter(config)
		return f
	case "Lowercase":
		f := NewLowercaseFilter(config)
		return f
	case "Uppercase":
		f := NewUppercaseFilter(config)
		return f
	case "Split":
		f := NewSplitFilter(config)
		return f
	case "Grok":
		f := NewGrokFilter(config)
		return f
	case "Date":
		f := NewDateFilter(config)
		return f
	case "Drop":
		f := NewDropFilter(config)
		return f
	case "Json":
		f := NewJsonFilter(config)
		return f
	case "Translate":
		f := NewTranslateFilter(config)
		return f
	case "Convert":
		f := NewConvertFilter(config)
		return f
	case "URLDecode":
		f := NewURLDecodeFilter(config)
		return f
	case "Replace":
		f := NewReplaceFilter(config)
		return f
	case "KV":
		f := NewKVFilter(config)
		return f
	case "IPIP":
		f := NewIPIPFilter(config)
		return f
	case "Filters":
		f := NewFiltersFilter(config)
		return f
	case "LinkMetric":
		f := NewLinkMetricFilter(config)
		return f
	case "LinkStatsMetric":
		f := NewLinkStatsMetricFilter(config)
		return f
	default:
		p, err := plugin.Open(filterType)
		if err != nil {
			glog.Fatalf("could not open %s: %s", filterType, err)
		}
		newFunc, err := p.Lookup("New")
		if err != nil {
			glog.Fatalf("could not find New function in %s: %s", filterType, err)
		}
		return newFunc.(func(map[interface{}]interface{}) interface{})(config).(topology.Filter)
	}
}
