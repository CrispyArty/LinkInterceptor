package state

type State struct {
	Caller   string
	Url      string
	Browsers *BrowserStore
}

func NewState() *State {
	return &State{
		Browsers: NewBrowserStore(),
	}
}
