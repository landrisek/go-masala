package masala

import ("encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync")

type Server struct {
	addClient    chan *Client
	channels     map[string]*Channel
	closeChannel chan string
	headers map[string]string
	logger *log.Logger
	mutex       sync.RWMutex
	removeClient chan *Client
	retry int
	shutdown     chan bool
}

func (server Server) Inject() *Server {
	pointer := &Server{make(chan *Client),
		make(map[string]*Channel),
		make(chan string),
		map[string]string{"Cache-Control":"no-cache","Content-Type":"text/event-stream","Connection":"keep-alive"},
		log.New(os.Stdout, "masala server: ", log.LstdFlags),
		sync.RWMutex{},
		make(chan *Client),
		1,
		make(chan bool)}
	go pointer.start()
	return pointer
}

func (server *Server) addChannel(name string) *Channel {
	channel := Channel{}.Inject(name)
	server.mutex.Lock()
	server.channels[channel.name] = channel
	server.mutex.Unlock()
	server.logger.Printf("channel '%s' created.", channel.name)
	return channel
}

func (server *Server) Channels() []string {
	channels := []string{}
	server.mutex.RLock()
	for name := range server.channels {
		channels = append(channels, name)
	}
	server.mutex.RUnlock()
	return channels
}

func (server *Server) CloseChannel(name string) {
	server.closeChannel <- name
}

func (server *Server) ClientCount() int {
	i := 0
	server.mutex.RLock()
	for _, channel := range server.channels {
		i += channel.ClientCount()
	}
	server.mutex.RUnlock()
	return i
}

func (server *Server) close() {
	for _, channel := range server.channels {
		server.removeChannel(channel)
	}
}

func (server *Server) getChannel(name string) (*Channel, bool) {
	server.mutex.RLock()
	channel, ok := server.channels[name]
	server.mutex.RUnlock()
	return channel, ok
}

func (server *Server) GetChannel(name string) (*Channel, bool) {
	return server.getChannel(name)
}

func (server *Server) HasChannel(name string) bool {
	_, ok := server.getChannel(name)
	return ok
}

func (server *Server) Restart() {
	server.logger.Print("restarting server.")
	server.close()
}

func (server *Server) removeChannel(channel *Channel) {
	server.mutex.Lock()
	delete(server.channels, channel.name)
	server.mutex.Unlock()
	channel.Close()
	server.logger.Printf("channel '%s' closed.", channel.name)
}

func (server *Server) SetHeader(header string, value string) {
	server.headers[header] = value
}

func (server *Server) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	flusher, ok := response.(http.Flusher)
	if !ok {
		http.Error(response, "Streaming unsupported.", http.StatusInternalServerError)
		return
	}
	headers := response.Header()
	for header, value := range server.headers {
			headers.Set(header, value)
	}
	if "GET" == request.Method {
		var state State
		json.Unmarshal([]byte(request.URL.Query()["masala"][0]), &state)
		lastEventID := request.Header.Get("Last-Event-ID")
		client := Client{}.Inject(lastEventID, request.URL.Path)
		server.addClient <- client
		closeNotify := response.(http.CloseNotifier).CloseNotify()
		go func() {
			<-closeNotify
			server.removeClient <- client
		}()
		response.WriteHeader(http.StatusOK)
		flusher.Flush()
		for message := range client.send {
			fmt.Fprintf(response, message.String(state))
			flusher.Flush()
		}
	} else if  "OPTIONS" != request.Method {
		response.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (server *Server) SendMessage(channelName string, message IMessage) {
	if len(channelName) == 0 {
		server.logger.Print("broadcasting message to all channels.")
		server.mutex.RLock()
		for _, channel := range server.channels {
			channel.SendMessage(message)
		}
		server.mutex.RUnlock()
	} else if channel, ok := server.getChannel(channelName); ok {
		channel.SendMessage(message)
		server.logger.Printf("message sent to channel '%s'.", channelName)
	} else {
		server.logger.Printf("message not sent because channel '%s' has no clients.", channelName)
	}
}

func (server *Server) start() {
	server.logger.Print("server started.")
	for {
		select {
		case client := <-server.addClient:
			channel, exists := server.getChannel(client.channel)
			if !exists {
				channel = server.addChannel(client.channel)
			}
			channel.addClient(client)
			server.logger.Printf("new client connected to channel '%s'.", channel.name)
		case client := <-server.removeClient:
			if channel, exists := server.getChannel(client.channel); exists {
				channel.removeClient(client)
				server.logger.Printf("client disconnected from channel '%s'.", channel.name)
				if channel.ClientCount() == 0 {
					server.logger.Printf("channel '%s' has no clients.", channel.name)
					server.removeChannel(channel)
				}
			}
		case channel := <-server.closeChannel:
			if ch, exists := server.getChannel(channel); exists {
				server.removeChannel(ch)
			} else {
				server.logger.Printf("requested to close nonexistent channel '%s'.", channel)
			}
		case <-server.shutdown:
			server.close()
			close(server.addClient)
			close(server.removeClient)
			close(server.closeChannel)
			close(server.shutdown)
			server.logger.Print("server stopped.")
			return
		}
	}
}

func (server *Server) Shutdown() {
	server.shutdown <- true
}



