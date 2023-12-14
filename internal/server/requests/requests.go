package requests

import "sync"

type Requests struct {
	list []Request
	lock sync.RWMutex
}

func (w *Requests) GetNext() (Request, bool) {
	w.lock.Lock()
	defer w.lock.Unlock()
	var request Request
	var ok bool
	if len(w.list) > 0 {
		ok = true
		request = w.list[0]
		w.list = w.list[1:]
	}
	return request, ok
}

func (w *Requests) Add(request Request) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.list = append(w.list, request)
}

func (w *Requests) Len() int {
	w.lock.RLock()
	defer w.lock.RUnlock()
	return len(w.list)
}

func (w *Requests) List() []string {
	w.lock.RLock()
	defer w.lock.RUnlock()
	files := make([]string, len(w.list))
	for index := range w.list {
		files[index] = w.list[index].Source
	}
	return files
}
