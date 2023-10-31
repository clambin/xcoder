package inspector

import (
	"context"
	"github.com/clambin/vidconv/internal/feeder"
	"github.com/clambin/vidconv/internal/inspector/mocks"
	"github.com/clambin/vidconv/internal/testutil"
	"github.com/clambin/vidconv/internal/video"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestInspector_Run(t *testing.T) {
	vp := mocks.NewVideoProcessor(t)
	stats := testutil.MakeProbe("h264", 5000, 1080, time.Hour)
	vp.EXPECT().Probe(mock.Anything, mock.AnythingOfType("string")).Return(stats, nil)

	ch := make(chan feeder.Entry)
	i := New(ch, "hevc", slog.Default())
	i.VideoProcessor = vp

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error)
	go func() {
		errCh <- i.Run(ctx)
	}()

	ch <- feeder.Entry{
		Path: "/foo/bar.mkv",
		DirEntry: testutil.FakeDirEntry{
			FName:    "bar.mkv",
			FIsDir:   false,
			FModTime: time.Date(2023, time.November, 7, 0, 0, 0, 0, time.UTC),
		},
	}

	assert.Equal(t, ConversionRequest{
		Input: video.Video{
			Path:    "/foo/bar.mkv",
			ModTime: time.Date(2023, time.November, 7, 0, 0, 0, 0, time.UTC),
			Info:    video.VideoInfo{Name: "bar", Extension: "mkv"},
			Stats:   stats,
		},
		TargetFile:    "/foo/bar.hevc.mkv",
		TargetCodec:   "hevc",
		TargetBitrate: 4000 * 1024,
	}, <-i.Output)
	cancel()
	assert.NoError(t, <-errCh)
}

func TestInspector_makeRequest(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.RemoveAll(tmpDir))
	}()

	var tests = []struct {
		name        string
		input       video.Video
		makeOutput  string
		wantRequest ConversionRequest
		wantReason  string
		wantConvert bool
	}{
		{
			name: "convert",
			input: video.Video{
				Path:    filepath.Join(tmpDir, "foo.mkv"),
				ModTime: time.Date(2023, time.November, 11, 0, 0, 0, 0, time.UTC),
				Info:    video.VideoInfo{Name: "foo", Extension: "mkv"},
				Stats:   testutil.MakeProbe("h264", 9000, 1080, time.Hour),
			},
			wantRequest: ConversionRequest{
				TargetFile:    filepath.Join(tmpDir, "foo.hevc.mkv"),
				TargetCodec:   "hevc",
				TargetBitrate: 4000 * 1024,
			},
			wantConvert: true,
		},
		{
			name: "file exists",
			input: video.Video{
				Path:    filepath.Join(tmpDir, "foo.mkv"),
				ModTime: time.Date(2023, time.November, 11, 0, 0, 0, 0, time.UTC),
				Info:    video.VideoInfo{Name: "foo", Extension: "mkv"},
				Stats:   testutil.MakeProbe("h264", 9000, 1080, time.Hour),
			},
			makeOutput:  filepath.Join(tmpDir, "foo.hevc.mkv"),
			wantReason:  "file already converted",
			wantConvert: false,
		},
		{
			name: "bitrate too low",
			input: video.Video{
				Path:    filepath.Join(tmpDir, "foo.mkv"),
				ModTime: time.Time{},
				Info:    video.VideoInfo{Name: "foo", Extension: "mkv"},
				Stats:   testutil.MakeProbe("h264", 1000, 1080, time.Hour),
			},
			wantReason:  "bitrate too low: 1000 kbps",
			wantConvert: false,
		},
		{
			name: "height too low",
			input: video.Video{
				Path:    filepath.Join(tmpDir, "foo.mkv"),
				ModTime: time.Time{},
				Info:    video.VideoInfo{Name: "foo", Extension: "mkv"},
				Stats:   testutil.MakeProbe("h264", 1000, 720, time.Hour),
			},
			wantReason:  "height too low: 720",
			wantConvert: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := New(nil, "hevc", slog.Default())
			if tt.makeOutput != "" {
				require.NoError(t, os.WriteFile(tt.makeOutput, []byte(``), 0644))
				defer func() {
					require.NoError(t, os.Remove(tt.makeOutput))
				}()
			}

			request, reason, convert := i.makeRequest(tt.input)
			require.Equal(t, tt.wantConvert, convert)
			if !convert {
				assert.Equal(t, tt.wantReason, reason)
			} else {
				assert.Equal(t, tt.wantRequest.TargetFile, request.TargetFile)
			}
		})
	}
}
