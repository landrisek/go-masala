package examples

import ("masala")

type MyController struct {
        builder *masala.SqlBuilder
        limit int
        translatorRepository MyTranslator
}

type MyTranslator struct {

}

func NewMyController() MyController {
        builder := masala.NewSqlBuilder()
        controller := MyController{builder}
        return controller
}

func (controller MyController) Id() string {
        return "MyId"
}

func (controller MyController) Data(state masala.State) masala.State {
        return controller.builder.Table("myTable").Select("myColumn").Group("myColumnForGroup").State(masala.State)
}

func (repository MyTranslator) Translate(term string) string {
        return term
}