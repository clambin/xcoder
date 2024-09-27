package worklist

import (
	"github.com/clambin/videoConvertor/internal/ffmpeg"
	"slices"
	"sync"
	"time"
)

type WorkList struct {
	list   []*WorkItem
	queued []*WorkItem
	lock   sync.RWMutex
	active bool
}

func (wl *WorkList) Add(filename string) *WorkItem {
	wl.lock.Lock()
	defer wl.lock.Unlock()
	item := &WorkItem{Source: filename}
	wl.list = append(wl.list, item)
	return item
}

func (wl *WorkList) NextToConvert() *WorkItem {
	// convert any items the user manually asked to convert?
	if item := wl.dequeue(); item != nil {
		return item
	}
	// is the worklist active?
	if !wl.Active() {
		return nil
	}
	// return the next item ready for conversion
	return wl.checkout(Inspected, Converting)
}

func (wl *WorkList) dequeue() *WorkItem {
	wl.lock.Lock()
	defer wl.lock.Unlock()
	var item *WorkItem
	if len(wl.queued) > 0 {
		item = wl.queued[0]
		item.setStatus(Converting, nil)
		wl.queued = wl.queued[1:]
	}
	return item
}

func (wl *WorkList) checkout(current, next WorkStatus) *WorkItem {
	wl.lock.RLock()
	defer wl.lock.RUnlock()
	for _, item := range wl.list {
		if status, _ := item.Status(); status == current {
			item.setStatus(next, nil)
			return item
		}
	}
	return nil
}

// Queue adds an item ready to be converted. This item will be processed, regardless of whether the queue is active or not.
func (wl *WorkList) Queue(item *WorkItem) {
	wl.lock.Lock()
	defer wl.lock.Unlock()
	wl.queued = append(wl.queued, item)
}

func (wl *WorkList) List() []*WorkItem {
	wl.lock.RLock()
	defer wl.lock.RUnlock()
	return slices.Clone(wl.list)
}

func (wl *WorkList) Active() bool {
	wl.lock.RLock()
	defer wl.lock.RUnlock()
	return wl.active
}

func (wl *WorkList) SetActive(active bool) {
	wl.lock.Lock()
	defer wl.lock.Unlock()
	wl.active = active
}

func (wl *WorkList) ToggleActive() {
	wl.lock.Lock()
	defer wl.lock.Unlock()
	wl.active = !wl.active
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type WorkStatus int

const (
	Waiting WorkStatus = iota
	Inspecting
	Skipped
	Rejected
	Inspected
	Converting
	Converted
	Failed
)

var workStatusToString = map[WorkStatus]string{
	Waiting:    "waiting",
	Inspecting: "inspecting",
	Skipped:    "skipped",
	Rejected:   "rejected",
	Inspected:  "inspected",
	Converting: "converting",
	Converted:  "converted",
	Failed:     "failed",
}

func (ws WorkStatus) String() string {
	if label, ok := workStatusToString[ws]; ok {
		return label
	}
	return "unknown"
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type WorkItem struct {
	Source string
	status WorkStatus
	Progress
	err         error
	sourceStats ffmpeg.VideoStats
	targetStats ffmpeg.VideoStats
	lock        sync.RWMutex
}

func (w *WorkItem) Status() (WorkStatus, error) {
	w.lock.RLock()
	defer w.lock.RUnlock()
	return w.status, w.err
}

func (w *WorkItem) setStatus(status WorkStatus, err error) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.status = status
	w.err = err
}

func (w *WorkItem) Done(status WorkStatus, err error) {
	w.setStatus(status, err)
}

func (w *WorkItem) SourceVideoStats() ffmpeg.VideoStats {
	w.lock.RLock()
	defer w.lock.RUnlock()
	return w.sourceStats
}

func (w *WorkItem) AddSourceStats(stats ffmpeg.VideoStats) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.sourceStats = stats
	w.Progress.Duration = stats.Duration()
}

func (w *WorkItem) TargetVideoStats() ffmpeg.VideoStats {
	w.lock.RLock()
	defer w.lock.RUnlock()
	return w.targetStats
}

func (w *WorkItem) AddTargetStats(stats ffmpeg.VideoStats) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.targetStats = stats
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type Progress struct {
	lock     sync.RWMutex
	Duration time.Duration
	progress ffmpeg.Progress
}

func (p *Progress) Update(progress ffmpeg.Progress) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.progress = progress
}

func (p *Progress) Completed() float64 {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return float64(p.progress.Converted) / float64(p.Duration)
}

func (p *Progress) Remaining() time.Duration {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if p.progress.Speed == 0 {
		return -1
	}
	remaining := p.Duration - p.progress.Converted
	return time.Duration(float64(remaining) / p.progress.Speed)
}
