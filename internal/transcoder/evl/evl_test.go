package evl_test

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/clambin/xcoder/internal/transcoder/evl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventLoop(t *testing.T) {
	var client eventClient

	e := evl.New(&client)
	defer e.Stop()

	go func() {
		require.NoError(t, e.Run(t.Context()))
	}()

	assert.Eventually(t, func() bool {
		return client.count.Load() >= 10
	}, 10*time.Second, 10*time.Millisecond)

	assert.GreaterOrEqual(t, client.count.Load(), int64(10))

}

func TestBatch(t *testing.T) {
	var client batchEventClient

	e := evl.New(&client)
	defer e.Stop()

	go func() {
		require.NoError(t, e.Run(t.Context()))
	}()

	assert.Eventually(t, func() bool {
		return client.count.Load() >= 10
	}, 10*time.Second, 10*time.Millisecond)

	assert.GreaterOrEqual(t, client.count.Load(), int64(10))
}

var (
	_ evl.Handler = (*eventClient)(nil)
	_ evl.Handler = (*batchEventClient)(nil)
)

type tickMsg struct{}

type eventClient struct {
	count atomic.Int64
}

func (e *eventClient) Init() evl.Cmd {
	return e.Tick()
}

func (e *eventClient) Update(msg evl.Event) evl.Cmd {
	switch msg.(type) {
	case tickMsg:
		e.count.Add(1)
		return evl.Tick(100*time.Millisecond, e.Tick())
	default:
		panic(fmt.Sprintf("unknown message type: %T", msg))
	}
}

func (e *eventClient) Tick() evl.Cmd {
	return func() evl.Event {
		return tickMsg{}
	}
}

type batchEventClient struct {
	count atomic.Int64
}

func (b *batchEventClient) Init() evl.Cmd {
	cmds := make([]evl.Cmd, 10)
	for i := range cmds {
		cmds[i] = func() evl.Event { return tickMsg{} }
	}
	return evl.Batch(cmds...)
}

func (b *batchEventClient) Update(event evl.Event) evl.Cmd {
	switch event.(type) {
	case tickMsg:
		b.count.Add(1)
		return nil
	default:
		panic(fmt.Sprintf("unknown message type: %T", event))
	}
}
