package processor

import (
	"bytes"
	"context"
	"github.com/clambin/vidconv/internal/feeder"
	"github.com/clambin/vidconv/internal/processor/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProcessor_Run(t *testing.T) {
	ch := make(chan feeder.Video)
	p := New(ch, slog.Default(), "hevc")

	errCh := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		errCh <- p.Run(ctx)
	}()
	cancel()
	assert.NoError(t, <-errCh)

	ctx = context.Background()

	go func() {
		errCh <- p.Run(ctx)
	}()
	close(ch)
	assert.NoError(t, <-errCh)
}

func TestProcessor_process(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.RemoveAll(tmpDir))
	}()

	p := New(nil, slog.Default(), "hevc")
	c := mocks.NewVideoConvertor(t)
	p.VideoConvertor = c

	const targetCodec = "hevc"
	ctx := context.Background()
	source := filepath.Join(tmpDir, "foo.mp4")
	target := filepath.Join(tmpDir, "foo."+targetCodec+".mp4")

	video := feeder.Video{
		Path:    source,
		ModTime: time.Now(),
		Info:    feeder.VideoInfo{Name: "foo", Extension: "mp4"},
		Codec:   "x264",
	}

	// case 1: source exists. destination does not. convert.
	touch(t, source)
	c.EXPECT().Convert(mock.Anything, source, target, targetCodec).Return(nil).Once()
	assert.NoError(t, err, p.process(ctx, video))

	// case 2: both source and destination exists. source is older. don't convert.
	touch(t, source)
	touch(t, target)
	assert.NoError(t, err, p.process(ctx, video))

	// case 3: both source and destination exist. source is more recent. convert.
	touch(t, source)
	touch(t, target)
	video.ModTime = time.Now()

	c.EXPECT().Convert(mock.Anything, source, target, targetCodec).Return(nil).Once()
	assert.NoError(t, err, p.process(ctx, video))
}

func touch(t *testing.T, path string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, nil, 0644))
}

func TestProcessor_Health(t *testing.T) {
	p := New(nil, slog.Default(), "hevc")
	ctx := context.Background()

	tt := []struct {
		name       string
		preActions func(p *Processor)
		want       Health
	}{
		{
			name:       "empty",
			preActions: func(p *Processor) {},
			want:       Health{Processing: []string{}},
		},
		{
			name: "adding",
			preActions: func(p *Processor) {
				p.received.Add(5)
				p.accepted.Add(2)
				p.processing.Add("foo")
				p.processing.Add("bar")
			},
			want: Health{Received: 5, Accepted: 2, Processing: []string{"bar", "foo"}},
		},
		{
			name: "removing",
			preActions: func(p *Processor) {
				p.processing.Remove("foo")
				p.processing.Remove("bar")
			},
			want: Health{Received: 5, Accepted: 2, Processing: []string{}},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tc.preActions(p)
			assert.Equal(t, tc.want, p.Health(ctx))
		})
	}
}

func TestCompressionFactor_LogValue(t *testing.T) {
	var output bytes.Buffer
	l := slog.New(slog.NewTextHandler(&output, &slog.HandlerOptions{ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey {
			return slog.Attr{}
		}
		return a
	}}))

	l.Info("compression", "factor", compressionFactor(0.12345))
	assert.Equal(t, "level=INFO msg=compression factor=0.12\n", output.String())
}
