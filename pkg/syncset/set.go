package syncset

import (
	"cmp"
	"github.com/clambin/go-common/set"
	"slices"
	"sync"
)

type Set[T cmp.Ordered] struct {
	content set.Set[T]
	lock    sync.RWMutex
}

func New[T cmp.Ordered]() *Set[T] {
	return &Set[T]{content: set.Create[T]()}
}

func (s *Set[T]) Add(t T) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.content.Contains(t) {
		return false
	}
	s.content.Add(t)
	return true
}

func (s *Set[T]) Remove(t T) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	if !s.content.Contains(t) {
		return false
	}
	s.content.Remove(t)
	return true
}

func (s *Set[T]) List() []T {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.content.List()
}

func (s *Set[T]) ListOrdered() []T {
	content := s.List()
	slices.Sort(content)
	return content
}
