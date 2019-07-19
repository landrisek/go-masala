package masala

import ("database/sql"
	"encoding/json"
	"github.com/jinzhu/configor"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings")

type Paginator struct {
	Current int
	Last int
	Sum int
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
	criteria map[string]string
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
	group string
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

func NewSqlBuilder() *SqlBuilder {
	builder := &SqlBuilder{}
	var file, err = os.OpenFile("../../log/builder.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if nil != err {
		log.Panic(err)
	}
	builder.logger = log.New(file, "", log.LstdFlags|log.Lshortfile)
	configor.Load(&builder.Config, "../config.yml")
	host, err := os.Hostname()
	builder.log(err)
	configor.Load(&builder.Config, "../config." + host + ".yml")
	builder.database, err = sql.Open("mysql", builder.Config.Database.User + ":" +
											builder.Config.Database.Password  + "@tcp(" +
											builder.Config.Database.Host + ":" + strconv.Itoa(builder.Config.Database.Port) + ")/" +
											builder.Config.Database.Name + "?parseTime=true&collation=utf8_czech_ci")
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

func(builder *SqlBuilder) AsyncFetchAll(subquery string, arguments []interface{}) *sql.Rows {
	for _, argument := range builder.arguments {
		arguments = append([]interface{}{argument}, arguments...)
	}
	query, err := builder.database.Prepare("SELECT " + builder.columns  + " FROM " + builder.table + " " + builder.query + " " + subquery)
	builder.log(err)
	rows, err := query.Query(arguments...)
	if nil != err {
		log.Panic(err)
	}
	return rows
}

func (builder *SqlBuilder) Criteria(state *State) {
	var where string
	for alias, value := range state.Where {
		regex, err := regexp.Compile("(>|<|=|\\s)")
		builder.log(err)
		column := regex.ReplaceAllString(alias, "")
		if _, ok := builder.criteria[column]; ok {
			where += alias + " ? AND "
			builder.arguments = append(builder.arguments, value)
		}
	}
	if len(where) > 0 {
		builder.query += "WHERE " + strings.TrimRight(where, "AND ")
	}
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

func(builder *SqlBuilder) Paginator(limit int, state *State) {
	builder.query = "SELECT COUNT(*) AS count FROM " + builder.table + " "
	builder.Criteria(state)
	builder.query = strings.TrimRight(builder.query, "AND ") + builder.group
	var count int
	builder.log(builder.database.QueryRow(builder.query, builder.arguments...).Scan(&count))
	state.Paginator.Sum = count
	state.Paginator.Last = count / limit
}

func(builder *SqlBuilder) Props(request *http.Request, translatorRepository ITranslator) map[string]interface{} {
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

func(builder *SqlBuilder) Rows(limit int, state *State) {
	builder.query = "SELECT " + builder.columns  + " FROM " + builder.table + " "
	builder.Criteria(state)
	offset := (state.Paginator.Current - 1) * 20
	builder.query = strings.TrimRight(builder.query, "AND ") + builder.group + " LIMIT " + strconv.Itoa(limit) + " OFFSET " + strconv.Itoa(offset)
	query, err := builder.database.Prepare(builder.query)
	builder.log(err)
	rows, err := query.Query(builder.arguments...)
	builder.log(err)
	columns, err := rows.Columns()
	builder.log(err)
	data := make([][]byte, len(columns))
	pointers := make([]interface{}, len(columns))
	for i := range data {
		pointers[i] = &data[i]
	}
	state.Rows = make([]map[string]string, 0)
	for rows.Next() {
		row := make(map[string]string, 0)
		rows.Scan(pointers...)
		for key := range data {
			row[columns[key]] = string(data[key])
		}
		state.Rows = append(state.Rows, row)
	}
	rows.Close()
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