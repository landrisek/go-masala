package masala

import ("database/sql"
	"encoding/json"
	"github.com/jinzhu/configor"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings")

type Paginator struct {
	Current string
	Last string
	Sum string
}

type State struct {
	Autocomplete struct {
		Data map[string]string
		Position int
	}
	Clicked map[string]bool
	Crops map[string]string
	Group string
	Menu string
	Order map[string]string
	Paginator Paginator
	Rows []map[string]string
	Where map[string]string
	Wysiwyg map[string]string
}

type SqlBuilder struct {
	arguments []interface{}
	database *sql.DB
	columns string
	Config struct {
		Database struct {
			Host string
			Name string
			Password string
			Port int
			User string
		}
		Scheme string
		Tables map[string]string
	}
	leftJoin []string
	logger *log.Logger
	rows *sql.Rows
	state State
	table string
	query string
}

type ITranslator interface {
	Translate(term string) string
}

func(builder SqlBuilder) Inject() SqlBuilder {
	configor.Load(&builder.Config, "../config.yml")
	host, _ := os.Hostname()
	configor.Load(&builder.Config, "../config." + host + ".yml")
	builder.database, _ = sql.Open("mysql", builder.Config.Database.User + ":" +
											builder.Config.Database.Password  + "@tcp(" +
											builder.Config.Database.Host + ":" + strconv.Itoa(builder.Config.Database.Port) + ")/" +
											builder.Config.Database.Name + "?parseTime=true&collation=utf8_czech_ci")
	var file, _ = os.OpenFile("../../log/builder.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	builder.logger = log.New(file, "", log.LstdFlags|log.Lshortfile)
	return builder
}

func(builder SqlBuilder) And(and string) SqlBuilder {
	builder.query = strings.Join([]string{builder.query, " AND ", and}, " ")
	return builder
}

func(builder SqlBuilder) Arguments(arguments []interface{}) SqlBuilder {
	builder.arguments = arguments
	return builder
}

func(builder SqlBuilder) FetchAll() []map[string]string {
	query, _ := builder.database.Prepare("SELECT " + builder.columns  + " FROM " + builder.table + " " + builder.query)
	rows, _ := query.Query(builder.arguments...)
	results := make([]map[string]string, 0)
	columns, _ := rows.Columns()
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

func(builder SqlBuilder) Fetch() *sql.Row {
	return builder.database.QueryRow("SELECT " + builder.columns  + " FROM " + builder.table + " " + builder.query, builder.arguments...)
}

func(builder SqlBuilder) GetState(request *http.Request) State {
	body, _ := ioutil.ReadAll(request.Body)
	json.Unmarshal(body, &builder.state)
	return builder.state
}

func(builder SqlBuilder) Group(group string) SqlBuilder {
	builder.query = strings.Join([]string{builder.query, " GROUP BY ", group}, " ")
	return builder
}

func(builder SqlBuilder) LeftJoin(join string) SqlBuilder {
	builder.query = strings.Join([]string{builder.query, " LEFT JOIN ", join}, " ")
	return builder
}

func(builder SqlBuilder) log(error error) {
	if nil != error {
		builder.logger.Output(2, error.Error())
	}
}

func(builder SqlBuilder) Limit(limit int) SqlBuilder {
	builder.query = strings.Join([]string{builder.query, " LIMIT ", strconv.Itoa(limit)}, " ")
	return builder
}

func(builder SqlBuilder) Insert(data map[string]interface{}) SqlBuilder {
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

func(builder SqlBuilder) Or(or string) SqlBuilder {
	builder.query = strings.Join([]string{builder.query, " OR ", or}, " ")
	return builder
}

func(builder SqlBuilder) Order(order string) SqlBuilder {
	builder.query = strings.Join([]string{builder.query, " ORDER BY ", order}, " ")
	return builder
}

func(builder SqlBuilder) Page(group string, limit int, request *http.Request) Paginator {
	body, _ := ioutil.ReadAll(request.Body)
	json.Unmarshal(body, &builder.state)
	builder.query = "SELECT COUNT(*) AS count FROM " + builder.table + " "
	if len(builder.state.Where) > 0 {
		builder.query += "WHERE "
	}
	for column, value := range builder.state.Where {
		builder.query += builder.table + "." + column + " = ? AND "
		builder.arguments = append(builder.arguments, value)
	}
	if len(group) > 0 {
		group = " GROUP BY " + group
	}
	builder.query = strings.TrimRight(builder.query, "AND ") + group
	var count int
	builder.log(builder.database.QueryRow(builder.query, builder.arguments...).Scan(&count))
	builder.state.Paginator.Sum = strconv.Itoa(count)
	builder.state.Paginator.Last = strconv.Itoa(count / limit)
	return builder.state.Paginator
}

func(builder SqlBuilder) Props(request *http.Request, translatorRepository ITranslator) map[string]interface{} {
	link := builder.Config.Scheme + "://" +  request.Host + request.URL.Path + "/"
	return map[string]interface{}{"download":map[string]string{"label":translatorRepository.Translate("Click here to download your file.")},
		"export":map[string]string{"label":translatorRepository.Translate("export"),"link":link + "export"},
		"Paginator":map[string]string{"link":link + "page",
			"next":translatorRepository.Translate("next"),
			"page":strings.Title(translatorRepository.Translate("page")),
			"previous":translatorRepository.Translate("previous"),
			"sum":translatorRepository.Translate("total")},
		"submit":map[string]string{"label":translatorRepository.Translate("filter data"),"link":link + "state"}}
}

func(builder SqlBuilder) Select(columns string) SqlBuilder {
	builder.columns = columns
	return builder
}

func(builder SqlBuilder) Set(set string) SqlBuilder {
	builder.query = strings.Join([]string{builder.query, " SET ", set}, " ")
	return builder
}

func(builder SqlBuilder) State(group string, limit int, request *http.Request) State {
	body, _ := ioutil.ReadAll(request.Body)
	json.Unmarshal(body, &builder.state)
	builder.query = "SELECT " + builder.columns  + " FROM " + builder.table + " "
	if len(builder.state.Where) > 0 {
		builder.query += "WHERE "
	}
	for column, value := range builder.state.Where {
		builder.query += builder.table + "." + column + " = ? AND "
		builder.arguments = append(builder.arguments, value)
	}
	current, _ := strconv.Atoi(builder.state.Paginator.Current)
	offset := (current - 1) * 20
	if len(builder.query) > 0 {
		group = " GROUP BY " + group
	}
	builder.query = strings.TrimRight(builder.query, "AND ") + group + " LIMIT " + strconv.Itoa(limit) + " OFFSET " + strconv.Itoa(offset)
	query, error := builder.database.Prepare(builder.query)
	builder.log(error)
	rows, error := query.Query(builder.arguments...)
	builder.log(error)
	columns, _ := rows.Columns()
	data := make([][]byte, len(columns))
	pointers := make([]interface{}, len(columns))
	for i := range data {
		pointers[i] = &data[i]
	}
	builder.state.Rows = make([]map[string]string, 0)
	for rows.Next() {
		row := make(map[string]string, 0)
		rows.Scan(pointers...)
		for key := range data {
			row[columns[key]] = string(data[key])
		}
		builder.state.Rows = append(builder.state.Rows, row)
	}
	rows.Close()
	return builder.state
}

func(builder SqlBuilder) Table(table string) SqlBuilder {
	builder.columns = " * "
	builder.table = table
	builder.query = ""
	return builder
}

func(builder SqlBuilder) Update() {
	query, error := builder.database.Prepare("UPDATE " + builder.table + builder.query)
	if nil != error {
		builder.logger.Output(2, error.Error())
	}
	query.Exec(builder.arguments...)
	query.Close()
}

func(builder SqlBuilder) Where(where string) SqlBuilder {
	builder.query = strings.Join([]string{builder.query, " WHERE ", where}, " ")
	return builder
}