package transcoder

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"codeberg.org/clambin/go-common/pubsub"
	"github.com/clambin/xcoder/ffmpeg"
	"github.com/clambin/xcoder/internal/transcoder/evl"
	"golang.org/x/sync/semaphore"
)

const (
	maxConcurrentScans    = 4
	maxConcurrentSessions = 2
	scheduleInterval      = 100 * time.Millisecond
	logProgressInterval   = time.Minute
)

// controller exports selected methods from engine, so they're available from Transcoder
type controller interface {
	Active() bool
	SetActive(bool)
	Subscribe() <-chan SessionEvent
	Unsubscribe(<-chan SessionEvent)
	SessionCount() int
	OverwriteTarget() bool
	RemoveSource() bool
}

type Configuration struct {
	BaseDir         string
	Profile         Profile
	OverwriteTarget bool
	RemoveSource    bool
}

// A Transcoder takes files from the WorkItems list and transcodes them.
type Transcoder struct {
	eventLoop *evl.EventLoop
	controller
}

// New creates a new Transcoder instance
func New(workItems *WorkItems, cfg Configuration, logger *slog.Logger) *Transcoder {
	e := engine{
		probeSema: semaphore.NewWeighted(maxConcurrentScans),
		workItems: workItems,
		logger:    logger,
		profile:   cfg.Profile,
		sessionTracker: sessionTracker{
			sessions:              make(map[*Session]struct{}),
			maxConcurrentSessions: maxConcurrentSessions,
		},
		overwriteTarget: cfg.OverwriteTarget,
		removeSource:    cfg.RemoveSource,
	}

	return &Transcoder{
		controller: &e,
		eventLoop:  evl.New(&e),
	}
}

// Run starts the transcoder event loop
func (t *Transcoder) Run(ctx context.Context) error {
	defer t.eventLoop.Stop()
	return t.eventLoop.Run(ctx)
}

// AddMediaFile adds a media file to the transcoder for processing.
// The transcoder needs to be running, or this will block.
func (t *Transcoder) AddMediaFile(path string) {
	t.eventLoop.Send(newMediaEvent(path))
}

var (
	_ evl.Handler = (*engine)(nil)
	_ controller  = (*engine)(nil)
)

type engine struct {
	probeSema     *semaphore.Weighted
	workItems     *WorkItems
	logger        *slog.Logger
	probeFunc     func(path string) (ffmpeg.VideoStats, error) // only used during testing to stub probe
	transcodeFunc func(session *Session) error                 // only used during testing to stub transcode
	sessionTracker
	profile Profile
	pubsub.Publisher[SessionEvent]
	active          atomic.Bool
	overwriteTarget bool
	removeSource    bool
}

// Init implements the evl.Handler interface.
// It sends the first tickEvent to the event loop.
func (e *engine) Init() evl.Cmd {
	return func() evl.Event { return tickEvent{} }
}

// Update implements the evl.Handler interface.
// It processes events from the event loop and creates commands to handle them.
func (e *engine) Update(msg evl.Event) evl.Cmd {
	//e.logger.Debug("processing event", "event", fmt.Sprintf("%T", msg))
	switch msg := msg.(type) {
	case tickEvent:
		e.queueNextItem()
		return evl.Batch(
			e.startQueuedWorkItemCmd(),
			evl.Tick(scheduleInterval, func() evl.Event { return tickEvent{} }),
		)
	case newMediaEvent:
		// add the file to the work list
		e.logger.Debug("newMediaEvent", "path", string(msg))
		workItem := &WorkItem{Source: File{Path: string(msg)}}
		e.workItems.Add(workItem)
		// scan the workItem
		return e.scanCmd(workItem)
	case transcodeCompleteEvent:
		if status, _ := msg.workItem.Status(); status != StatusConverted {
			return nil
		}
		e.logger.Debug("transcodeCompleteEvent")
		// if we need to remove the source, remove it both from the work list and from the filesystem
		if e.removeSource {
			// delete the file
			err := os.Remove(msg.workItem.Source.Path)
			switch err {
			case nil:
				// remove from the workItems
				e.workItems.Remove(msg.workItem)
			default:
				e.logger.Warn("failed to remove source file", "path", msg.workItem.Source.Path, "err", err)
			}
		}
		// add the converted file to the work list
		e.logger.Debug("queueing newMediaEvent", "path", msg.workItem.Target.Path)
		return func() evl.Event { return newMediaEvent(msg.workItem.Target.Path) }
	default:
		return nil
	}
}

// Active returns whether the Transcoder is active i.e., if it automatically queues items
// that are ready to be transcoded.
func (e *engine) Active() bool {
	return e.active.Load()
}

