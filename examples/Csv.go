package examples

import (
	"bytes"
	"encoding/json"
	"fmt"
	"masala"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Export struct {
	builder *masala.SqlBuilder
	limit int
	logEvent *masala.Logger
	offset chan int
	source string
}

var limit = 1000

func NewCsv(limit int, source string) *Csv {
	builder := masala.NewSqlBuilder()
	return &Export{builder,
		limit,
		masala.NewLogger(),
		make(chan int),
		builder.Config.Tables[source]}
}

func (message *Csv) Data(parameters url.Values, buffer *bytes.Buffer) {
	var state CsvState
	json.Unmarshal([]byte(parameters["masala"][0]), &state)
	message.builder.Table(message.source).Select("myColumn").Criteria(&state)
	data, _ := json.Marshal(state)
	output := string(data)
	buffer.WriteString(fmt.Sprintf("data: %s\n", strings.Replace(output, "\n", "\ndata: ", -1)))
	sum := state.GetPaginator().Sum
	go message.send(sum)
	var group sync.WaitGroup
	csv, err := os.Create("output.csv")
	message.logEvent.Error(err)
	defer csv.Close()
	jobs := sum / limit
	if 0 != sum % limit {
		jobs++
	}
	jobsPerRoutine := 2
	for jobs > 0 {
		group.Add(1)
		if jobsPerRoutine > jobs {
			go message.listen(csv, &group, jobs)
		} else {
			go message.listen(csv, &group, jobsPerRoutine)
		}
		jobs = jobs - jobsPerRoutine
	}
	group.Wait()
}

func (message *Export) send(last int) {
	start := 0
	for start < last {
		message.offset <- start
		start = start + limit
	}
}

func (message *Export) Id() string {
	return "Export"
}

func (message *Export) listen(csv *os.File, group *sync.WaitGroup, jobs int) {
	defer group.Done()
	job := 0
	for job < jobs {
		select {
		case offset := <-message.offset:
			rows := message.builder.AsyncFetchAll("LIMIT ? OFFSET ?", []interface{}{limit, offset})
			var id int64
			var date *time.Time
			var price float32
			for rows.Next() {
				err := rows.Scan(&id, &date, &price)
				message.logEvent.Panic(err)
				csv.Write([]byte(strconv.FormatInt(id, 10) + "\n"))
				offset++
			}
			rows.Close()
			job++
		default:
			/** fmt.Println("no message received") */
		}
	}
}