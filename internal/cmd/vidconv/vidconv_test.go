package vidconv

import (
	"context"
	"github.com/clambin/vidconv/internal/inspector/mocks"
	mocks2 "github.com/clambin/vidconv/internal/processor/mocks"
	"github.com/clambin/vidconv/internal/testutil"
	"github.com/clambin/vidconv/pkg/ffmpeg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log/slog"
	"os"
	"testing"
	"time"
)

var validVideoFiles = []string{"foo.mkv"}

func TestApplication_Run(t *testing.T) {
	fs := testutil.MakeTempFS(t, validVideoFiles)
	defer func() {
		assert.NoError(t, os.RemoveAll(fs))
	}()

	a := New(fs, "hevc", ":9090", slog.Default())
	vp := mocks.NewVideoProcessor(t)
	vp.EXPECT().Probe(mock.Anything, mock.AnythingOfType("string")).RunAndReturn(func(_ context.Context, s string) (ffmpeg.Probe, error) {
		return testutil.MakeProbe("h264", 2000, 720, time.Hour), nil
	})
	a.Inspector.VideoProcessor = vp
	vc := mocks2.NewVideoConvertor(t)
	done := make(chan struct{})
	vc.EXPECT().ConvertWithProgress(mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), "hevc", 1800*1024, mock.AnythingOfType("func(ffmpeg.Progress)")).RunAndReturn(func(_ context.Context, _, _, _ string, _ int, _ func(ffmpeg.Progress)) error {
		done <- struct{}{}
		return nil
	})
	a.Processor.VideoConvertor = vc

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error)
	go func() {
		errCh <- a.Run(ctx, 1)
	}()

	<-done
	cancel()
	assert.NoError(t, <-errCh)
}
