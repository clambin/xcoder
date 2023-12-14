package scanner

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/clambin/videoConvertor/internal/server/requests"
	"github.com/clambin/videoConvertor/internal/server/scanner/inspector"
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

var validVideoFiles = []string{"foo.mkv"}

func TestApplication_Run(t *testing.T) {
	fs := testutil.MakeTempFS(t, validVideoFiles)
	defer func() { assert.NoError(t, os.RemoveAll(fs)) }()

	var r requests.Requests
	a, _ := New(fs, "hevc-max", &r, slog.Default())
	a.Inspector.VideoProcessor = fakeProcessor{stats: testutil.MakeProbe("h264", 8000, 1080, time.Hour)}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error)
	go func() {
		errCh <- a.Run(ctx, 1)
	}()

	assert.Eventually(t, func() bool {
		return r.Len() > 0
	}, time.Second, time.Millisecond)

	req, ok := r.GetNext()
	assert.True(t, ok)
	assert.Equal(t, "foo.mkv", filepath.Base(req.Source))

	cancel()
	assert.NoError(t, <-errCh)
}

func TestScanner_Health(t *testing.T) {
	var r requests.Requests
	a, _ := New("/tmp", "hevc-max", &r, slog.Default())
	r.Add(requests.Request{
		Request: ffmpeg.Request{Source: "foo.mkv"},
	})

	var body bytes.Buffer
	require.NoError(t, json.NewEncoder(&body).Encode(a.Health(context.Background())))
	assert.Equal(t, `{"Queued":["foo.mkv"]}
`, body.String())
}

var _ inspector.VideoProcessor = fakeProcessor{}

type fakeProcessor struct {
	stats ffmpeg.VideoStats
	err   error
}

func (f fakeProcessor) Probe(_ context.Context, _ string) (ffmpeg.VideoStats, error) {
	return f.stats, f.err
}