// SetActive sets the active state of the Transcoder.
func (e *engine) SetActive(active bool) {
	e.active.Store(active)
}

// OverwriteTarget returns whether the Transcoder will overwrite the target file if it exists.
func (e *engine) OverwriteTarget() bool {
	return e.overwriteTarget
}

// RemoveSource returns whether the Transcoder will remove the source file after transcoding successfully.
func (e *engine) RemoveSource() bool {
	return e.removeSource
}

// scanCmd returns an evl.Cmd that scans a new WorkItem to determine its media properties
// and uses the profile to check if the media file can be transcoded and determine the target media properties.
func (e *engine) scanCmd(workItem *WorkItem) evl.Cmd {
	return func() evl.Event {
		logger := e.logger.With(slog.String("source", workItem.Source.Path))
		probe := e.probeFunc
		if probe == nil {
			probe = ffmpeg.Probe
		}

		logger.Debug("acquiring probe semaphore")

		// wait for a probe slot
		_ = e.probeSema.Acquire(context.Background(), 1)
		defer e.probeSema.Release(1)

		logger.Debug("scanning media file")
		start := time.Now()

		// determine source media video stats
		workItem.SetStatus(StatusScanning, nil)
		var err error
		if workItem.Source.VideoStats, err = probe(workItem.Source.Path); err != nil {
			workItem.SetStatus(StatusScanFailed, err)
			logger.Warn("failed to probe media file", "err", err)
			return nil
		}

		// determine target media filename
		workItem.Target.Path = buildTargetFilename(workItem.Source, e.profile.TargetCodec, "mkv")

		// determine target media video stats
		workItem.Target.VideoStats, err = e.profile.Analyze(workItem.Source)

		// set the workItem status
		var status Status
		if err == nil {
			status = StatusScanned
		} else if _, ok := errors.AsType[*SourceSkippedError](err); ok {
			status = StatusSkipped
		} else if _, ok := errors.AsType[*SourceRejectedError](err); ok {
			status = StatusRejected
		} else {
			status = StatusScanFailed
		}
		workItem.SetStatus(status, err)

		logger.Debug("scanned media file", "status", status.String(), "err", err, "duration", time.Since(start))
		return nil
	}
}

// queueNextItem queues the next available work item for transcoding
// if the Transcoder is active and there are available transcoder slots.
func (e *engine) queueNextItem() {
	// don't queue items if we are not active (user will queue manually) or if we don't have any transcoder slots left
	if !e.Active() || e.SessionCount() >= e.maxConcurrentSessions {
		return
	}
	if workItem, ok := e.workItems.GetFirst(StatusScanned); ok {
		if strings.Contains(workItem.Source.Path, ".hevc.") {
			panic("should never happen")
		}
		workItem.SetStatus(StatusQueued, nil)
		e.logger.Debug("queued media file", "path", workItem.Source.Path)
	}
}

// startQueuedWorkItemCmd returns an evl.Cmd that runs the transcoding session. This applies to items
// that are automatically queued (queueNextItem) or manually queued by the user (by setting the StatusQueued status).
// If no session slot is available (as per sessionTracker.maxConcurrentSessions), the command returns nil
// and the work item remains queued.
func (e *engine) startQueuedWorkItemCmd() evl.Cmd {
	// get the next queued item
	workItem, ok := e.workItems.GetFirst(StatusQueued)
	if !ok {
		return nil
	}

	// allocate a transcode session
	session, ok := e.allocateSession(workItem)
	if !ok {
		// no sessions available
		// note: this only happens if the user is manually queuing items,
		// as queueNextItem() checks for a free slot before enqueuing the next work item
		e.logger.Debug("skipping queued item due to max concurrent scans", "count", e.SessionCount())
		return nil
	}

	// mark the workItem here; if we wait until we're in transcodeCmd,
	// startQueuedWorkItemCmd() may pick up the same item twice.
	workItem.SetStatus(StatusTranscoding, nil)

	// start the session
	return e.transcodeCmd(session)
}

