package output

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	clickhouse "github.com/ClickHouse/clickhouse-go"
	"github.com/IBM/sarama"
	"github.com/childe/gohangout/topology"
	"github.com/childe/gohangout/value_render"
	"github.com/spf13/cast"
	"k8s.io/klog/v2"
)

const (
	CLICKHOUSE_DEFAULT_BULK_ACTIONS   = 1000
	CLICKHOUSE_DEFAULT_FLUSH_INTERVAL = 30
)

type ClickhouseOutput struct {
	config map[interface{}]interface{}

	bulk_actions int
	hosts        []string
	fields       []string
	table        string
	username     string
	password     string

	fieldsLength int
	query        string
	desc         map[string]*rowDesc
	defaultValue map[string]interface{} // columnName -> defaultValue

	bulkChan   chan []map[string]interface{}
	concurrent int

	events       []map[string]interface{}
	execution_id uint64

	dbSelector HostSelector

	mux       sync.Mutex
	wg        sync.WaitGroup
	closeChan chan bool

	autoConvert         bool
	transIntColumn      []string
	transFloatColumn    []string
	transIntArrayColumn []string

	reliableCommit   bool
	kafkaFieldRender value_render.ValueRender // get kafka offset info from render
}

type rowDesc struct {
	Name              string `json:"name"`
	Type              string `json:"type"`
	DefaultType       string `json:"default_type"`
	DefaultExpression string `json:"default_expression"`
}

func (c *ClickhouseOutput) setTableDesc() {
	c.desc = make(map[string]*rowDesc)

	query := fmt.Sprintf("desc table %s", c.table)
	klog.V(5).Info(query)

	for i := 0; i < c.dbSelector.Size(); i++ {
		nextdb := c.dbSelector.Next()

		db := nextdb.(*sql.DB)

		rows, err := db.Query(query)
		if err != nil {
			klog.Errorf("query %q error: %s", query, err)
			continue
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			klog.Fatalf("could not get columns from query `%s`: %s", query, err)
		}
		klog.V(10).Infof("desc table columns: %v", columns)

		descMap := make(map[string]string)
		for _, c := range columns {
			descMap[c] = ""
		}

		for rows.Next() {
			values := make([]interface{}, 0)
			for range columns {
				var a string
				values = append(values, &a)
			}

			if err := rows.Scan(values...); err != nil {
				klog.Fatalf("scan rows error: %s", err)
			}

			descMap := make(map[string]string)
			for i, c := range columns {
				value := *values[i].(*string)
				if c == "type" {
					// 特殊处理枚举类型
					if strings.HasPrefix(value, "Enum16") {
						value = "Enum16"
					} else if strings.HasPrefix(value, "Enum8") {
						value = "Enum8"
					}
				}
				descMap[c] = value
			}

			b, err := json.Marshal(descMap)
			if err != nil {
				klog.Fatalf("marshal desc error: %s", err)
			}

			rowDesc := rowDesc{}
			err = json.Unmarshal(b, &rowDesc)
			if err != nil {
				klog.Fatalf("marshal desc error: %s", err)
			}

			klog.V(5).Infof("row desc: %#v", rowDesc)

			c.desc[rowDesc.Name] = &rowDesc
		}

		for key1, value1 := range c.desc {
			switch value1.Type {
			case "Int64", "UInt64", "Int32", "UInt32", "Int16", "UInt16", "Int8", "UInt8", "Nullable(Int64)", "Nullable(Int32)", "Nullable(Int16)", "Nullable(Int8)":
				c.transIntColumn = append(c.transIntColumn, key1)
			case "Array(Int64)", "Array(Int32)", "Array(Int16)", "Array(Int8)":
				c.transIntArrayColumn = append(c.transIntArrayColumn, key1)
			case "Float64", "Float32", "Nullable(Float32)", "Nullable(Float64)":
				c.transFloatColumn = append(c.transFloatColumn, key1)
			}
		}

		if len(c.fields) == 0 {
			for key1 := range c.desc {
				c.fields = append(c.fields, key1)
			}
		}
		return
	}
}

