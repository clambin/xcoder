package convertor

import (
	"github.com/clambin/go-common/set"
	"sync"
)

type stats struct {
	processing set.Set[string]
	lock       sync.RWMutex
}

func (s *stats) push(path string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.processing == nil {
		s.processing = set.New[string]()
	}
	s.processing.Add(path)
}

func (s *stats) pop(path string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.processing.Remove(path)
}

func (s *stats) getStats() []string {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.processing.ListOrdered()
}