// transcodeCmd returns an evl.Cmd that runs the transcoding session for the given session,
// managing the session lifecycle, the workItem state  and session reporting accordingly.
func (e *engine) transcodeCmd(session *Session) evl.Cmd {
	return func() evl.Event {
		start := time.Now()
		logger := e.logger.With(slog.String("source", session.WorkItem.Source.Path))
		logger.Info("started transcoding")

		// inform the listeners that a new session is starting.
		// we do this here rather than in the caller, so we don't block the event loop.
		e.Publish(SessionEvent{Session: session, Type: SessionStartedEvent})

		// run the transcoder session. transcodeFunc allows us to stub transcoding during testing.
		f := e.transcodeFunc
		if f == nil {
			f = e.transcode
		}
		err := f(session)

		// mark the workItem status
		switch err {
		case nil:
			session.WorkItem.SetStatus(StatusConverted, nil)
			logger.Info("finished transcoding", "duration", time.Since(start))
		default:
			session.WorkItem.SetStatus(StatusFailed, err)
			logger.Warn("finished transcoding with errors", "err", err, "duration", time.Since(start))
		}

		// inform listeners that the session has stopped
		e.Publish(SessionEvent{Session: session, Type: SessionStoppedEvent})

		// delete the session
		e.freeSession(session)
		e.logger.Debug("removed work item from session map")
		return transcodeCompleteEvent{workItem: session.WorkItem}
	}
}

// transcode runs the transcoding session for the given session.
func (e *engine) transcode(session *Session) error {
	// add progress monitor to the transcoder
	tmpDir, err := os.MkdirTemp("", "xcoder")
	if err != nil {
		return fmt.Errorf("create temp directory: %w", err)
	}

	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			e.logger.Warn("failed to remove temp directory", "err", err)
		}
	}()

	// encoding arguments
	args, err := encoderArguments(session.WorkItem.Target.VideoStats)
	if err != nil {
		return err
	}

	// callback function to mark and log progress
	logger := e.logger.With(slog.String("source", session.WorkItem.Source.Path))
	lastLogTimestamp := time.Now()
	cb := func(p ffmpeg.Progress) {
		session.progress.Store(&p)
		if time.Since(lastLogTimestamp) > logProgressInterval {
			speed, eta := processSessionProgress(session, p)
			etaString := "N/A"
			if speed > 0 {
				etaString = eta.String()
			}
			logger.Info("transcoding progress", "progress", speed, "eta", etaString)
			lastLogTimestamp = time.Now()
		}
	}

	t := ffmpeg.
		Decode(session.WorkItem.Source.Path, DecoderArguments(session.WorkItem.Source.VideoStats)...).
		Encode(args...).
		Muxer("matroska"). // mkv only
		NoStats().
		LogLevel("error").
		Progress(cb, filepath.Join(tmpDir, "transcoder.sock")).
		Output(session.WorkItem.Target.Path)
	if e.overwriteTarget {
		t = t.OverWriteTarget()
	}

	err = t.Run(context.Background(), e.logger.With(slog.String("source", session.WorkItem.Source.Path)))
	return err
}

func processSessionProgress(session *Session, progress ffmpeg.Progress) (float64, time.Duration) {
	var eta time.Duration
	if progress.Speed > 0 {
		eta = time.Duration(float64(session.WorkItem.Source.VideoStats.Duration-progress.Converted) / progress.Speed)
	}
	return progress.Speed, eta
}

// sessionTracker is a helper for engine that tracks active sessions.
type sessionTracker struct {
	sessions              map[*Session]struct{}
	maxConcurrentSessions int
	mu                    sync.Mutex
}

// allocateSession returns a new session for the workItem.
// If the maximum number of active sessions is reached, it returns nil and false.
func (t *sessionTracker) allocateSession(workItem *WorkItem) (*Session, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// don't exceed maximum number of concurrent transcode sessions
	if len(t.sessions) >= t.maxConcurrentSessions {
		return nil, false
	}

	// add a new session and inform listeners
	session := &Session{WorkItem: workItem}
	t.sessions[session] = struct{}{}
	return session, true
}

// freeSession removes the session from the tracker.
func (t *sessionTracker) freeSession(session *Session) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.sessions, session)
}

// SessionCount returns the number of active sessions.
func (t *sessionTracker) SessionCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.sessions)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// events

type tickEvent struct{}

type newMediaEvent string

type transcodeCompleteEvent struct {
	workItem *WorkItem
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type Session struct {
	WorkItem *WorkItem
	progress atomic.Pointer[ffmpeg.Progress]
}

func (s *Session) Progress() ffmpeg.Progress {
	if p := s.progress.Load(); p != nil {
		return *p
	}
	return ffmpeg.Progress{}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type SessionEvent struct {
	Session *Session
	Type    EventType
}

const (
	SessionStartedEvent EventType = iota
	SessionStoppedEvent
)

type EventType int

func (t EventType) String() string {
	switch t {
	case SessionStartedEvent:
		return "SessionStartedEvent"
	case SessionStoppedEvent:
		return "SessionStoppedEvent"
	default:
		return "UnknownEventType"
	}
}
