package inspector

import (
	"context"
	"errors"
	"github.com/clambin/videoConvertor/internal/server/requests"
	"github.com/clambin/videoConvertor/internal/server/scanner/feeder"
	"github.com/clambin/videoConvertor/internal/testutil"
	"github.com/clambin/videoConvertor/pkg/ffmpeg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestInspector_process(t *testing.T) {
	type probe struct {
		stats ffmpeg.VideoStats
		err   error
	}
	testCases := []struct {
		name    string
		entry   feeder.Entry
		probe   probe
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "success",
			entry: feeder.Entry{
				Path:     "foo.mkv",
				DirEntry: testutil.FakeDirEntry{},
			},
			probe:   probe{stats: testutil.MakeProbe("h264", 6*1024, 1080, time.Hour)},
			wantErr: assert.NoError,
		},
		{
			name: "probe failed",
			entry: feeder.Entry{
				Path:     "foo.mkv",
				DirEntry: testutil.FakeDirEntry{},
			},
			probe:   probe{err: errors.New("probe failed")},
			wantErr: assert.Error,
		},
		{
			name: "already converted",
			entry: feeder.Entry{
				Path:     "foo.mkv",
				DirEntry: testutil.FakeDirEntry{},
			},
			//probe:  probe{stats: testutil.MakeProbe("hevc", 5000, 1080, time.Hour)},
			wantErr: assert.Error,
		},
		{
			name: "too small",
			entry: feeder.Entry{
				Path:     "foo.mkv",
				DirEntry: testutil.FakeDirEntry{},
			},
			probe:   probe{stats: testutil.MakeProbe("h264", 5000, 100, time.Hour)},
			wantErr: assert.Error,
		},
		{
			name: "bitrate too low",
			entry: feeder.Entry{
				Path:     "foo.mkv",
				DirEntry: testutil.FakeDirEntry{},
			},
			probe:   probe{stats: testutil.MakeProbe("h264", 1000, 1080, time.Hour)},
			wantErr: assert.Error,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "")
			require.NoError(t, err)
			defer func() { assert.NoError(t, os.RemoveAll(tmpDir)) }()

			r := requests.Requests{}
			i := New(make(chan feeder.Entry), "hevc-max", &r, slog.Default())
			i.VideoProcessor = fakeProcessor{stats: tt.probe.stats, err: tt.probe.err}

			tt.entry.Path = filepath.Join(tmpDir, tt.entry.Path)
			require.NoError(t, os.WriteFile(tt.entry.Path, []byte{}, 0644))

			req, err := i.process(context.Background(), tt.entry)
			tt.wantErr(t, err)

			if err == nil {
				assert.Equal(t, tt.entry.Path, req.Source)
				assert.Equal(t, "hevc", req.VideoCodec)
			}
		})
	}
}

var _ VideoProcessor = &fakeProcessor{}

type fakeProcessor struct {
	stats ffmpeg.VideoStats
	err   error
}

func (f fakeProcessor) Probe(_ context.Context, _ string) (ffmpeg.VideoStats, error) {
	return f.stats, f.err
}
