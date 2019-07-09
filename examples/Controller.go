package examples

import ("bytes"
        "fmt"
        "masala"
        "strings")

type MyController struct {
        builder masala.SqlBuilder
}

func NewMyController() MyController {
        builder := masala.SqlBuilder{}.Inject()
        controller := MyController{builder}
        return controller
}

func (controller MyController) Id() string {
        return "MyId"
}

func (message MyController) String(state masala.State) string {
        /** ...aplication logic **/
        fmt.Print(state, "\n")
        var buffer bytes.Buffer
        buffer.WriteString(fmt.Sprintf("id: %s\n", "1"))
        buffer.WriteString(fmt.Sprintf("data: %s\n", strings.Replace("{}", "\n", "\ndata: ", -1)))
        buffer.WriteString("\n")
        return buffer.String()
}