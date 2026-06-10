package transcoder

import (
	"slices"
	"strings"
	"sync"

	"github.com/clambin/xcoder/ffmpeg"
)

type Status int

const (
	StatusFound Status = iota
	StatusScanning
	StatusScanFailed
	StatusScanned
	StatusQueued
	StatusRejected
	StatusSkipped
	StatusTranscoding
	StatusConverted
	StatusFailed
)

var statusStrings = map[Status]string{
	StatusFound:       "found",
	StatusScanning:    "scanning",
	StatusScanFailed:  "scan failed",
	StatusScanned:     "scanned",
	StatusRejected:    "rejected",
	StatusSkipped:     "skipped",
	StatusQueued:      "queued",
	StatusTranscoding: "transcoding",
	StatusFailed:      "failed",
	StatusConverted:   "converted",
}

func (s Status) String() string {
	if str, ok := statusStrings[s]; ok {
		return str
	}
	return "unknown"
}

type File struct {
	Path       string
	VideoStats ffmpeg.VideoStats
}

type WorkItem struct {
	err    error
	Source File
	Target File
	status Status
	mu     sync.Mutex
}

func (w *WorkItem) Status() (status Status, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.status, w.err
}

func (w *WorkItem) SetStatus(status Status, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.status = status
	w.err = err
}

type WorkItems struct {
	items []*WorkItem
	mu    sync.Mutex
}

func (q *WorkItems) Add(items ...*WorkItem) {
	q.mu.Lock()
	defer q.mu.Unlock()
	// if item already exists, delete previous ones
	q.items = slices.DeleteFunc(q.items, func(a *WorkItem) bool {
		return slices.ContainsFunc(items, func(b *WorkItem) bool {
			return a.Source.Path == b.Source.Path
		})
	})
	// Add the new item and sort
	q.items = append(q.items, items...)
	slices.SortFunc(q.items, func(a, b *WorkItem) int {
		return strings.Compare(a.Source.Path, b.Source.Path)
	})
}

func (q *WorkItems) Remove(items ...*WorkItem) {
	q.mu.Lock()
	defer q.mu.Unlock()
	filenames := make(map[string]struct{}, len(items))
	for _, item := range items {
		filenames[item.Source.Path] = struct{}{}
	}
	q.items = slices.DeleteFunc(q.items, func(item *WorkItem) bool {
		_, ok := filenames[item.Source.Path]
		return ok
	})
}

func (q *WorkItems) Items() []*WorkItem {
	q.mu.Lock()
	defer q.mu.Unlock()
	return slices.Clone(q.items)
}

func (q *WorkItems) ItemsWithStatus(status Status) []*WorkItem {
	workItems := q.Items()
	return slices.DeleteFunc(workItems, func(item *WorkItem) bool {
		itemStatus, _ := item.Status()
		return itemStatus != status
	})
}

func (q *WorkItems) GetFirst(status Status) (*WorkItem, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	idx := slices.IndexFunc(q.items, func(item *WorkItem) bool {
		itemStatus, _ := item.Status()
		return itemStatus == status
	})
	if idx != -1 {
		return q.items[idx], true
	}
	return nil, false
}
