package masala

import ("encoding/json"
	"html/template"
	"io/ioutil"
	"net/http")

type MyController struct {
	builder SqlBuilder
	limit int
	translatorRepository MyTranslator
}

type MyTranslator struct {

}

func (repository MyTranslator) Translate(term string) string {
	return term
}

func (controller MyController) Inject() MyController {
	controller.builder = SqlBuilder{}.Inject()
	controller.limit = 20
	controller.translatorRepository = MyTranslator{}
	return controller
}

func (controller MyController) Props(response http.ResponseWriter, request *http.Request) {
	body, _ := ioutil.ReadAll(request.Body)
	var payload struct{ Value string }
	json.Unmarshal(body, &payload)
	props := controller.builder.Props(request, controller.translatorRepository)
	props["myIdLabel"] = controller.translatorRepository.Translate("myLabel")
	data, _ := json.Marshal(props)
	file := template.Must(template.New("props.html").ParseFiles("../templates/props.html"))
	file.Execute(response, map[string]interface{}{"props":template.HTML(string(data))})
}

func (controller MyController) State(response http.ResponseWriter, request *http.Request) {
	myLimit := 20
	state, _ := json.Marshal(controller.builder.Table("myTable").Select("myColumns").State(
		"MyColumnForGroup", myLimit, request))
	file := template.Must(template.New("state.html").ParseFiles("../templates/state.html"))
	file.Execute(response, map[string]interface{}{"state":template.HTML(string(state))})
}