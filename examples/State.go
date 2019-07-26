package examples

import ("masala")

type IExport interface {
	masala.IState
	GetUser() int
}

type State struct {
	Autocomplete struct {
		Data map[string]string
		Position int
	}
	Clicked map[string]bool
	Crops map[string]string
	Group string
	Menu string
	Order map[string]string
	Paginator masala.Paginator
	Rows []map[string]string
	User int
	Where map[string]interface{}
	Wysiwyg map[string]string
}

func (state *State) GetCriteria() map[string]interface{} {
	return state.Where
}

func (state *State) GetGroup() string {
	return state.Group
}

func (state *State) GetOrder() map[string]string {
	return state.Order
}

func (state *State) GetPaginator() masala.Paginator {
	return state.Paginator
}

func (state *State) GetRows() []map[string]string {
	return state.Rows
}

func (state *State) GetUser() int {
	return state.User
}

func (state *State) SetPaginator(paginator masala.Paginator) masala.IState {
	state.Paginator = paginator
	return state
}

func (state *State) SetRows(rows []map[string]string) masala.IState {
	state.Rows = rows
	return state
}