package masala

type IMessage interface {
	Id() string
	String(state State) string
}