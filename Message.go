package masala

type IMessage interface {
	Data(state *State)
	Id() string
}