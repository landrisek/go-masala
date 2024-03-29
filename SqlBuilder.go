package masala

import ("database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings")

type SqlBuilder struct {
	arguments []interface{}
	criteria map[string]string
	database *sql.DB
	columns string
	group string
	leftJoin []string
	logger *log.Logger
	table string
	query string
}

type ISqlBuilder interface {
	HealthCheck() bool
	Table() ISqlBuilder
}

type ITranslator interface {
	Translate(term string) string
}

func NewSqlBuilder(config IConfig) *SqlBuilder {
	builder := &SqlBuilder{}
	var file, err = os.OpenFile("builder.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if nil != err {
		log.Panic(err)
	}
	builder.logger = log.New(file, "", log.LstdFlags|log.Lshortfile)
	builder.database, err = sql.Open("mysql", config.GetUser() + ":" +
											config.GetPassword()  + "@tcp(" +
											config.GetHost() + ":" + strconv.Itoa(config.GetPort()) + ")/" +
											config.GetName() + "?parseTime=true&collation=utf8_czech_ci")
	builder.database.SetMaxIdleConns(20)
	builder.log(err)
	return builder
}

func(builder *SqlBuilder) And(and string) *SqlBuilder {
	builder.query = strings.Join([]string{builder.query, " AND ", and}, " ")
	return builder
}

func(builder *SqlBuilder) Arguments(arguments []interface{}) *SqlBuilder {
	builder.arguments = arguments
	return builder
}

func(builder *SqlBuilder) AsyncFetchAll(subquery string, criteria []interface{}) *sql.Rows {
	query, err := builder.database.Prepare(subquery)
	builder.log(err)
	rows, err := query.Query(criteria...)
	if nil != err {
		log.Panic(err)
	}
	return rows
}

func (builder *SqlBuilder) Criteria(state IState) (string, []interface{}) {
	regex, err := regexp.Compile("(>|<|=|\\s)")
	builder.log(err)
	var where string
	for alias, value := range state.GetCriteria() {
		column := regex.ReplaceAllString(alias, "")
		if _, ok := builder.criteria[column]; ok {
			where += alias + " ? AND "
			builder.arguments = append(builder.arguments, value)
		}
	}
	if len(where) > 0 {
		builder.query += "WHERE " + strings.TrimRight(where, "AND ")
	}
	builder.query += builder.group
	var order string
	for alias, asorting := range state.GetOrder() {
		column := regex.ReplaceAllString(alias, "")
		if _, ok := builder.criteria[column]; ok {
			order += " `" + alias + "` " + asorting + ", "
		}
	}
	if len(order) > 0 {
		builder.query += " ORDER BY " + strings.TrimRight(order, ", ")
	}
	return "SELECT " + builder.columns  + " FROM " + builder.table + " " + builder.query, builder.arguments
}

func(builder *SqlBuilder) FetchAll() []map[string]string {
	query, err := builder.database.Prepare("SELECT " + builder.columns  + " FROM " + builder.table + " " + builder.query)
	builder.log(err)
	rows, err := query.Query(builder.arguments...)
	builder.log(err)
	results := make([]map[string]string, 0)
	columns, err := rows.Columns()
	builder.log(err)
	data := make([][]byte, len(columns))
	pointers := make([]interface{}, len(columns))
	for i := range data {
		pointers[i] = &data[i]
	}
	for rows.Next() {
		row := make(map[string]string, 0)
		rows.Scan(pointers...)
		for key := range data {
			row[columns[key]] = string(data[key])
		}
		results = append(results, row)
	}
	rows.Close()
	return results
}

func(builder *SqlBuilder) Fetch() *sql.Row {
	return builder.database.QueryRow("SELECT " + builder.columns  + " FROM " + builder.table + " " + builder.query, builder.arguments...)
}

func(builder *SqlBuilder) GetState(request *http.Request) State {
	body, _ := ioutil.ReadAll(request.Body)
	var state State
	json.Unmarshal(body, &state)
	return state
}

func (builder *SqlBuilder) HealthCheck(query string, arguments []interface{}) bool {
	err := builder.database.Ping()
	builder.log(err)
	if nil != err {
		return false
	}
	if len(query) > 0 {
		healthCheck, err := builder.database.Prepare(query)
		builder.log(err)
		result, err := healthCheck.Exec(arguments...)
		builder.log(err)
		fmt.Print(result)
	}
	return true
}

