Gohangout is an application to do data transport. It consumes data from `input plugin` such as kafka or tcp/udp , and do data transforms using `filter plugin`, and then emit data to `output plugin`, such as Elasticsearch or Clickhouse.

## Install

We could build it from source code , or download binary app.

### build from source code

just clone code and run make

```
make
```

It is recommended to compile it with CGO disabled if you want to run it in docker.

```
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 make
```

### download binary

[https://github.com/childe/gohangout/releases](https://github.com/childe/gohangout/releases) 

### go get

```
go get github.com/childe/gohangout
```

### Third party plugin

- Exmaples for developing 3th party plugin [gohangout-plugin-examples](https://github.com/childe/gohangout-plugin-examples)
- [Kafka Input using Saramp](https://github.com/DukeAnn/gohangout-input-kafka_sarama)
- [Kafka Input using kafka-go](https://github.com/huangjacky/gohangout-input-kafkago)
- [Redis Input](https://github.com/childe/gohangout-input-redis)
- [Split Filter](https://github.com/childe/gohangout-plugin-examples/tree/master/gohangout-filter-split) Split one message to multi
- [File Output](https://github.com/childe/gohangout-plugin-examples/tree/master/gohangout-file-output) file output

## Run

```
gohangout --config config.yml
```

### log

Gohangout use glog.

use `-v n` to set log level. 

I usually set n to 5. You can set it to 10 or 20 to see more detailed log.

### multi threads

--worker 4 (default 1)

above args make gohangout use 4 goroutines to process data.  Notice: one thread to consume from input, and then have 4 goroutines to do filter and output.

### reload

--reload

above args enables reload. Gohangout will relaod config file when it changes.

`kill -USER $pid` also triggers reload.

## Simple Config Example

```
inputs:
    - Kafka:
        topic:
            weblog: 1 # One Kafka consumer thread
        codec: json
        consumer_settings:
            bootstrap.servers: "10.0.0.100:9092"
            group.id: gohangout.weblog
filters:
    - Grok:
        src: message
        match:
            - '^(?P<logtime>\S+) (?P<name>\w+) (?P<status>\d+)$'
            - '^(?P<logtime>\S+) (?P<status>\d+) (?P<loglevel>\w+)$'
        remove_fields: ['message']
    - Date:
        location: 'UTC'
        src: logtime
        formats:
            - 'RFC3339'
        remove_fields: ["logtime"]
outputs:
    - Elasticsearch:
        hosts:
            - 'http://admin:password@127.0.0.1:9200'
        index: 'web-%{appid}-%{+2006-01-02}'
        index_type: "logs"
        bulk_actions: 5000
        bulk_size: 20
        flush_interval: 60
```

## Value Render Protocol

some exmaples and explanation below

```
fields:
    logtime: '%{date} {%time}'
    type: 'weblog'
    hostname: '[host]'
    name: '{{.firstname}}.{{.lastname}}'
    name2: '$.name'
    city: '[geo][cityname]'
    '[a][b]': '[stored][message]'
```

### format 1 JSONPATH

Gohangout use JsonPath to render value if it begins witch `$.`

```
$.store.book[0].title

$['store']['book'][0]['title']

$.store.book[(@.length-1)].title

$.store.book[?(@.price < 10)].title
```

More usage and examples: [https://goessner.net/articles/JsonPath/](https://goessner.net/articles/JsonPath/)

### format2 [X][Y]

**Not recommended, please use format 1**

`city: '[geo][cityname]'` equals to `$.geo.cityname` . It must be strictly `[X][Y]`, in other words, there can not be any other words in front of `[X][Y]` or after it.

### format 3 {{XXX}}

Gohangout will render value using [Golang Template]((https://golang.org/pkg/text/template/)). It could contains other words before or after {{XXX}}, such as `name: 'my name is {{.firstname}}.{{.lastname}}'`

One example you may use: We get a time-type field with `Date` filter, and then render a string with customed format.

```
Add:
  fields:
    ts: '{{ .ts.Format "2006.01.02" }}'
```

### format 4 %{XXX}

it itherits from [logstash](https://www.elastic.co/logstash)

for example, render index name in Elasticsearch output: `web-%{appid}-%{+2006-01-02}` .

## Input

All settings in below plugins could be checked in [Chinese doc](https://github.com/childe/gohangout/blob/master/README.md#input). 

Setting and explanation in English doc will be added later.

- Stdin
- TCP
- Kafka

## Output

- Stdout
- TCP
- Elasticsearch
- Kafka
- ClickHouse

## Filter

### general settings

#### if

syntax example

```
Drop:
    if:
      - 'EQ(name,"childe")'
      - 'Before(-24h) || After(24h)'
```

Relationship between conditions is **AND**, if passes only if all conditions pass.

more complicated example using bool operator: `Exist(a) && (!Exist(b) || !Exist(c))`

All functions supported for now:

**NOtice**: value in EQ/IN functions must be quoted by " , because the value could be a number or a string.  User must tell Gohangout whether it is a string or a number.  
value in other functions could be quoted by " or not , gohangout will treat it as string.

- `Exist(user,name)` if [user][name] exists

- `EQ(user,age,20)` `EQ($.user.age,20)` if [user][age] exists and equal to 20

- `EQ(user,age,"20")` `EQ($.user.age,"20")` if [user][age] exists and equal to "20" (string)

- `IN(tags,"app")` `IN($.tags,"app")` tags is a list or do not pass. if "app" contained in the list

- `HasPrefix(user,name,liu)` `HasPrefix($.user.name,"liu")`

- `HasSuffix(user,name,jia)` `HasSuffix($.user.name,"jia")`

- `Contains(user,name,jia)` `Contains($.user.name,"jia")`

- `Match(user,name,^liu.*a$)` `Match($.user.name,"^liu.*a$")`

- `Random(20)` return true with 5% probability

- `Before(24h)`  *@timestamp* field exists and it must be a Time type
- `After(-24h)`  *@timestamp* field exists and it must be a Time type

#### add_fields

example:

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

Fields could be added if the filter process the event successfully. And it is ignored if filter failed to process the event.

#### remove_fields

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

remove some fields if the filter process the event successfully. And it is ignored if filter failed to process the event. 

### Filter Plugins

- Add
- Convert
- Date # it convert one string-type field to Time-type field
- Drop
- Filters
- Grok
- IPIP
- KV
- Lowercase
- Remove
- Rename
- Split
- Translate
- Uppercase
- Replace
- URLDecode
