package evl

import (
	"context"
	"time"
)

type Event any
type Cmd func() Event

type batchEvent []Cmd

func Batch(cmd ...Cmd) Cmd {
	return func() Event {
		return batchEvent(cmd)
	}
}

func Tick(delay time.Duration, cmd Cmd) Cmd {
	return func() Event {
		time.Sleep(delay)
		return cmd()
	}
}

type Handler interface {
	Init() Cmd
	Update(Event) Cmd
}

type EventLoop struct {
	events  chan Event
	cmds    chan Cmd
	done    chan struct{}
	handler Handler
}

func New(h Handler) *EventLoop {
	return &EventLoop{
		events:  make(chan Event),
		cmds:    make(chan Cmd),
		done:    make(chan struct{}),
		handler: h,
	}
}

func (e *EventLoop) Run(ctx context.Context) error {
	e.handleCmd(e.handler.Init())
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-e.done:
			return nil
		case ev := <-e.events:
			go e.handleEvent(ev)
		case cmd := <-e.cmds:
			e.handleCmd(cmd)
		}
	}
}

func (e *EventLoop) Stop() {
	close(e.done)
}

func (e *EventLoop) Send(ev Event) {
	e.events <- ev
}

func (e *EventLoop) handleEvent(ev Event) {
	e.cmds <- e.handler.Update(ev)
}

func (e *EventLoop) handleCmd(cmd Cmd) {
	if cmd == nil {
		return
	}
	go func() {
		switch ev := cmd().(type) {
		case batchEvent:
			for _, c := range ev {
				e.handleCmd(c)
			}
		default:
			e.Send(ev)
		}
	}()
}