func(builder *SqlBuilder) Insert(data map[string]interface{}) *SqlBuilder {
	var columns string
	var placeholders string
	for column, value := range data {
		builder.arguments = append(builder.arguments, value)
		columns += column + ", "
		placeholders += "?, "
	}
	columns = strings.TrimRight(columns, ", ")
	placeholders = strings.TrimRight(placeholders, ", ")
	builder.query = "INSERT INTO " + builder.table + "(" + columns + ") VALUES (" + placeholders + ") "
	query, error := builder.database.Prepare(builder.query)
	builder.log(error)
	query.Exec(builder.arguments...)
	query.Close()
	return builder
}

func(builder *SqlBuilder) Group(group string) *SqlBuilder {
	builder.group = " GROUP BY " + group + " "
	return builder
}

func(builder *SqlBuilder) LeftJoin(join string) *SqlBuilder {
	builder.query = strings.Join([]string{builder.query, " LEFT JOIN ", join}, " ")
	return builder
}

func(builder *SqlBuilder) log(error error) {
	if nil != error {
		builder.logger.Output(2, error.Error())
	}
}

func(builder *SqlBuilder) Limit(limit int) *SqlBuilder {
	builder.query = strings.Join([]string{builder.query, " LIMIT ", strconv.Itoa(limit)}, " ")
	return builder
}

func(builder *SqlBuilder) Or(or string) *SqlBuilder {
	builder.query = strings.Join([]string{builder.query, " OR ", or}, " ")
	return builder
}

func(builder *SqlBuilder) Order(order string) *SqlBuilder {
	builder.query = strings.Join([]string{builder.query, " ORDER BY ", order}, " ")
	return builder
}

func(builder *SqlBuilder) Paginator(limit int, state IState) {
	builder.query = "SELECT COUNT(*) AS count FROM " + builder.table + " "
	builder.Criteria(state)
	var count int
	builder.log(builder.database.QueryRow(builder.query, builder.arguments...).Scan(&count))
	paginator := state.GetPaginator()
	paginator.Sum = count
	paginator.Last = count / limit
	state.SetPaginator(paginator)
}

func(builder *SqlBuilder) Select(columns string) *SqlBuilder {
	builder.columns = columns
	builder.criteria = map[string]string{}
	criterias := strings.Split(columns, ",")
	for _, criteria := range criterias {
		builder.criteria[strings.TrimSpace(criteria)] = criteria
	}
	return builder
}

func(builder *SqlBuilder) Set(set string) *SqlBuilder {
	builder.query = strings.Join([]string{builder.query, " SET ", set}, " ")
	return builder
}

func(builder *SqlBuilder) Rows(limit int, state IState) {
	builder.query = "SELECT " + builder.columns  + " FROM " + builder.table + " "
	builder.Criteria(state)
	offset := (state.GetPaginator().Current - 1) * 20
	builder.query += " LIMIT " + strconv.Itoa(limit) + " OFFSET " + strconv.Itoa(offset)
	query, err := builder.database.Prepare(builder.query)
	builder.log(err)
	result, err := query.Query(builder.arguments...)
	builder.log(err)
	columns, err := result.Columns()
	builder.log(err)
	data := make([][]byte, len(columns))
	pointers := make([]interface{}, len(columns))
	for i := range data {
		pointers[i] = &data[i]
	}
	rows := make([]map[string]string, 0)
	for result.Next() {
		row := make(map[string]string, 0)
		result.Scan(pointers...)
		for key := range data {
			row[columns[key]] = string(data[key])
		}
		rows = append(rows, row)
	}
	state.SetRows(rows)
	result.Close()
}

func(builder *SqlBuilder) Table(table string) *SqlBuilder {
	builder.columns = "*"
	builder.table = table
	builder.query = ""
	return builder
}

func(builder *SqlBuilder) Update() {
	query, error := builder.database.Prepare("UPDATE " + builder.table + builder.query)
	if nil != error {
		builder.logger.Output(2, error.Error())
	}
	query.Exec(builder.arguments...)
	query.Close()
}

func(builder *SqlBuilder) Where(where string) *SqlBuilder {
	builder.query = strings.Join([]string{builder.query, " WHERE ", where}, " ")
	return builder
}