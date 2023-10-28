package synclist

import (
	"cmp"
	"slices"
	"sync"
)

type UniqueList[T cmp.Ordered] struct {
	content map[T]struct{}
	lock    sync.RWMutex
}

func (s *UniqueList[T]) Add(t T) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.content == nil {
		s.content = make(map[T]struct{})
	}
	if _, ok := s.content[t]; ok {
		return false
	}
	s.content[t] = struct{}{}
	return true
}

func (s *UniqueList[T]) Remove(t T) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.content == nil {
		return false
	}
	if _, ok := s.content[t]; !ok {
		return false
	}
	delete(s.content, t)
	return true
}

func (s *UniqueList[T]) List() []T {
	s.lock.RLock()
	defer s.lock.RUnlock()
	content := make([]T, 0, len(s.content))
	for key := range s.content {
		content = append(content, key)
	}
	return content
}

func (s *UniqueList[T]) ListOrdered() []T {
	content := s.List()
	slices.Sort(content)
	return content
}
