package masala

import (
	"testing"
	"time"
)

/** go test -v masala */
func TestAsynFetchAll(test *testing.T) {
	test.Run("AsynFetchAll", func(test *testing.T) {
		builder := NewSqlBuilder()
		if table, ok := builder.Config.Tables["test"]; ok {
			var state State
			state.Where = map[string]interface{}{"date >":time.Now().String()}
			builder.Table(table).Select("date").Criteria(&state)
			rows := builder.AsyncFetchAll("LIMIT ? OFFSET ?", []interface{}{10, 0})
			for rows.Next() {
				test.Fatal("There should be no future dates")
			}
		} else {
			test.Fatal("No test table find in default configuration.")
		}

	})
}
