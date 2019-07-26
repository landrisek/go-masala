package masala

type IState interface {
	GetCriteria() map[string]interface{}
	GetGroup() string
	GetOrder() map[string]string
	GetPaginator() Paginator
	SetPaginator(paginator Paginator) IState
	GetRows() []map[string]string
	SetRows([]map[string]string) IState
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
	Paginator Paginator
	Rows []map[string]string
	Where map[string]interface{}
	Wysiwyg map[string]string
}

type Paginator struct {
	Current int
	Last int
	Sum int
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

func (state *State) GetPaginator() Paginator {
	return state.Paginator
}

func (state *State) GetRows() []map[string]string {
	return state.Rows
}

func (state *State) SetPaginator(paginator Paginator) IState {
	state.Paginator = paginator
	return state
}

func (state *State) SetRows(rows []map[string]string) IState {
	state.Rows = rows
	return state
}