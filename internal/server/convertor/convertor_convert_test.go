package convertor_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/clambin/vidconv/internal/server/convertor"
	"github.com/clambin/vidconv/internal/server/requests"
	"github.com/clambin/vidconv/pkg/ffmpeg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConvertor_convert(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.RemoveAll(tmpDir))
	}()

	l := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	store := requests.Requests{}

	c := convertor.New(&store, true, l)
	c.VideoConvertor = &fakeConvertor{}

	const payload = "0123456789"
	c.SetActive(true)

	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan error)
	go func() { ch <- c.Run(ctx, 1) }()

	source := filepath.Join(tmpDir, "input.mkv")
	require.NoError(t, os.WriteFile(source, []byte(payload), 0644))

	target := filepath.Join(tmpDir, "output.mkv")
	store.Add(requests.Request{
		Request: ffmpeg.Request{
			Source:        source,
			Target:        target,
			VideoCodec:    "hevc",
			BitRate:       0,
			BitsPerSample: 0,
		},
		SourceStats: ffmpeg.VideoStats{},
	})

	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, time.Second, time.Millisecond)

	assert.Eventually(t, func() bool {
		content, err := os.ReadFile(target)
		return err == nil && string(content) == payload
	}, time.Second, time.Millisecond)

	assert.Eventually(t, func() bool {
		_, err := os.ReadFile(source)
		return errors.Is(err, os.ErrNotExist)
	}, time.Second, time.Millisecond)
	cancel()
	assert.NoError(t, <-ch)
}

var _ convertor.VideoConvertor = &fakeConvertor{}

type fakeConvertor struct{}

func (f fakeConvertor) Convert(_ context.Context, request ffmpeg.Request) error {
	body, err := os.ReadFile(request.Source)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}
	return os.WriteFile(request.Target, body, 0644)
}
