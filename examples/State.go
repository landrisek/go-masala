package examples

import ("masala")

type State struct {
        builder *masala.SqlBuilder
        limit int
}

func NewState() *State {
        builder := masala.NewSqlBuilder()
        return &State{builder}
}

func (message *State) Id() string {
        return "state"
}

func (message State) Data(state *masala.State) {
        message.builder.Table("myTable").Select("myColumn").Group("myColumnForGroup").Rows(masala.State)
}