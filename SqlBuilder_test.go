package masala

import (
	"github.com/jinzhu/configor"
	"testing"
	"time"
)

var mockConfig = map[string]string{"test":"fc_transactions"}

/** go test -v masala */
func TestCriteria(test *testing.T) {
	test.Run("FetchAll", func(test *testing.T) {
		config := &Config{}
		configor.Load(&config, "config.test.yml")
		builder := NewSqlBuilder(config)
		if table, ok := mockConfig["test"]; ok {
			var state State
			state.Where = map[string]interface{}{"date >":time.Now().String()}
			state.Paginator.Current = 1
			query, arguments := builder.Table(table).Select("date").Criteria(&state)
			rows := builder.AsyncFetchAll(query, arguments)
			for rows.Next() {
				test.Fatal("There should be no future dates")
			}
		} else {
			test.Fatal("No test table find in default configuration.")
		}

	})
}

func TestAsynFetchAll(test *testing.T) {
	test.Run("AsynFetchAll", func(test *testing.T) {
		config := &Config{}
		configor.Load(&config, "config.test.yml")
		builder := NewSqlBuilder(config)

		if table, ok := mockConfig["test"]; ok {
			rows := builder.AsyncFetchAll("SELECT * FROM " + table + " WHERE date > ? LIMIT ? OFFSET ?", []interface{}{time.Now().String(), 10, 0})
			for rows.Next() {
				test.Fatal("There should be no future dates")
			}
		} else {
			test.Fatal("No test table find in default configuration.")
		}

	})
}

func TestRows(test *testing.T) {
	test.Run("FetchAll", func(test *testing.T) {
		config := &Config{}
		configor.Load(&config, "config.test.yml")
		builder := NewSqlBuilder(config)
		if table, ok := mockConfig["test"]; ok {
			var state State
			state.Where = map[string]interface{}{"date >":time.Now().String()}
			state.Paginator.Current = 1
			builder.Table(table).Select("date")
			builder.Rows(10, &state)
			if len(state.Rows) > 0 {
				test.Fatal("There should be no future dates")
			}
		} else {
			test.Fatal("No test table find in default configuration.")
		}

	})
}
