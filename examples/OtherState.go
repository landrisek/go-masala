package examples

import ("masala")

type OtherState struct {
        builder *masala.SqlBuilder
        limit int
        translatorRepository MyTranslator
}

func NewOtherState() *OtherState {
        builder := masala.NewSqlBuilder()
        return &OtherState{builder}
}

func (message *OtherState) Id() string {
        return "state"
}

func (message OtherState) Data(state *masala.State) {
        state.Menu = "myMenu"
}