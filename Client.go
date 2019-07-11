package masala

type Client struct {
	lastEventID,
	channel string
	send chan IMessage
}

func NewClient(lastEventID, channel string) *Client {
	return &Client{
		lastEventID,
		channel,
		make(chan IMessage),
	}
}

func (client *Client) SendMessage(message IMessage) {
	client.lastEventID = message.Id()
	client.send <- message
}

func (client *Client) Channel() string {
	return client.channel
}

func (client *Client) LastEventID() string {
	return client.lastEventID
}
