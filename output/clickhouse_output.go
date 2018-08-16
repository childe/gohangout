package output

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"

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
	fieldsLength int
	query        string

	events []map[string]interface{}

	db *sql.DB

	mux       sync.Mutex
	wg        sync.WaitGroup
	semaphore *semaphore.Weighted
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

	n := rand.Int() % len(p.hosts)
	host := p.hosts[n]
	db, err := sql.Open("clickhouse", host)
	if err == nil {
		p.db = db
	} else {
		glog.Fatalf("open %s error: %s", host, err)
	}

	if err := db.Ping(); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			glog.Fatalf("[%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		} else {
			glog.Fatalf("clickhouse ping error: %s", err)
		}
	}

	if v, ok := config["concurrent"]; ok {
		p.semaphore = semaphore.NewWeighted(int64(v.(int)))
	} else {
		p.semaphore = semaphore.NewWeighted(1)
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
	glog.Infof("write %d docs to clickhouse", len(events))
	if len(events) == 0 {
		return
	}

	tx, err := p.db.Begin()
	if err != nil {
		glog.Errorf("db begin to create transaction error: %s", err)
		return
	}
	stmt, err := tx.Prepare(p.query)
	if err != nil {
		glog.Errorf("transaction prepare statement error: %s", err)
		return
	}
	defer stmt.Close()

	for _, event := range events {
		args := make([]interface{}, p.fieldsLength)
		for i, field := range p.fields {
			if v, ok := event[field]; ok {
				args[i] = v
			} else {
				args[i] = ""
			}
		}
		if _, err := stmt.Exec(args...); err != nil {
			glog.Errorf("exec clickhouse insert %v error: %s", event, err)
		}
	}

	if err := tx.Commit(); err != nil {
		glog.Errorf("exec clickhouse commit error: %s", err)
	}
	glog.Infof("%d docs has been committed to clickhouse", len(events))
}

func (p *ClickhouseOutput) Flush() {
	defer p.wg.Done()
	defer p.semaphore.Release(1)
	defer p.mux.Unlock()

	p.mux.Lock()
	p.semaphore.Acquire(context.TODO(), 1)
	p.wg.Add(1)

	events := p.events
	p.events = make([]map[string]interface{}, 0, p.bulk_actions)
	p.innerFlush(events)
}

func (p *ClickhouseOutput) Emit(event map[string]interface{}) {
	p.events = append(p.events, event)

	if len(p.events) >= p.bulk_actions {
		p.Flush()
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
