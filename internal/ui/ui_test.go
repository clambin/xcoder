package ui

import "github.com/clambin/xcoder/internal/transcoder"

var _ Transcoder = (*fakeTranscoder)(nil)

type fakeTranscoder struct {
	active bool
	count  int
}

func (f *fakeTranscoder) SessionCount() int {
	return f.count
}

func (f *fakeTranscoder) Active() bool {
	return f.active
}

func (f *fakeTranscoder) SetActive(_ bool) {
	panic("implement me")
}

func (f *fakeTranscoder) Subscribe() <-chan transcoder.SessionEvent {
	panic("implement me")
}

func (f *fakeTranscoder) OverwriteTarget() bool {
	return true
}

func (f *fakeTranscoder) RemoveSource() bool {
	return true
}
