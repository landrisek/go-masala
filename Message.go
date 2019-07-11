package masala

type IMessage interface {
	Data(state State) State
	Id() string
}