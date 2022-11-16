package socket

import (
	"sync"
)

type Void struct{}

var NULL Void

type Set struct {
	cache map[*Namespace]Void
	// mu
	mu sync.RWMutex
}

func NewSet(keys ...*Namespace) *Set {
	s := &Set{cache: map[*Namespace]Void{}}
	s.Add(keys...)
	return s
}

func (s *Set) Add(keys ...*Namespace) bool {
	if len(keys) == 0 {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, key := range keys {
		s.cache[key] = NULL
	}
	return true
}

func (s *Set) Delete(keys ...*Namespace) bool {
	if len(keys) == 0 {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, key := range keys {
		delete(s.cache, key)
	}
	return true
}

func (s *Set) Clear() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cache = map[*Namespace]Void{}
	return true
}

func (s *Set) Has(key *Namespace) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.cache[key]
	return exists
}

func (s *Set) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.cache)
}

func (s *Set) All() map[*Namespace]Void {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_tmp := map[*Namespace]Void{}

	for k := range s.cache {
		_tmp[k] = NULL
	}

	return _tmp
}

func (s *Set) Keys() (list []*Namespace) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for k := range s.cache {
		list = append(list, k)
	}

	return list
}