// TODO only string, number and ip DEFAULT expression is supported for now
func (c *ClickhouseOutput) setColumnDefault() {
	c.setTableDesc()

	c.defaultValue = make(map[string]interface{})

	var defaultValue *string

	for columnName, d := range c.desc {
		switch d.DefaultType {
		case "DEFAULT":
			defaultValue = &(d.DefaultExpression)
		case "MATERIALIZED":
			klog.Fatal("parse default value: MATERIALIZED expression not supported")
		case "ALIAS":
			klog.Fatal("parse default value: ALIAS expression not supported")
		case "":
			defaultValue = nil
		default:
			klog.Fatal("parse default value: only DEFAULT expression supported")
		}

		switch d.Type {
		case "String", "LowCardinality(String)":
			if defaultValue == nil {
				c.defaultValue[columnName] = ""
			} else {
				c.defaultValue[columnName] = *defaultValue
			}
		case "Date", "DateTime", "DateTime64":
			c.defaultValue[columnName] = time.Unix(0, 0)
		case "Nullable(Int64)", "Nullable(Int32)", "Nullable(Int16)", "Nullable(Int8)", "Nullable(Float32)", "Nullable(Float64)":
			c.defaultValue[columnName] = nil
		case "UInt8", "UInt16", "UInt32", "UInt64", "Int8", "Int16", "Int32", "Int64":
			if defaultValue == nil {
				c.defaultValue[columnName] = 0
			} else {
				i, e := strconv.ParseInt(*defaultValue, 10, 64)
				if e == nil {
					c.defaultValue[columnName] = i
				} else {
					klog.Fatalf("parse default value `%v` error: %v", defaultValue, e)
				}
			}
		case "Float32", "Float64":
			if defaultValue == nil {
				c.defaultValue[columnName] = 0.0
			} else {
				i, e := strconv.ParseFloat(*defaultValue, 64)
				if e == nil {
					c.defaultValue[columnName] = i
				} else {
					klog.Fatalf("parse default value `%v` error: %v", defaultValue, e)
				}
			}
		case "IPv4":
			c.defaultValue[columnName] = "0.0.0.0"
		case "IPv6":
			c.defaultValue[columnName] = "::"
		case "Array(String)", "Array(IPv4)", "Array(IPv6)", "Array(Date)", "Array(DateTime)":
			c.defaultValue[columnName] = clickhouse.Array([]string{})
		case "Array(UInt8)":
			c.defaultValue[columnName] = clickhouse.Array([]uint8{})
		case "Array(UInt16)":
			c.defaultValue[columnName] = clickhouse.Array([]uint16{})
		case "Array(UInt32)":
			c.defaultValue[columnName] = clickhouse.Array([]uint32{})
		case "Array(UInt64)":
			c.defaultValue[columnName] = clickhouse.Array([]uint64{})
		case "Array(Int8)":
			c.defaultValue[columnName] = clickhouse.Array([]int8{})
		case "Array(Int16)":
			c.defaultValue[columnName] = clickhouse.Array([]int16{})
		case "Array(Int32)":
			c.defaultValue[columnName] = clickhouse.Array([]int32{})
		case "Array(Int64)":
			c.defaultValue[columnName] = clickhouse.Array([]int64{})
		case "Array(Float32)":
			c.defaultValue[columnName] = clickhouse.Array([]float32{})
		case "Array(Float64)":
			c.defaultValue[columnName] = clickhouse.Array([]float64{})
		case "Enum16":
			// 需要要求列声明的最小枚举值为 ''
			c.defaultValue[columnName] = ""
		case "Enum8":
			// 需要要求列声明的最小枚举值为 ''
			c.defaultValue[columnName] = ""
		default:
			klog.Errorf("column: %s, type: %s. unsupported column type, ignore.", columnName, d.Type)
			continue
		}

	}
}

func (c *ClickhouseOutput) getDatabase() string {
	dbAndTable := strings.Split(c.table, ".")
	dbName := "default"
	if len(dbAndTable) == 2 {
		dbName = dbAndTable[0]
	}
	return dbName
}

func init() {
	Register("Clickhouse", newClickhouseOutput)
}

