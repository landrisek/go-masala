package examples

import ("masala"
	"time")

func NewController(server *masala.Server) {
	for {
		server.SendMessage("/myUrl", examples.NewState())
		server.SendMessage("/myUrl", examples.NewOtherState())
		time.Sleep(5 * time.Second)
	}
}
