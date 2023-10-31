package processor

import (
	"bytes"
	"context"
	"github.com/clambin/vidconv/internal/inspector"
	"github.com/clambin/vidconv/internal/processor/mocks"
	"github.com/clambin/vidconv/internal/testutil"
	"github.com/clambin/vidconv/pkg/ffmpeg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestProcessor_Run(t *testing.T) {
	ch := make(chan inspector.Video)
	p := New(ch, "hevc", slog.Default())

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

	p := New(nil, "hevc", slog.Default())
	c := mocks.NewVideoConvertor(t)
	p.VideoConvertor = c

	const targetCodec = "hevc"
	ctx := context.Background()
	source := filepath.Join(tmpDir, "foo.mp4")
	target := filepath.Join(tmpDir, "foo."+targetCodec+".mp4")

	video := inspector.Video{
		Path:    source,
		ModTime: time.Now(),
		Info:    inspector.VideoInfo{Name: "foo", Extension: "mp4"},
		Stats: ffmpeg.Probe{
			Streams: []ffmpeg.Stream{{CodecType: "video", CodecName: "x264"}},
			Format:  ffmpeg.Format{BitRate: strconv.Itoa(4 * 1024 * 1024)},
		},
	}

	// case 1: source exists. destination does not. convert.
	touch(t, source)
	c.EXPECT().ConvertWithProgress(mock.Anything, source, target, targetCodec, mock.AnythingOfType("int"), mock.Anything).Return(nil).Once()
	assert.NoError(t, err, p.process(ctx, video))

	// case 2: both source and destination exists. source is older. don't convert.
	touch(t, source)
	touch(t, target)
	assert.NoError(t, err, p.process(ctx, video))

	// case 3: both source and destination exist. source is more recent. convert.
	touch(t, source)
	touch(t, target)
	video.ModTime = time.Now()

	c.EXPECT().ConvertWithProgress(mock.Anything, source, target, targetCodec, mock.AnythingOfType("int"), mock.Anything).Return(nil).Once()
	assert.NoError(t, err, p.process(ctx, video))
}

func touch(t *testing.T, path string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, nil, 0644))
}

func TestProcessor_Callback(t *testing.T) {
	var buf bytes.Buffer
	l := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey || a.Key == "expected" {
			return slog.Attr{}
		}
		return a
	}}))
	p := New(nil, "hevc", l)
	vc := mocks.NewVideoConvertor(t)
	p.VideoConvertor = vc

	video := inspector.Video{Path: "foo.mp4", Stats: ffmpeg.Probe{Format: ffmpeg.Format{Duration: "3600.00", BitRate: strconv.Itoa(1024 * 1024)}}}
	vc.EXPECT().ConvertWithProgress(mock.Anything, mock.Anything, mock.Anything, "hevc", 921600, mock.Anything).RunAndReturn(
		func(ctx context.Context, s string, s2 string, s3 string, i1 int, f func(ffmpeg.Progress)) error {
			f(ffmpeg.Progress{Converted: 15 * time.Minute, Speed: 1.0})
			f(ffmpeg.Progress{Converted: 30 * time.Minute, Speed: 1.0})
			f(ffmpeg.Progress{Converted: 45 * time.Minute, Speed: 1.0})
			f(ffmpeg.Progress{Converted: 60 * time.Minute, Speed: 1.0})
			return nil
		})

	require.NoError(t, p.process(context.Background(), video))
	assert.Contains(t, buf.String(), `
level=INFO msg=converting path=foo.mp4 progress(%)=25 speed=1
level=INFO msg=converting path=foo.mp4 progress(%)=50 speed=1
level=INFO msg=converting path=foo.mp4 progress(%)=75 speed=1
level=INFO msg=converting path=foo.mp4 progress(%)=100 speed=1
`)
}

func TestProcessor_shouldConvert(t *testing.T) {
	tests := []struct {
		name        string
		file        inspector.Video
		wantBitrate int
		wantReason  string
		wantOK      bool
		wantErr     assert.ErrorAssertionFunc
	}{
		{
			name: "1080p - high",
			file: inspector.Video{
				Path:  "foo.mkv",
				Info:  inspector.VideoInfo{Name: "foo", Extension: "mkv"},
				Stats: testutil.MakeProbe("h264", 9000, 1080, time.Hour),
			},
			wantBitrate: 4800 * 1024,
			wantReason:  "",
			wantOK:      true,
			wantErr:     assert.NoError,
		},
		{
			name: "1080p - same",
			file: inspector.Video{
				Path:  "foo.mkv",
				Info:  inspector.VideoInfo{Name: "foo", Extension: "mkv"},
				Stats: testutil.MakeProbe("h264", 4800, 1080, time.Hour),
			},
			wantBitrate: 4800 * 1024,
			wantReason:  "",
			wantOK:      true,
			wantErr:     assert.NoError,
		},
		{
			name: "1080p - too low",
			file: inspector.Video{
				Path:  "foo.mkv",
				Info:  inspector.VideoInfo{Name: "foo", Extension: "mkv"},
				Stats: testutil.MakeProbe("h264", 10, 1080, time.Hour),
			},
			wantReason: "input video's bitrate is too low (bitrate: 10, minimum: 4800)",
			wantOK:     false,
			wantErr:    assert.NoError,
		},
	}

	p := New(nil, "hevc", slog.Default())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bitrate, reason, ok, err := p.shouldConvert("", tt.file)
			tt.wantErr(t, err)
			assert.Equal(t, tt.wantReason, reason)
			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.wantBitrate, bitrate)
		})
	}
}
