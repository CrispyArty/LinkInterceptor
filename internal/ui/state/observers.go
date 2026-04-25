package state

import (
	"sync"
)

type ObserveList[D any] struct {
	data      []D
	mu        sync.Mutex
	listeners map[int]func([]D)
	nextID    int
}

func NewObserveList[D any]() *ObserveList[D] {
	return &ObserveList[D]{
		listeners: make(map[int]func([]D)),
	}
}

func (s *ObserveList[D]) All() []D {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.data == nil {
		return nil
	}
	return append([]D(nil), s.data...)
}

type Unsubscriber func()

func (s *ObserveList[D]) Subscribe(f func([]D)) Unsubscriber {
	s.mu.Lock()
	id := s.nextID
	s.listeners[id] = f
	s.nextID++

	snapshot := append([]D(nil), s.data...)
	s.mu.Unlock()

	if len(snapshot) > 0 {
		go f(snapshot)
	}

	return func() { s.Unsubscribe(id) }
}

func (s *ObserveList[D]) Unsubscribe(id int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.listeners, id)
}

func (s *ObserveList[D]) Update(newData []D) {
	s.mu.Lock()
	s.data = append([]D(nil), newData...)
	snapshot, callbacks := s.snapLocked()

	s.mu.Unlock()

	s.notify(snapshot, callbacks)
}

func (s *ObserveList[D]) Add(item D) {
	s.mu.Lock()
	s.data = append(s.data, item)

	snapshot, callbacks := s.snapLocked()
	// snapshot := append([]D(nil), s.data...)
	// callbacks := make([]func([]D), len(s.listeners))
	// i := 0
	// for _, f := range s.listeners {
	// 	callbacks[i] = f
	// 	i++
	// }

	s.mu.Unlock()

	s.notify(snapshot, callbacks)
}

func (s *ObserveList[D]) snapLocked() ([]D, []func([]D)) {
	snapshot := append([]D(nil), s.data...)
	callbacks := make([]func([]D), len(s.listeners))
	i := 0
	for _, f := range s.listeners {
		callbacks[i] = f
		i++
	}
	return snapshot, callbacks
}

func (s *ObserveList[D]) notify(snapshot []D, callbacks []func([]D)) {
	for _, f := range callbacks {
		if f != nil {
			go f(snapshot)
		}
	}
}
