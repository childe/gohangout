[ENG](https://github.com/childe/gohangout/blob/master/README-EN.md#input)

之前因为 [logstash](https://www.elastic.co/products/logstash) 处理数据的效率比较低, 用 java 模仿 Logstash 写了一个java版本的 [https://github.com/childe/hangout](https://github.com/childe/hangout).  不知道现在 Logstash 效率怎么样了, 很久不用了.

后来因为Java的太吃内存了, 而且自己对java不熟, 又加上想学习一下golang, 就用golang又写了一次. 内存问题得到了很大的缓解. 目前我们使用golang版本的gohangout每天处理2000亿条以上的数据.

创建一个 QQ 群交流吧

![QQ](https://user-images.githubusercontent.com/2444825/67572448-ca0cb280-f768-11e9-9990-68afd80801ff.png)



## 安装

### 从源码编译

  使用 go module 管理依赖. 直接 make 就可

   > make

### go get

  > go get github.com/childe/gohangout

### docker

镜像地址 https://hub.docker.com/repository/docker/rmself/gohangout/general

### 第三方 Plugin

使用 Plugin 的话，自己编译一下，将 CGO_ENABLED 打开：`CGO_ENABLED=1`

- 开发 Plugin 的例子 [gohangout-plugin-examples](https://github.com/childe/gohangout-plugin-examples)
- [使用sarama 的Kafka Input](https://github.com/DukeAnn/gohangout-input-kafka_sarama)
- [使用 confluent-kafka-go 的Kafka Input](https://github.com/arterhuo/gohangout-input-confluent-kafka-go)
- [使用kafka-go 的Kafka Input](https://github.com/huangjacky/gohangout-input-kafkago)
- [Redis Input](https://github.com/childe/gohangout-input-redis)
- [Split Filter](https://github.com/childe/gohangout-plugin-examples/tree/master/gohangout-filter-split) 一条消息Split 成多条
- [File Output](https://github.com/childe/gohangout-plugin-examples/tree/master/gohangout-file-output) 输出到文件
- [zinc Output](https://github.com/9ji/gohangout-output-zinc) 输出到[zinc](https://github.com/prabhatsharma/zinc)


## 运行

gohangout --config config.yml

一个简单的配置文件如下，从标准输入读取数据，输出到标准输出。具体的配置说明见 [配置一节](#配置)

```yaml
inputs:
    - Stdin: {}

outputs:
    - Stdout: {}
```

### 日志

日志模块使用 github.com/golang/glog , 几个常用参数如下:

- -logtostderr
日志打印出标准错误

-  -v 5
设置日志级别.  我这边一般设置到 5. 如果要看更详细的日志, 可以设置到 10 或者20

### pprof debug

- -pprof=true
(默认是不开启 pprof的)

- -pprof-address 127.0.0.1:8899
pprof 的http地址

### prometheus metrics

运行时加参数 `--prometheus 0.0.0.0:2112`，可以开一个 prometheus 监听服务。

在 Input/Output/Filter 里面配置 `prometheus_counter`

如下例子表示，如果数据通过 if 条件，则此 Add Filter 的计数加 1。

注意，多个 counter 使用同样的配置可能会 Panic

```
Add:
    prometheus_counter: 
        name: gohangout_add_filter
        namespace: rack_a
        help: 'rack_a gohangout add filter counter'
        constLabels:
            env: prod
    if:
    - 'EQ(a,nil)'
    fields:
        op: xyz
```

### 多线程处理

默认是一个线程

--worker 4

使用四个线程(goroutine)处理数据. 每个线程拥有自己的filter, output. 比如说translate filter, 每个线程有自己的字典, 他们占用多份内存.  elasticsearch output也是一样的, 如果每个 elasticsearch 设置了2并发, 那一共就是8个并发.

进一步说明一下为什么添加了这个配置:

最开始是没有这个配置的, 如果需要多线程并发处理数据, 依赖 Input 里面的配置, 比如说 Kafka 配置 `topicname: 2` 就是两个线程去消费(需要 Topic 有至少2个Partition, 保证每个线程可以消费到一个 Partition 里面的数据).

但是后面出现一些矛盾, 比如说, Kafka 的 Consumer 个数多的情况下, 给 Kafka 带来更大压力, 可能导致 Rebalance 更频繁等. 所以如果 Kafka 消费数据没有瓶颈的情况下, 希望控制尽量少的 Consumer, 后面多线程的处理这些数据.

### 自动更新配置

默认不会监听文件系统更新，只在首次初始化时加载配置
--reload

开启这个参数后，当配置文件发生改变会马上触发shutdown，然后重新加载配置文件后运行

除此之外，`kill -USR1 $pid`也会触发重新加载配置文件

## 开发新的插件

- Filter 插件示例参考  [gohangout-filter-title](https://github.com/childe/gohangout-filter-title)
- Input 插件示例参考 [gohangout-input-dot](https://github.com/childe/gohangout-input-dot)
- Output 插件示例参考 [gohangout-output-dash](https://github.com/childe/gohangout-output-dash)
- Decoder 插件示例参考 [gohangout-decode-empty](https://github.com/childe/gohangout-decode-empty)

## 配置

配置文件是 [Yaml](https://yaml.org/) 格式

### 一个简单的配置示例

filters 是一个列表，会**依次**执行里面的每一个 Filter。

如下例，会先执行第一个 Grok Filter，解析 message 字段，按正则表达式提取出一些其他字段。

再执行第二个 Grok Filter，在这个 Grok 中，会首先判断 if 条件是不是符合，如果不符合就跑过不执行这个 Grok 了。

然后执行第三个 Date Filter，将 logtime 字符串转成 Date 类型的字段，存到 timestamp 字段中。

如果有多个 Output，数据会**串行**写到每一个 Output。

如果有多个 Input，每个 Input 进来的数据会**并行**处理后面的 Filter 和 Output。

```
inputs:
    - Kafka:
        topic:
            weblog: 1
        codec: json
        consumer_settings:
            bootstrap.servers: "10.0.0.100:9092"
            group.id: gohangout.weblog
filters:
    - Grok:
        src: message
        match:
            - '^(?P<logtime>\S+) (?P<name>\w+) (?P<cmd>.+)$'
            - '^(?P<logtime>\S+) (?P<name>\w+) (?P<status>\d+)$'
        remove_fields: ['message']
    - Grok:
        if:
          - EQ($.name,"childe")
        src: cmd
        match:
            - '^gohangout .*--config (?P<config_file>\S+)'
    - Date:
        location: 'Asia/Shanghai'
        src: logtime
        target: timestamp
        formats:
            - 'RFC3339'
        remove_fields: ["logtime"]
outputs:
    - Elasticsearch:
        hosts:
            - 'http://admin:password@127.0.0.1:9200'
        index: 'web-%{appid}-%{+2006-01-02}' #golang里面的渲染方式就是用数字, 而不是用YYMM.
        index_type: "logs"
        bulk_actions: 5000
        bulk_size: 20
        flush_interval: 60
```

## 字段格式约定

以 Add Filter 举例

```
fields:
    logtime: '%{date} %{time}'
    type: 'weblog'
    hostname: '[host]'
    name: '{{.firstname}}.{{.lastname}}'
    name2: '$.name'
    city: '[geo][cityname]'
    '[a][b]': '[stored][message]'
    indename: 'app-%{@metadata}{kafka}{topic}-%{+2006-01-02}-log'
```

### 格式1 JSONPATH 格式

**相比格式2, 更推荐使用这种格式. 更标准, 也灵活, 性能也足够**

如果以 `$.` 开头, 认为是这种格式

给几个下面文中的例子

```
$.store.book[0].title

$['store']['book'][0]['title']

$.store.book[(@.length-1)].title

$.store.book[?(@.price < 10)].title
```

具体的格式和例子参见 [https://goessner.net/articles/JsonPath/](https://goessner.net/articles/JsonPath/)

### 格式2 [XX][YY]

**不再推荐使用, 请使用格式1**

`city: '[geo][cityname]'` 是把 geo.cityname 的值赋值给 city 字段. 必须严格 [XX][YY] 格式, 前后不能有别的内容


### 格式3 {{XXX}}

如果含有 `{{XXX}}` 的内容, 就认为是 golang template 格式, 具体语法可以参考 [https://golang.org/pkg/text/template/](https://golang.org/pkg/text/template/). 前后及中间可以含有别的内容, 像 `name: 'my name is {{.firstname}}.{{.lastname}}'`

Gohangout 使用了 [https://github.com/Masterminds/sprig/](https://github.com/Masterminds/sprig/) 的函数库

来举个例子吧, Date Filter 得到一个 Time 类型的字段, 然后按自己的格式格式化一个字符串出来

```
Add:
  fields:
    ts: '{{ .ts.Format "2006.01.02" }}'  ## 这里是使用了 Time 类型的自己的函数, 相当于 ts = ts.Format("2006.01.02")
    c: '{{ add .a .b }}' ## add 是 sprig 库里面的函数，相当于 c = a + b
```

### 格式4 %{XXX}{YYY}

含有 `%{XXX}{YYY}` 的内容, 使用自己定义的格式处理, 像上面的 `%{date} %{time}` 是把 date 字段和 time 字段组合成一个 logtime 字段. 前后以及中间可以有任何内容. 像 Elasticsearch 中的 index: `web-%{appid}-%{+2006-01-02}` 也是这种格式, %{+XXX} , 前面一个加号, 代表时间字段, 会按时间格式做格式化处理.

2006 01 02 15 04 05 这几个数字是 golang 里面特定的数字, 代表年月日时分秒. 1月2号3点4分5秒06年. 其实就像hangout里面的YYYY MM dd HH mm SS.
如果日期月份包含英文，也可把01换成Jan，比如：02-Jan-2006.

### 格式5 除了1,2,3,4 之外的其它

在不同Filter中, 可能意义不同. 像 Date 中的 src: logtime, 是说取 logtime 字段的值.
Elasticsearch 中的 index_type: logs , 这里的 logs 不是指字段名, 就是字面值.


## INPUT

### Stdin

```
Stdin:
    codec: json
```

从标准输入读取数据.

#### codec
目前有json/plain/json:not_usenumber三种.

- json 对数据做 json 解析, 如果解析失败, 则将整条数据写到 message 字段, 并添加当前时间到 `@timestamp` 字段. 如果解析成功而且数据中没有 `@timestamp` 字段, 则添加当前时间到 `@timestamp` 字段.
- plain 将整条数据写到 message 字段, 并添加当前时间到 `@timestamp` 字段.
- json:not_usenumber 因为数字类型的位数有限, 会有一个最高精度, 为了不损失精度, 默认的 json 配置情况下, 数字类型的值默认转成字符串保存. 如果需要存成数字, 比如后续是要写 clickhouse, 可以使用 json:not_usenumber.  如果使用 json codec, 也可以配置 Convert Filter 转换成数字.

### TCP

```
TCP:
    network: tcp4
    address: 0.0.0.0:10000
    codec: plain
```

#### network

默认为 tcp , 可以明确指定使用 tcp4 或者 tcp6

#### address

监听端口, 无默认值, 必须设置

#### codec

默认 plain


### Kafka

```
Kafka:
    decorate_events: false
    topic:
        weblog: 1
    #assign:
    #   weblog: [0,9]
    codec: json
    consumer_settings:
        bootstrap.servers: "10.0.0.100:9092,10.0.0.101:9092"
        group.id: gohangout.weblog
        max.partition.fetch.bytes: '10485760'
        auto.commit.interval.ms: '5000'
        from.beginning: 'true'
        messages_queue_length: 10

        # sasl.mechanism: PLAIN
        # sasl.user: admin
        # sasl.password: admin-secret

        # tls.enabled: true
        # tls:
        #     cert: 'path/to/cert'
        #     key: 'path/to/key'
        #     ca: 'path/to/ca'
        #     insecure.skip.verify: false
        #     servername: xx

```

**特别注意** 参数需要是字符串, 像 `auto.commit.interval.ms: '5000'` , 以及 `from.beginning: 'true'` , 等等

#### decorate_events

默认为 false
配置为 true 的话, 可以把 topic/partition/offset 信息添加到 ["@metadata"]["kafka"] 字段中


#### topic

`weblog: 1` 是指开一个goroutine去消费 weblog 这个topic. 可以配置多个topic, 多个goroutine, 但我这边在实践中都是使用多进程(docker), 而不是多goroutine.

#### assign

assign 配置用来只消费特定的partition, 和 `topic` 配置是冲突的, 只能选择一个.

#### consumer_settings

bootstrap.servers group.id 必须配置

auto.commit.interval.ms 是指多久commit一次offset, 太长的话有可能造成数据重复消费,太短的话可能会对kafka造成太大压力.

max.partition.fetch.bytes 是指kafka client一次从kafka server读取多少数据,默认是10MB

from.beginning 如果第一次消费此topic, 或者是offset已经失效, 是从头消费还是从最新消费. 默认是 false. 但是如果已经有过commit offset, 会接着之前的消费.

messages_queue_length: 内部使用的消息 channel 的长度，默认为10.

sasl.mechanism 认证方式, 目前还只支持 PLAIN 一种

sasl.user sasl认证的用户名

sasl.password sasl认证的密码

servername 如果 servername 不为空的话，证书中的 IP 或者 DNS 名字，需要包含servername

更多配置参见 [https://github.com/childe/healer/blob/dev/config.go#L40](https://github.com/childe/healer/blob/dev/config.go#L40)

## OUTPUT

### Stdout

```
Stdout:
    if:
        - '{{if .error}}y{{end}}'
```

输出到标准输出

if的语法参考下面 [IF语法](#if)

### TCP

```
TCP:
    network: tcp4
    address: 127.0.0.1:10000
    concurrent: 2
```

#### network

默认为 tcp , 可以明确指定使用 tcp4 或者 tcp6

#### address

TCP 远端地址, 无默认值, 必须设置

#### concurrent

开几个 tcp 连接一起写, 默认1


### Elasticsearch

```
Elasticsearch:
    hosts:
        - 'http://10.0.0.100:9200'
        - 'http://admin:password@10.0.0.101:9200'
    # sniff:
        # refresh_interval: 3600
        # match: 'EQ($.attributes.type,"hot")'
    index: 'web-%{appid}-%{+2006-01-02}' #golang里面的渲染方式就是用数字, 而不是用YYMM.
    index_time_location: 'Local'
    index_type: "logs"
    bulk_actions: 5000
    routing: '[domain]'
    id: '[orderid]'
    bulk_size: 20
    flush_interval: 60
    concurrent: 3
    compress: false
    es_version: 7
    retry_response_code: [401, 502]
```

#### sniff

[功能需求 es output 支持特定节点名的 sniffer](#117) 

- refresh_interval 是指多后台长时间去 Sniff 一次, 设置为 0 的话不会在后台刷新
- match 是过滤条件, 符合条件的节点才会加到 Bulk 使用的列表中

Sniff 会调用 `_nodes/_all/http` 获取节点信息, 返回 `publish_address` 信息

#### index_time_location

渲染索引名字时, 使用什么时区. 默认是 UTC. 北京时间 2019-10-25 07:00:00 的日志, 会写到 2019.10.24 这个索引中.

内容如 `Asia/Shanghai` 等, 参考 [https://timezonedb.com/time-zones](https://timezonedb.com/time-zones)

两个特殊值: `UTC` `Local`

#### bulk_actions

多少次提交一次Bulk请求到ES集群. 默认 5000

#### bulk_size

单位是MB, 多少大写提交一次到ES. 默认 15MB

#### flush_interval

单位秒, 间隔多少时间提交一次到ES. 默认 30

#### concurrent

bulk 的goroutine 最大值, 默认1

举例来说, 如果Bulk 1W条数据到ES需要5秒, 1W条数据从Input处理完所有Filters然后到Output也需要5秒. 那么把concurrent设置为1就合适, Bulk是异步的, 这5秒钟gohangout会去Filter接下来的数据.

如果Bulk 1W条数据需要10秒, Filter只要5秒, 那么concurrent设置为2可以达到更大的吞吐量.

#### routing

默认为空, 不做routing

#### id
默认为空, 不设置id (文档id由ES生成)

#### compress

默认 true, http请求时做zip压缩

#### es_version

默认为6，可以适配es6的版本，如果设置为7，则可以适配Elasticsearch7以上版本

#### retry_response_code

默认 [401, 502] , 当Bulk请求的返回码是401或者502时, 会重试.

#### 两个额外的配置

```
source_field: _source
bytes_source_field: _source
```

没有这个配置的时候, 会把日志做 json.dump, 拿到dump后的[]byte写ES. 如果source_field或者bytes_source_field配置了, 则直接把配置的字段(上面的例子是 `_source` 字段)做为[]byte写到ES.

bytes_source_field优先级高于source_field.  bytes_source_field是指字段是[]byte类型, source_field是指字段是string类型

增加这个配置的来由是这样的. 上游数据源已经是 json.dump之后的[]byte数据, 做一次json.parse, 然后再json.dump, 耗费了大量CPU做无用功.

### Kafka

**特别注意** 参数需要是字符串, 像 `flush.interval.ms: "3000"` , 等等

```
Kafka:
    topic: applog
    producer_settings:
        bootstrap.servers: node1.kafka.corp.com:9092,node2.kafka.corp.com:9092,node3.kafka.corp.com:9092
        flush.interval.ms: "3000"
        metadata.max.age.ms: "10000"
        # healer.magicbyte: "1"
        # sasl.mechanism: PLAIN
        # sasl.user: admin
        # sasl.password: admin-secret
```

healer.magicbyte: "1" 是说写入的Kafka消息以版本1编码，默认是版本0. 版本1会在消息中加入时间戳信息。

### clickhouse

```
Clickhouse:
    table: 'hotel.weblog'
    conn_max_life_time: 1800
    username: admin
    password: XXX
    hosts:
    - 'tcp://10.100.0.101:9000'
    - 'tcp://10.100.0.102:9000'
    # fields: ['datetime', 'appid', 'c_ip', 'domain', 'cs_method', 'cs_uri', 's_ip', 'sc_status', 'time_taken']
    bulk_actions: 1000
    flush_interval: 30
    concurrent: 1
```

*Notice:* 如果表中字段有 default 值, 目前只支持字符串和数字 的 DEFAULT 表达式解析和处理, 如果像 IPv4设置了default 值, 是处理不了的. 代码中写死了 IPv4 和 IPv6 的默认值都是0

#### table

表名. 必须配置

#### hosts

clickhouse 节点列表. 必须配置

#### fields

初始化的时候会从 ClickHouse 里面读取所有字段。

也可以手工配置，会优先使用手工配置。 为了暂时缓解 #159

#### bulk_actions

多少次提交一次Bulk请求到ES集群. 默认 1000

#### flush_interval

单位秒, 间隔多少时间提交一次到ES. 默认 30

#### concurrent

bulk 的goroutine 最大值, 默认1

#### conn_max_life_time

到 ClickHouse 的连接的生存时间, 单位为秒. 默认不设置, 也就是生存时间无限长.

## FILTER

### 通用字段

#### if

if 的语法如下

```
Drop:
    if:
      - '{{if .name}}y{{end}}'
      - '{{if eq .name "childe"}}y{{end}}'
      - '{{if or (before . "-24h") (after . "24h")}}y{{end}}'
```

if 数组中的条件是 AND 关系, 需要全部满足.

目前 if 支持两种语法, 一种是 golang 自带的 template 语法, 一种是我自己实现的一套简单的DSL, 实现的常用的一些功能, 性能远超 template , 我把上面的语法按自己的DSL翻译一下.

```
Drop:
    if:
      - 'EQ(name,"childe")'
      - 'Before(-24h) || After(24h)'
```

也支持括号和逻辑运算符, 像 `Exist(a) && (!Exist(b) || !Exist(c))`

目前支持的函数如下:

注意:

**EQ/IN 函数需要使用双引号代表字符串, 因为他们也可能做数字的比较, 其他所有函数都不需要双引号, 因为他们肯定是字符串函数**

**EQ IN HasPrefix HasSuffix Contains Match , 这几个函数可以使用 [jsonpath](https://github.com/oliveagle/jsonpath) 表示, 除 EQ/IN 外需要使用双引号**

- `Exist(user,name)` [user][name]存在

- `EQ(user,age,20)` `EQ($.user.age,20)` [user][age]存在并等于20

- `EQ(user,age,"20")` `EQ($.user.age,"20")` [user][age]存在并等于"20" (字符串)

- `IN(tags,"app")` `IN($.tags,"app")` "app"存在于 tags 数组中, tags 一定要是数组,否则认为条件不成立

- `HasPrefix(user,name,liu)` `HasPrefix($.user.name,"liu")` [user][name]存在并以 liu 开头

- `HasSuffix(user,name,jia)` `HasSuffix($.user.name,"jia")` [user][name]存在并以 jia 结尾

- `Contains(user,name,jia)` `Contains($.user.name,"jia")` [user][name]存在并包含 jia

- `Match(user,name,^liu.*a$)` `Match($.user.name,"^liu.*a$")` [user][name]存在并能匹配正则 `^liu.*a$`

- `Random(20)` 1/20 的概率返回 true

- `Before(24h)`  *@timestamp* 字段存在, 并且是 time.Time 类型, 并且在`当前时间+24小时`之前
- `After(-24h)`  *@timestamp* 字段存在, 并且是 time.Time 类型, 并且在`当前时间-24小时`之后

#### add_fields

例:

```
Grok:
    src: message
    match:
        - '^(?P<logtime>\S+) (?P<name>\w+) (?P<status>\d+)$'
        - '^(?P<logtime>\S+) (?P<status>\d+) (?P<loglevel>\w+)$'
    remove_fields: ['message']
    add_fields:
      grok_result: 'ok'
```

当Filter执行成功时, 可以添加一些字段. 如果Filter失败, 则忽略. 下面具体的Filter说明中, 提到的"返回false", 就是指Filter失败

#### remove_fields

例子如上. 当Filter执行成功时, 可以删除一些字段. 如果Filter失败, 则忽略.

#### failTag

当Filter执行失败时, 可以添加内容到 `tags` 字段. 如果Filter成功, 则忽略. 如果 tags 字段已经存在, 则将 tags 设置为数组并添加新的数据.

```
Grok:
    src: message
    match:
        - '^(?P<logtime>\S+) (?P<name>\w+) (?P<status>\d+)$'
        - '^(?P<logtime>\S+) (?P<status>\d+) (?P<loglevel>\w+)$'
    remove_fields: ['message']
    add_fields:
      grok_result: 'ok'
    failTag: grokfail
```

#### overwrite

配置的新字段要不要覆盖之前已有的字段, 默认 true


### Add

```
Add:
  overwrite: true
  fields:
      name: childe
      hostname: '[host]'
      logtime: '{{.date}} {{.time}}'
      message: '[stored][message]'
      '[a][b]': '[stored][message]'
```

**更多写法参见 [字段格式约定](https://github.com/childe/gohangout/#%E5%AD%97%E6%AE%B5%E6%A0%BC%E5%BC%8F%E7%BA%A6%E5%AE%9A)**

1. 增加 name 字段, 内容是 childe
2. 增加 hostname 字段, 内容是原 host 字段中的内容. (相当于改名)
3. 增加 logtime 字段, 内容是 date 和 time 两个字段的拼接
4. 增加 message 字段, 是 event.stored.message 中的内容
4. 将 event.stored.message 中的内容写入 event.a.b 字段中(如果没有则创建)

overwrite: true 的情况下, 这些新字段会覆盖老字段(如果有的话).

### Convert

现在只支持转成 float/int/string/bool 这四种类型

```
Convert:
    fields:
        time_taken:
            remove_if_fail: false
            setto_if_nil: 0.0
            setto_if_fail: 0.0
            to: float
        sc_bytes:
            to: int
            remove_if_fail: true
        status:
            to: bool
            remove_if_fail: false
            setto_if_fail: true
        map_struct:
            to: string
            setto_if_fail: ""
```

#### remove_if_fail

如果转换失败刚删除这个字段, 默认 false

#### setto_if_fail: XX

如果转换失败, 刚将此字段的值设置为 XX . 优先级比 remove_if_fail 低.  如果 remove_if_fail 设置为 true, 则setto_if_fail 无效.

#### setto_if_nil: XX

如果没有这个字段, 刚将此字段的值设置为 XX . 优先级最高


### Date

```
Date:
    src: 'logtime'
    target: '@timestamp'
    location: Asia/Shanghai
    add_year: false
    overwrite: true
    formats:
        - 'RFC3339'
        - '2006-01-02T15:04:05'
        - '2006-01-02T15:04:05Z07:00'
        - '2006-01-02T15:04:05Z0700'
        - '2006-01-02'
        - 'UNIX'
        - 'UNIX_MS'
    remove_fields: ["logtime"]
```

Date Filter 的作用是把一个字符串类型的字段, 转成一个 Time 类型的字段, 存到 target 里面去.

一个比较常见的问题是, 如果写数据到 Clickhouse, 其中有 Datetime 类型的字段, 比如叫 createTime, 建议先用 Date Filter 转成(生成)一个 Time 类型的字段,  存到 createTime 里面.

如果源字段不存在, 返回 false. 如果所有 formats 都匹配失败, 返回 false

#### src

源字段, 必须配置.

#### target

目标字段, 默认 `@timestamp`

#### overwrite

默认 true, 如果目标字段已经存在, 会覆盖

#### add_year

有些日志中的时间戳不带年份信息, 默认 false . add_year: true 可以先在源字段最前面加四位数的年份信息然后再解析.

#### formats

必须设置. 格式参考 [https://golang.org/pkg/time/](https://golang.org/pkg/time/)

除此外, 还有 UNIX UNIX_MS 两个可以设置

### Drop

丢弃此条消息, 配置 if 条件使用

```
Drop:
    if:
        - '{{if .name}}y{{end}}'
        - '{{if eq .name "childe"}}y{{end}}'
        - '{{if or (before . "-24h") (after . "24h")}}y{{end}}'
```

### Filters

目的是为了一个 if 条件后跟多个Filter

```
Filters:
    if:
        - '{{if eq .name "childe"}}y{{end}}'
    filters:
        - Add:
            fields:
                a: 'xyZ'
        - Lowercase:
            fields: ['url', 'domain']
```

### Grok

Grok 是使用正则表达式来提取内容的 Filter。

```
Grok:
    src: message
    match:
        - '^(?P<logtime>\S+) (?P<name>\w+) (?P<status>\d+)$'
        - '^(?P<logtime>\S+) (?P<status>\d+) (?P<loglevel>\w+)$'
    ignore_blank: true
    remove_fields: ['message']
    pattern_paths:
    - 'https://raw.githubusercontent.com/vjeantet/grok/master/patterns/grok-patterns'
    - '/opt/gohangout/patterns/'
```

源字段不存在, 返回 false. 所有格式不匹配, 返回 false

#### src

源字段, 默认 message

#### target

目标字段, 默认为空, 直接写入根下. 如果不为空, 则创建target字段, 并把解析后的字段写到target下.

#### match

依次匹配, 直到有一个成功.

#### pattern_paths

会加载定义的 patterns 文件. 如果是目录会加载目录下的所有文件.

这里推荐 [https://github.com/vjeantet/grok](https://github.com/vjeantet/grok) 项目, 里面把 logstash 中使用的 pattern 都翻译成了 golang 的正则库可以使用的.

#### ignore_blank

默认 true. 如果匹配到的字段为空字符串, 则忽略这个字段. 如果 ignore_blank: false , 则添加此字段, 其值为空字符串.

### IPIP

根据 IP 信息补充地址信息, 会生成如下字段.

country_name province_name city_name

下面四个字段视情况生成, 可能会缺失. latitude longitude location country_code

如果没有源字段, 或者寻找失败, 返回 false

```
IPIP:
    src: clientip
    target: geoip
    database: /opt/gohangout/mydata4vipday2.datx
    type: datx
```

#### database

数据库地址. 数据可以在 [https://www.ipip.net/](https://www.ipip.net/) 下载

#### type
数据文件的类型，可选值ipdb和datx，默认是datx

#### language
ipdb查找城市时候需要传入语言，默认是CN

#### src

源字段, 必须设置

#### target

目标字段, 如果不设置, 则将IPIP Filter生成的所有字段写入到根一层.

### KV
将 a=1&b=2, 或者name=app id=123 type=nginx 这样的字符串拆分成{a:1,b:2}  {name:app, id:123, type:nginx} 等多个字段, 放到日志里面去.

配置如下

如果targete有定义, 会把拆分出来字段放在这个字段中, 如果没有定义,放到在顶层.
trim 是把拆分出来的字段内容做前后修整. 将不需要的字符去掉. 下面的示例就是说把双引号和tag都去掉.
trim_key和trim类似, 处理的是字段名称.

```
KV:
  src: msg
  target: kv
  field_split: ','
  value_split: '='
  trim: '\t "'
  trim_key: '"'
  tag_on_failure: "KVfail"
  remove_fields: ['msg']
```

#### src

源字段, 必须设置

#### target

目标字段, 如果不设置, 则将IPIP Filter生成的所有字段写入到根一层.

#### field_split

各字段&值之间以什么分割, 一般都是逗号或者空格之类. 必须设置

#### value_split

字段名和值之间以什么连接, 一般是等号. 必须设置

### Json

如果源字段不存在, 或者Json.parse 失败, 返回 false

```
Json:
    field: request
    target: request_fields
```

#### field

源字段

#### target

目标字段, 如果不设置, 则将Json Filter生成的所有字段写入到根一层.

### LinkMetric

做简单的流式统计, 统计多个字段之间的聚合数据.

```
LinkMetric:
    fieldsLink: 'domain->serverip->status_code'
    timestamp: '@timestamp'
    reserveWindow: 1800
    batchWindow: 600
    windowOffset: 0
    accumulateMode: cumulative
    drop_original_event: false
    reduce: false
```

每600s输出一次, 输出结果形式如下:

```
{"@timestamp":1540794825600,"domain":"www.ctrip.com","serverip":"10.0.0.100","status_code":"200",count:10}
{"@timestamp":1540794825600,"domain":"www.ctrip.com","serverip":"10.0.0.200","status_code":"404",count:1}
...
```

#### fieldsLink

字段以 `->` 间隔, 统计一定时间内的聚合信息

#### timestamp

使用哪个字段做时间戳. 这个字段必须是通过 Date Filter 生成的(保证是 time.Time 类型)

#### batchWindow

多长时间内的数据聚合在一起, 单独是秒. 每隔X秒输出一次. 如果设置为1800 (半小时), 那么延时半小时以上的数据会被丢弃.

#### reserveWindow

保留多久的数据, 单独是秒. 因为数据可能会有延时, 所以需要额外保存一定时间的数据在内存中.

#### accumulateMode

两种聚合模式.

1. cumulative 累加模式. 假设batchWindow 是300, reserveWindow 是 1800. 在每5分钟时, 会输出过去5分钟的一批聚合数据, 同时因为延时的存在, 可能还会有(过去10分钟-过去5分钟)之间的一批数据. cumulative 配置下, 会保留(过去10分钟-过去5分钟)之前count值的内存中, 新的数据进来时, 累加到一起, 下个5分钟时, 输出一个累加值.

2. separate 独立模式. 每个5分钟输出之后, 把各时间段的值清为0, 从头计数.

#### windowOffset

延时输出, 默认为0. 如果设置 windowOffset 为1 , 那么每个5分钟输出时, 最近的一个window数据保留不会输出.

#### drop_original_event

是否丢弃原始数据, 默认为 false. 如果设置为true, 则丢弃原始数据, 只输出聚合统计数据.

#### reduce

是否 reduce , 默认 false.  如果为true, 则会解析数据中的 `count`, `sum` 字段. 举例来说, 一个 topic 有10个 partitions, 有10个消费者做聚合, 聚合后的数据通过 tcp output 吐给另外一个进程R, 进程R对聚合数据进行 reduce 操作, 然后再把二次聚合后的数据吐到ES或者Influxdb, 这样最终写到ES或Influxdb的数据就是10个partitions的总的数据.

### LinkStatsMetric

和 LinkMetric 类似, 但最后一个字段需要是数字类型, 对它进行统计.

举例:

```
- LinkStatsMetric:
    fieldsLink: 'domain->serverip->status_code->response_time'
    timestamp: '@timestamp'
    reserveWindow: 1800
    batchWindow: 60
    accumulateMode: separate
```

### Lowercase

```
Lowercase:
  fields: ['domain', 'url']
```

### Remove

```
Remove:
  fields: ['domain', 'url']
```

### Rename

```
Rename:
  fields:
    host: hostname
    serverIP: server_ip
```


### Split

```
Split:
  src: message
  sep: "\t"
  maxSplit: -1
  fields: ['logtime', 'hostname', 'uri', 'return_code']
  ignore_blank: true
  overwrite: true
  trim: '"]['
```

#### src

数据来源字段, 默认 message , 如果字段不存在, 返回false

#### sep

分隔符, 在 strings.SplitN(src, sep, maxSplit) 中用被调用. 必须配置.

如果分隔符包含不可见字符, yaml配置以及gohangout也是支持的, 像下面这样

```
sep: "\x01"
```

#### dynamicSep

默认 false。如果设置为 true，则认为 sep 是一个模板，而不是写死的字符串。
比如 `sep: '[tt]'`，会使用event 里面的 tt 字段的值做为真正的 sep 。

#### maxSplit

在 strings.SplitN(src, sep, maxSplit) 中用被调用, 默认 -1, 代表无限制

#### fields

分割后的字符串数组长度不能比配置的 fields 长度小。

如果分割后的字符串数组长度比配置的 fields 多，则多余的会被忽略掉。可以配合 maxSplit 使用。

#### ignore_blank

如果分割后的某字段为空, 刚不放后 event 中, 默认 true

#### trim

用来把分割后的字段, 去除两边的一些空格或者是标点等.

### Translate

字段翻译. 字典使用 yaml 格式. 配置例子如下:

```
Translate:
    dictionary_path: http://git.corp.com/childe/dict/raw/master/ip2appid.yml
    refresh_interval: 3600
    source: server_ip
    target: app_id
```

### Uppercase

```
Uppercase:
  fields: ['loglevel']
```

### Replace

```
- Replace:
  fields:
    name: ['wang', 'Wang', 1]
- Replace:
  fields:
    name: ['en', 'eng']
```

最后面的 1 代表, 只替换一次. 如果不给这个值, 代表替换所有的.

比如上面, 就是把 name 字段中的第一个 wang 换成 Wang, 把所有 en 换成 eng

### Gsub

```
- Gsub:
  fields:
    - field: msg1
      src: "/"
      repl: "_"
    - field: msg2
      src: '[\\?#-]'
      repl: "."
- Gsub:
  fields:
    - field: msg
      src: "(^\\w+)"
      repl: "xxx-$1-yyy"
```

使用 [https://pkg.go.dev/regexp#Regexp.ReplaceAll](https://pkg.go.dev/regexp#Regexp.ReplaceAll) 做替换。
