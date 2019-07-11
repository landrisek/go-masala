package masala

import ("sync")

type Channel struct {
	clients     map[*Client]bool
	lastEventID string
	mutex          sync.RWMutex
	name        string
}

func NewChannel(name string) *Channel {
	return &Channel{
		make(map[*Client]bool),
		"",
		sync.RWMutex{},
		name}
}

func (channel *Channel) addClient(client *Client) {
	channel.mutex.Lock()
	channel.clients[client] = true
	channel.mutex.Unlock()
}

func (channel *Channel) ClientCount() int {
	channel.mutex.RLock()
	count := len(channel.clients)
	channel.mutex.RUnlock()
	return count
}

func (channel *Channel) Close() {
	for client := range channel.clients {
		channel.removeClient(client)
	}
}

func (channel *Channel) SendMessage(message IMessage) {
	channel.lastEventID = message.Id()
	channel.mutex.RLock()
	for channel, open := range channel.clients {
		if open {
			channel.send <- message
		}
	}
	channel.mutex.RUnlock()
}


func (channel *Channel) LastEventID() string {
	return channel.lastEventID
}

func (channel *Channel) removeClient(client *Client) {
	channel.mutex.Lock()
	channel.clients[client] = false
	delete(channel.clients, client)
	channel.mutex.Unlock()
	close(client.send)
}
