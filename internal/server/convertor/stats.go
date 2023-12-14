package convertor

import (
	"slices"
	"sync"
)

type stats struct {
	processing map[string]struct{}
	lock       sync.RWMutex
}

func (s *stats) push(path string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.processing == nil {
		s.processing = make(map[string]struct{})
	}
	s.processing[path] = struct{}{}
}

func (s *stats) pop(path string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.processing, path)
}

func (s *stats) getStats() []string {
	s.lock.RLock()
	defer s.lock.RUnlock()
	files := make([]string, 0, len(s.processing))
	for file := range s.processing {
		files = append(files, file)
	}
	slices.Sort(files)
	return files
}
