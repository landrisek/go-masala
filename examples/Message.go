package examples

import ("masala")

type MyMessage struct {
        builder *masala.SqlBuilder
        limit int
        translatorRepository MyTranslator
}

func NewMessage() *MyMessage {
        builder := masala.NewSqlBuilder()
        return &MyMessage{builder}
}

func (message *MyMessage) Id() string {
        return "myState"
}

func (message *MyMessage) Data(state *masala.State) {
        state.Menu = "myMenu"
}