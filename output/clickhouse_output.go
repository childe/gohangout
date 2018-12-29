package output

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"

	"github.com/kshvakov/clickhouse"
)

const (
	CLICKHOUSE_DEFAULT_BULK_ACTIONS   = 1000
	CLICKHOUSE_DEFAULT_FLUSH_INTERVAL = 30
)

type ClickhouseOutput struct {
	BaseOutput
	config map[interface{}]interface{}

	bulk_actions int
	hosts        []string
	fields       []string
	table        string
	username     string
	password     string

	fieldsLength int
	query        string
	desc         map[string]string      // columnName -> Type
	defaultValue map[string]interface{} // column -> defaultValue

	bulkChan chan []map[string]interface{}

	events []map[string]interface{}

	dbSelector HostSelector

	wg sync.WaitGroup
}

func (c *ClickhouseOutput) setTableDesc() {
	c.desc = make(map[string]string)

	query := fmt.Sprintf("desc table %s", c.table)
	glog.V(5).Info(query)

	for {
		nextdb := c.dbSelector.next()
		if nextdb == nil {
			glog.Fatalf("could not get desc from %s", c.table)
		}

		db := nextdb.(*sql.DB)

		rows, err := db.Query(query)
		if err != nil {
			glog.Errorf("query %q error: %s", query, err)
			c.dbSelector.reduceWeight()
			continue
		}
		defer rows.Close()

		var (
			column string
			typ    string
			def    string
			exp    string
		)
		for rows.Next() {

			if err := rows.Scan(&column, &typ, &def, &exp); err != nil {
				glog.Errorf("scan rows error: %s", err)
				continue
			}
			glog.V(5).Infof("%q %q %q %q", column, typ, def, exp)

			c.desc[column] = typ
		}

		return
	}
}

func (c *ClickhouseOutput) setColumnDefault() {
	c.setTableDesc()

	c.defaultValue = make(map[string]interface{})

	for columnName, typ := range c.desc {
		switch typ {
		case "String":
			c.defaultValue[columnName] = ""
		case "Date", "DateTime":
			c.defaultValue[columnName] = time.Unix(0, 0)
		case "UInt8", "UInt16", "UInt32", "UInt64", "Int8", "Int16", "Int32", "Int64":
			c.defaultValue[columnName] = 0
		case "Float32", "Float64":
			c.defaultValue[columnName] = 0.0
		case "Array(String)":
			c.defaultValue[columnName] = []string{}
		default:
			glog.Errorf("column: %s, type: %s. unsupported column type, ignore", columnName, typ)
			continue
		}
	}
}

