package state

import (
	"github.com/crispyarty/LinkInterceptor/internal/system"
)

type BrowserStore struct {
	Items *ObserveList[*system.Browser]
}

func NewBrowserStore() *BrowserStore {
	return &BrowserStore{
		Items: NewObserveList[*system.Browser](),
	}
}
