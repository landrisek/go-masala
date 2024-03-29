package main

import ("examples"
	"github.com/jinzhu/configor"
	"log"
	"masala"
	"net/http"
	"os")

type config struct {
	Domain string
	Server struct {
		Certificate string
		Key string
		Port string
	}
}

func main() {
	config := config{}
	configor.Load(&config, "config.yml")
	host, _ := os.Hostname()
	configor.Load(&config, "config." + host + ".yml")
	server := masala.NewServer()
	server.SetHeader("Access-Control-Allow-Origin", config.Domain)
	server.SetHeader("Access-Control-Allow-Credentials", "true")
	server.SetHeader("Access-Control-Allow-Methods", "GET,HEAD,OPTIONS,POST,PUT")
	server.SetHeader("Access-Control-Allow-Headers", "Access-Control-Allow-Headers, Origin,Accept, X-Requested-With, Content-Type, Access-Control-Request-Method, Access-Control-Request-Headers")
	defer server.Shutdown()
	http.Handle("/", server)
	go examples.NewController()
	err := http.ListenAndServeTLS(config.Server.Port, config.Server.Certificate, config.Server.Key, nil)
	if nil != err {
		log.Panic(err)
	}
}