func newClickhouseOutput(config map[interface{}]interface{}) topology.Output {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	p := &ClickhouseOutput{
		config: config,
	}

	if v, ok := config["fields"]; ok {
		for _, f := range v.([]interface{}) {
			p.fields = append(p.fields, f.(string))
		}
	}

	if v, ok := config["reliable_commit"]; ok {
		p.reliableCommit = v.(bool)
	} else {
		p.reliableCommit = true
	}
	if v, ok := config["kafka_meta_field"].(string); ok {
		p.kafkaFieldRender = value_render.GetValueRender(v)
	}

	if v, ok := config["auto_convert"]; ok {
		p.autoConvert = v.(bool)
	} else {
		p.autoConvert = true
	}

	if v, ok := config["table"]; ok {
		p.table = v.(string)
	} else {
		klog.Fatalf("table must be set in clickhouse output")
	}

	if v, ok := config["hosts"]; ok {
		for _, h := range v.([]interface{}) {
			p.hosts = append(p.hosts, h.(string))
		}
	} else {
		klog.Fatalf("hosts must be set in clickhouse output")
	}

	if v, ok := config["username"]; ok {
		p.username = v.(string)
	}

	if v, ok := config["password"]; ok {
		p.password = v.(string)
	}

	debug := false
	if v, ok := config["debug"]; ok {
		debug = v.(bool)
	}

	connMaxLifetime := 0
	if v, ok := config["conn_max_life_time"]; ok {
		connMaxLifetime = v.(int)
	}

	dbs := make([]*sql.DB, 0)

	for _, host := range p.hosts {
		dataSourceName := fmt.Sprintf("%s?database=%s&username=%s&password=%s&debug=%v", host, p.getDatabase(), p.username, p.password, debug)
		if db, err := sql.Open("clickhouse", dataSourceName); err == nil {
			if err := db.Ping(); err != nil {
				if exception, ok := err.(*clickhouse.Exception); ok {
					klog.Errorf("[%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
				} else {
					klog.Errorf("clickhouse ping error: %s", err)
				}
			} else {
				db.SetConnMaxLifetime(time.Second * time.Duration(connMaxLifetime))
				dbs = append(dbs, db)
			}
		} else {
			klog.Errorf("open %s error: %s", host, err)
		}
	}

	klog.V(5).Infof("%d available clickhouse hosts", len(dbs))
	if len(dbs) == 0 {
		klog.Fatal("no available host")
	}

	dbsI := make([]interface{}, len(dbs))
	for i, h := range dbs {
		dbsI[i] = h
	}
	p.dbSelector = NewRRHostSelector(dbsI, 3)

	p.setColumnDefault()
	if len(p.fields) <= 0 {
		klog.Fatalf("fields not set in clickhouse output and could get fields from clickhouse table")
	}
	p.fieldsLength = len(p.fields)

	fields := make([]string, p.fieldsLength)
	for i := range fields {
		fields[i] = fmt.Sprintf(`"%s"`, p.fields[i])
	}
	questionMarks := make([]string, p.fieldsLength)
	for i := 0; i < p.fieldsLength; i++ {
		questionMarks[i] = "?"
	}
	p.query = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", p.table, strings.Join(fields, ","), strings.Join(questionMarks, ","))
	klog.V(5).Infof("query: %s", p.query)

	concurrent := 1
	if v, ok := config["concurrent"]; ok {
		concurrent = v.(int)
	}
	p.concurrent = concurrent
	p.closeChan = make(chan bool, concurrent)

	p.bulkChan = make(chan []map[string]interface{}, concurrent)
	for i := 0; i < concurrent; i++ {
		p.wg.Add(1)
		go func() {
			for {
				select {
				case events := <-p.bulkChan:
					p.innerFlush(events)
				case <-p.closeChan:
					p.wg.Done()
					return
				}
			}
		}()
	}

	if v, ok := config["bulk_actions"]; ok {
		p.bulk_actions = v.(int)
	} else {
		p.bulk_actions = CLICKHOUSE_DEFAULT_BULK_ACTIONS
	}

	var flush_interval int
	if v, ok := config["flush_interval"]; ok {
		flush_interval = v.(int)
	} else {
		flush_interval = CLICKHOUSE_DEFAULT_FLUSH_INTERVAL
	}
	go func() {
		for range time.NewTicker(time.Second * time.Duration(flush_interval)).C {
			p.flush()
		}
	}()

	return p
}

// convert int and float fields to coresponding type
func (c *ClickhouseOutput) convert(event map[string]interface{}) {
	for _, key := range c.transIntColumn {
		if keyIntValue, ok := event[key]; ok {
			if intConverterValue, err := cast.ToInt64E(keyIntValue); err == nil {
				event[key] = intConverterValue
			} else {
				klog.V(10).Infof("ch_output convert intType error: %s", err)
				event[key] = nil
			}
		}
	}

	for _, key := range c.transIntArrayColumn {
		if keyArrayValue, ok := event[key]; ok {
			arrayIntValue := keyArrayValue.([]interface{})
			ints := make([]int64, len(arrayIntValue))
			for i, v := range arrayIntValue {
				if v, err := cast.ToInt64E(v); err == nil {
					ints[i] = v
				} else {
					klog.V(10).Infof("ch_output convert arrayIntType error: %s", err)
					ints[i] = 0
				}
				event[key] = ints
			}
		}
	}

	for _, key := range c.transFloatColumn {
		if keyFloatValue, ok := event[key]; ok {
			floatConverterValue, err := cast.ToFloat64E(keyFloatValue)
			if err == nil {
				event[key] = floatConverterValue
			} else {
				klog.V(10).Infof("ch_output convert floatType error: %s", err)
				event[key] = nil
			}
		}
	}
}

func (c *ClickhouseOutput) innerFlush(events []map[string]interface{}) {
	execution_id := atomic.AddUint64(&c.execution_id, 1)
	klog.Infof("write %d docs to clickhouse with execution_id %d", len(events), execution_id)

	for {
		nextdb := c.dbSelector.Next()

		/*** not ReduceWeight for now , so this should not happen
		if nextdb == nil {
			klog.Info("no available db, wait for 30s")
			time.Sleep(30 * time.Second)
			continue
		}
		****/

		tx, err := nextdb.(*sql.DB).Begin()
		if err != nil {
			klog.Errorf("db begin to create transaction error: %s", err)
			continue
		}
		defer tx.Rollback()

		stmt, err := tx.Prepare(c.query)
		if err != nil {
			klog.Errorf("transaction prepare statement error: %s", err)
			return
		}
		defer stmt.Close()

		for _, event := range events {
			if c.autoConvert {
				c.convert(event)
			}

			args := make([]interface{}, c.fieldsLength)
			for i, field := range c.fields {
				if v, ok := event[field]; ok && v != nil {
					args[i] = v
				} else {
					if vv, ok := c.defaultValue[field]; ok {
						args[i] = vv
					} else { // this should not happen
						args[i] = ""
					}
				}
			}

			if _, err := stmt.Exec(args...); err != nil {
				klog.Errorf("exec clickhouse insert %v error: %s", event, err)
				return
			}
		}

		if err := tx.Commit(); err != nil {
			klog.Errorf("exec clickhouse commit error: %s", err)
			return
		}
		klog.Infof("%d docs has been committed to clickhouse", len(events))

		for _, event := range events {
			kafkaMeta, ok := c.kafkaFieldRender.Render(event).(map[string]interface{})
			if !ok {
				klog.Error("kafka meta field not found, can not commit event: ", event)
				continue
			}
			err := CommitKafkaEvent(kafkaMeta)
			if err != nil {
				klog.Errorf("commit kafka event error: %s, evnet: %v", err, kafkaMeta)
			}
		}
		return
	}
}

func (c *ClickhouseOutput) flush() {
	c.mux.Lock()
	if len(c.events) > 0 {
		events := c.events
		c.events = make([]map[string]interface{}, 0, c.bulk_actions)
		c.bulkChan <- events
	}
	c.mux.Unlock()
}

// Emit appends event to c.events, and push to bulkChan if needed
func (c *ClickhouseOutput) Emit(event map[string]interface{}) {
	c.mux.Lock()
	c.events = append(c.events, event)
	if len(c.events) < c.bulk_actions {
		c.mux.Unlock()
		return
	}

	events := c.events
	c.events = make([]map[string]interface{}, 0, c.bulk_actions)
	c.mux.Unlock()

	c.bulkChan <- events
}

func (c *ClickhouseOutput) awaitclose(timeout time.Duration) {
	exit := make(chan bool)
	defer func() {
		select {
		case <-exit:
			klog.Info("all clickhouse flush job done. return")
			return
		case <-time.After(timeout):
			klog.Info("clickhouse await timeout. return")
			return
		}
	}()

	defer func() {
		go func() {
			c.wg.Wait()
			exit <- true
		}()
	}()

	klog.Info("try to write remaining docs to clickhouse")

	c.mux.Lock()
	if len(c.events) <= 0 {
		klog.Info("no docs remain, return")
		c.mux.Unlock()
	} else {
		events := c.events
		c.events = make([]map[string]interface{}, 0, c.bulk_actions)
		c.mux.Unlock()

		klog.Infof("ramain %d docs, write them to clickhouse", len(events))
		c.wg.Add(1)
		go func() {
			c.innerFlush(events)
			c.wg.Done()
		}()
	}

	klog.Info("check if there are events blocking in bulk channel")

	for {
		select {
		case events := <-c.bulkChan:
			c.wg.Add(1)
			go func() {
				c.innerFlush(events)
				c.wg.Done()
			}()
		default:
			return
		}
	}
}

// Shutdown would stop receiving message and emiting
func (c *ClickhouseOutput) Shutdown() {
	for i := 0; i < c.concurrent; i++ {
		c.closeChan <- true
	}
	c.awaitclose(30 * time.Second)
}

func CommitKafkaEvent(kafkaMeta map[string]interface{}) error {
	session, ok := kafkaMeta["session"].(sarama.ConsumerGroupSession)
	if !ok {
		return fmt.Errorf("kafka session field not found, can not commit")
	}
	if err := session.Context().Err(); err != nil {
		return fmt.Errorf("kafka session context error: %v", err)
	}
	topic, ok := kafkaMeta["topic"].(string)
	if !ok {
		return fmt.Errorf("kafka topic field not found, can not commit")
	}
	partition, ok := kafkaMeta["partition"].(int32)
	if !ok {
		return fmt.Errorf("kafka partition field not found, can not commit")
	}
	offset, ok := kafkaMeta["offset"].(int64)
	if !ok {
		return fmt.Errorf("kafka offset field not found, can not commit")
	}
	session.MarkOffset(topic, partition, offset+1, "")
	glog.V(10).Infof("commit offset: topic: %s, partition: %d, offset: %d", topic, partition, offset)
	return nil
}
