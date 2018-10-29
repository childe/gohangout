## RUN

<<<<<<< HEAD
gohangout --config config.yml

## 一个简单的配置

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
            - '^(?P<logtime>\S+) (?P<name>\w+) (?P<status>\d+)$'
            - '^(?P<logtime>\S+) (?P<status>\d+) (?P<loglevel>\w+)$'
        remove_fields: ['message']
    - Date:
        location: 'Asia/Shanghai'
        src: logtime
        formats:
            - 'RFC3339'
        remove_fields: ["logtime"]
outputs:
    - Elasticsearch:
        hosts:
            - http://127.0.0.1:9200
        index: 'web-%{+2006-01-02}' #golang里面的渲染方式就是用数字, 而不是用YYMM.
        index_type: "logs"
        bulk_actions: 5000
        bulk_size: 20
        flush_interval: 60
```

=======
## 一个简单的配置

>>>>>>> 042582bffa7582ac24b223f13d0690699821ec5b
## 字段格式约定

以 Add Filter 举例

```
fields:
<<<<<<< HEAD
    type: 'weblog'
    hostname: '[host]'
    name: '{{.firstname}}.{{.lastname}}'
    city: '[geo][cityname]'
    '[a][b]': '[stored][message]'
```

### 格式1 [XX][YY]

`city: '[geo][cityname]'` 是把 geo.cityname 的值赋值给 city 字段. 必须严格 [XX][YY] 格式, 前后不能有别的内容


### 格式2 {{XXX}}

如果含有 `{{XXX}}` 的内容, 就认为是 golang template 格式, 具体语法可以参考 [https://golang.org/pkg/text/template/](https://golang.org/pkg/text/template/). 前后及中间可以含有别的内容, 像 `name: 'my name is {{.firstname}}.{{.lastname}}'`

### 格式3 除了1,2 之外的其它

在不同Filter中, 可能意义不同. 像 Date 中的 src: logtime, 是说取 logtime 字段的值.  
Elasticsearch 中的 index_type: logs , 这里的 logs 不是指字段名, 就是字面值.


=======
    xxx: xxx
    yyy: '[client]'
    zzz: '[stored][message]'
    '[a][b]': '[stored][message]'
```

>>>>>>> 042582bffa7582ac24b223f13d0690699821ec5b
## INPUT

### 

## OUTPUT

## FILTER

### 通用字段

#### if

#### add_fields

<<<<<<< HEAD
例:

```
- Grok:
    src: message
    match:
        - '^(?P<logtime>\S+) (?P<name>\w+) (?P<status>\d+)$'
        - '^(?P<logtime>\S+) (?P<status>\d+) (?P<loglevel>\w+)$'
    remove_fields: ['message']
    add_fields:
      grok_result: 'ok'
```

当Filter执行成功时, 可以添加一些字段. 如果Filter失败, 则忽略.

#### remove_fields

例子如上. 当Filter执行成功时, 可以删除一些字段. 如果Filter失败, 则忽略.

#### failTag

当Filter执行失败时, 可以添加内容到 `tags` 字段. 如果Filter成功, 则忽略. 如果 tags 字段已经存在, 则将 tags 设置为数组并添加新的数据.

```
- Grok:
    src: message
    match:
        - '^(?P<logtime>\S+) (?P<name>\w+) (?P<status>\d+)$'
        - '^(?P<logtime>\S+) (?P<status>\d+) (?P<loglevel>\w+)$'
    remove_fields: ['message']
    add_fields:
      grok_result: 'ok'
    failTag: grokfail

=======
#### remove_fields

#### failTag

>>>>>>> 042582bffa7582ac24b223f13d0690699821ec5b
#### overwrite

配置的新字段要不要覆盖之前已有的字段, 默认 true


### Add

```
Add:
  overwrite: true
  fields:
<<<<<<< HEAD
      name: childe
=======
      name: liujia
>>>>>>> 042582bffa7582ac24b223f13d0690699821ec5b
      hostname: '[host]'
      logtime: '{{.date}} {{.{time}}
      message: '[stored][message]'
      '[a][b]': '[stored][message]'
```

<<<<<<< HEAD
1. 增加 name 字段, 内容是 childe
=======
1. 增加 name 字段, 内容是 liujia
>>>>>>> 042582bffa7582ac24b223f13d0690699821ec5b
2. 增加 hostname 字段, 内容是原 host 字段中的内容. (相当于改名)
3. 增加 logtime 字段, 内容是 date 和 time 两个字段的拼接
4. 增加 message 字段, 是 event.stored.message 中的内容
4. 将 event.stored.message 中的内容写入 event.a.b 字段中(如果没有则创建)

overwrite: true 的情况下, 这些新字段会覆盖老字段(如果有的话).

### Convert

```
- Convert:
    remove_if_fail: false
    setto_if_fail: 0
    fields:
        time_taken:
            to: float
        sc_bytes:
            to: int

- Convert:
    remove_if_fail: false
    setto_if_fail: true
    fields:
        status:
            to: bool
```

#### remove_if_fail

如果转换失败刚删除这个字段, 默认 false

#### setto_if_fail: XX

如果转换失败, 刚将此字段的值设置为 XX . 优先级比 remove_if_fail 低.  如果 remove_if_fail 设置为 true, 则setto_if_fail 无效.

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
<<<<<<< HEAD
        - '{{if eq .name "childe"}}y{{end}}'
=======
        - '{{if eq .name "liujia"}}y{{end}}'
>>>>>>> 042582bffa7582ac24b223f13d0690699821ec5b
        - '{{if or (before . "-24h") (after . "24h")}}y{{end}}'
```

### Filters

<<<<<<< HEAD
目的是为了一个 if 条件后跟多个Filter

```
  - Filters:
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

```
- Grok:
    src: message
    match:
        - '^(?P<logtime>\S+) (?P<name>\w+) (?P<status>\d+)$'
        - '^(?P<logtime>\S+) (?P<status>\d+) (?P<loglevel>\w+)$'
    ignore_blank: true
    remove_fields: ['message']
```

=======
### Grok

>>>>>>> 042582bffa7582ac24b223f13d0690699821ec5b
### IPIP

### Json

### LinkMetric

### LinkStatsMetric

### Lowercase

### Remove

### Rename

### Split

```
Split:
  src: message
  sep: "\t"
  maxSplit: -1
  fields: ['logtime', 'hostname', 'uri', 'return_code']
<<<<<<< HEAD
  ignore_blank: true
=======
  ignoreBlank: true
>>>>>>> 042582bffa7582ac24b223f13d0690699821ec5b
  overwrite: true
```

#### src

数据来源字段, 默认 message , 如果字段不存在, 返回false

#### sep

分隔符, 在 strings.SplitN(src, sep, maxSplit) 中用被调用. 必须配置.

如果分隔符包含不可见字符, yaml配置以及gohangout也是支持的, 像下面这样

```
sep: "\x01"
```

#### maxSplit

在 strings.SplitN(src, sep, maxSplit) 中用被调用, 默认 -1, 代表无限制

#### fields

如果分割后的字符串数组长度与 fields 长度不一样, 返回false

<<<<<<< HEAD
#### ignore_blank
=======
#### ignoreBlank
>>>>>>> 042582bffa7582ac24b223f13d0690699821ec5b

如果分割后的某字段为空, 刚不放后 event 中, 默认 true

### Uppercase
