module github.com/childe/gohangout

go 1.13

require (
	github.com/ClickHouse/clickhouse-go v1.5.4
	github.com/IBM/sarama v1.41.2
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/bkaradzic/go-lz4 v1.0.1-0.20160924222819-7224d8d8f27e // indirect
	github.com/childe/healer v0.5.5
	github.com/fsnotify/fsnotify v1.5.1
<<<<<<< HEAD
	github.com/golang/snappy v0.0.4 // indirect
=======
	github.com/golang/glog v1.0.0
>>>>>>> e465238 (1. 增加sarama版本的kafka input)
	github.com/google/uuid v1.3.0 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/ipipdotnet/datx-go v0.0.0-20181123035258-af996d4701a0
	github.com/ipipdotnet/ipdb-go v1.3.1
	github.com/json-iterator/go v1.1.12
	github.com/magiconair/properties v1.8.6
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0
	github.com/oliveagle/jsonpath v0.0.0-20180606110733-2e52cf6e6852
	github.com/prometheus/client_golang v1.12.1
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/spf13/cast v1.4.1
	github.com/yalp/jsonpath v0.0.0-20180802001716-5cc68e5049a0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/klog/v2 v2.100.1
)

replace github.com/spf13/cast v1.4.1 => github.com/oasisprotocol/cast v0.0.0-20220606122631-eba453e69641
