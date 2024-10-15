package worklist

import (
	"github.com/clambin/videoConvertor/internal/ffmpeg"
	"iter"
	"slices"
	"strconv"
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
		item.SetStatus(Converting, nil)
		wl.queued = wl.queued[1:]
	}
	return item
}

func (wl *WorkList) checkout(current, next WorkStatus) *WorkItem {
	wl.lock.RLock()
	defer wl.lock.RUnlock()
	for _, item := range wl.list {
		if status, _ := item.Status(); status == current {
			item.SetStatus(next, nil)
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

// List returns all WorkItem records in the list. This clones the contained slice. For performance reasons,
// this should only be used for testing. Use All(), which returns an iterator, instead.
func (wl *WorkList) List() []*WorkItem {
	wl.lock.RLock()
	defer wl.lock.RUnlock()
	return slices.Clone(wl.list)
}

func (wl *WorkList) Size() int {
	wl.lock.RLock()
	defer wl.lock.RUnlock()
	return len(wl.list)
}

func (wl *WorkList) All() iter.Seq[*WorkItem] {
	return func(yield func(*WorkItem) bool) {
		wl.lock.RLock()
		defer wl.lock.RUnlock()
		for _, item := range wl.list {
			if !yield(item) {
				return
			}
		}
	}
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
	err         error
	Source      string
	sourceStats ffmpeg.VideoStats
	Progress
	status             WorkStatus
	lock               sync.RWMutex
	constantRateFactor int
}

func (w *WorkItem) Status() (WorkStatus, error) {
	w.lock.RLock()
	defer w.lock.RUnlock()
	return w.status, w.err
}

func (w *WorkItem) SetStatus(status WorkStatus, err error) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.status = status
	w.err = err
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
	w.Progress.Duration = stats.Duration
}

func (w *WorkItem) SetConstantRateFactor(constantRateFactor int) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.constantRateFactor = constantRateFactor
}

func (w *WorkItem) RemainingFormatted() string {
	w.lock.RLock()
	defer w.lock.RUnlock()
	if w.status != Converting {
		return ""
	}
	var output string
	if d := w.Progress.Remaining(); d >= 0 { // not sure why I added this check?
		output = formatDuration(d)
	}
	return output
}

func formatDuration(d time.Duration) string {
	var output string
	var days int
	if days = int(d.Hours()) / 24; days > 0 {
		output = strconv.Itoa(days) + "d"
		d -= time.Duration(days) * 24 * time.Hour
	}
	if hours := int(d.Hours()); hours > 0 {
		output += strconv.Itoa(hours) + "h"
		d -= time.Duration(hours) * time.Hour
	}
	if minutes := int(d.Minutes()); minutes > 0 {
		output += strconv.Itoa(minutes) + "m"
		d -= time.Duration(minutes) * time.Minute
	}
	if seconds := int(d.Seconds()); seconds > 0 {
		output += strconv.Itoa(seconds) + "s"
	}
	return output
}

func (w *WorkItem) CompletedFormatted() string {
	w.lock.RLock()
	defer w.lock.RUnlock()
	if w.status != Converting {
		return ""
	}
	if p := w.Progress.Completed(); p > 0 {
		return strconv.FormatFloat(100*p, 'f', 1, 64) + "%"
	}
	return ""
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
	p.progress.Converted = progress.Converted
	// if speed is zero, we won't be able to calculate the remaining time. in this case, we don't update and the
	// remaining time will be calculated using the last  reported speed.
	if progress.Speed != 0 {
		p.progress.Speed = progress.Speed
	}
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
	return time.Duration(float64(p.Duration-p.progress.Converted) / p.progress.Speed)
}