func NewClickhouseOutput(config map[interface{}]interface{}) *ClickhouseOutput {
	rand.Seed(time.Now().UnixNano())
	p := &ClickhouseOutput{
		BaseOutput: NewBaseOutput(config),
		config:     config,
	}

	if v, ok := config["table"]; ok {
		p.table = v.(string)
	} else {
		glog.Fatalf("table must be set in clickhouse output")
	}

	if v, ok := config["hosts"]; ok {
		for _, h := range v.([]interface{}) {
			p.hosts = append(p.hosts, h.(string))
		}
	} else {
		glog.Fatalf("hosts must be set in clickhouse output")
	}

	if v, ok := config["username"]; ok {
		p.username = v.(string)
	}

	if v, ok := config["password"]; ok {
		p.password = v.(string)
	}

	if v, ok := config["fields"]; ok {
		for _, f := range v.([]interface{}) {
			p.fields = append(p.fields, f.(string))
		}
	} else {
		glog.Fatalf("fields must be set in clickhouse output")
	}
	if len(p.fields) <= 0 {
		glog.Fatalf("fields length must be > 0")
	}
	p.fieldsLength = len(p.fields)

	fields := make([]string, p.fieldsLength)
	for i, _ := range fields {
		fields[i] = fmt.Sprintf(`"%s"`, p.fields[i])
	}
	questionMarks := make([]string, p.fieldsLength)
	for i := 0; i < p.fieldsLength; i++ {
		questionMarks[i] = "?"
	}
	p.query = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", p.table, strings.Join(fields, ","), strings.Join(questionMarks, ","))
	glog.V(5).Infof("query: %s", p.query)

	dbs := make([]*sql.DB, 0)
	for _, host := range p.hosts {
		dataSourceName := fmt.Sprintf("%s?username=%s&password=%s", host, p.username, p.password)
		if db, err := sql.Open("clickhouse", dataSourceName); err == nil {
			if err := db.Ping(); err != nil {
				if exception, ok := err.(*clickhouse.Exception); ok {
					glog.Errorf("[%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
				} else {
					glog.Errorf("clickhouse ping error: %s", err)
				}
			} else {
				dbs = append(dbs, db)
			}
		} else {
			glog.Error("open %s error: %s", host, err)
		}
	}

	if len(dbs) == 0 {
		glog.Fatal("no available host")
	}

	dbsI := make([]interface{}, len(dbs))
	for i, h := range dbs {
		dbsI[i] = h
	}
	p.dbSelector = NewRRHostSelector(dbsI, 3)

	p.setColumnDefault()

	concurrent := 1
	if v, ok := config["concurrent"]; ok {
		concurrent = v.(int)
	}

	p.bulkChan = make(chan []map[string]interface{}, concurrent)
	for i := 0; i < concurrent; i++ {
		go func() {
			for {
				events := <-p.bulkChan
				p.innerFlush(events)
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
			p.Flush()
		}
	}()

	return p
}

func (p *ClickhouseOutput) innerFlush(events []map[string]interface{}) {
	if len(events) == 0 {
		return
	}

	glog.Infof("write %d docs to clickhouse", len(events))

	p.wg.Add(1)
	defer p.wg.Done()

	for {
		nextdb := p.dbSelector.next()
		if nextdb == nil {
			glog.Info("no available db, wait for 30s")
			time.Sleep(30 * time.Second)
			continue
		}

		tx, err := nextdb.(*sql.DB).Begin()
		if err != nil {
			glog.Errorf("db begin to create transaction error: %s", err)
			p.dbSelector.reduceWeight()
			continue
		}
		defer tx.Rollback()

		stmt, err := tx.Prepare(p.query)
		if err != nil {
			glog.Errorf("transaction prepare statement error: %s", err)
			p.dbSelector.reduceWeight()
			continue
		}
		defer stmt.Close()

		for _, event := range events {
			args := make([]interface{}, p.fieldsLength)
			for i, field := range p.fields {
				if v, ok := event[field]; ok && v != nil {
					args[i] = v
				} else {
					if vv, ok := p.defaultValue[field]; ok {
						args[i] = vv
					} else { // this should not happen
						args[i] = ""
					}
				}
			}
			if _, err := stmt.Exec(args...); err != nil {
				glog.Errorf("exec clickhouse insert %v error: %s", event, err)
			}
		}

		if err := tx.Commit(); err != nil {
			glog.Errorf("exec clickhouse commit error: %s", err)
			p.dbSelector.reduceWeight()
			continue
		}
		glog.Infof("%d docs has been committed to clickhouse", len(events))
	}
}

func (p *ClickhouseOutput) Flush() {
	events := p.events
	p.events = make([]map[string]interface{}, 0, p.bulk_actions)
	if len(events) > 0 {
		p.bulkChan <- events
	}
}

func (p *ClickhouseOutput) Emit(event map[string]interface{}) {
	p.events = append(p.events, event)

	if len(p.events) >= p.bulk_actions {
		events := p.events
		p.events = make([]map[string]interface{}, 0, p.bulk_actions)
		p.bulkChan <- events
	}
}

func (p *ClickhouseOutput) awaitclose(timeout time.Duration) {
	c := make(chan bool)
	defer func() {
		select {
		case <-c:
			glog.Info("all clickhouse flush job done. return")
			return
		case <-time.After(timeout):
			glog.Info("clickhouse await timeout. return")
			return
		}
	}()

	defer func() {
		go func() {
			p.wg.Wait()
			c <- true
		}()
	}()

	p.wg.Add(1)
	go func() {
		p.Flush()
		p.wg.Done()
	}()
}

func (p *ClickhouseOutput) Shutdown() {
	p.awaitclose(30 * time.Second)
}
