package processor

import (
	"context"
	"errors"
	"fmt"
	"github.com/clambin/vidconv/internal/health"
	"github.com/clambin/vidconv/internal/inspector"
	"github.com/clambin/vidconv/pkg/ffmpeg"
	"github.com/clambin/vidconv/pkg/syncset"
	"log/slog"
	"os"
	"sync/atomic"
	"time"
)

type Processor struct {
	VideoConvertor
	TargetCodec string
	Source      chan inspector.Video
	Logger      *slog.Logger
	received    atomic.Int64
	accepted    atomic.Int64
	processing  *syncset.Set[string]
}

var _ health.HealthChecker = &Processor{}

type VideoConvertor interface {
	ConvertWithProgress(ctx context.Context, input string, output string, targetCodec string, bitrate int, progressCallback func(probe ffmpeg.Progress)) error
}

const (
	loggingInterval = 0.05
)

func New(s chan inspector.Video, targetCodec string, l *slog.Logger) *Processor {
	return &Processor{
		VideoConvertor: ffmpeg.Processor{Logger: l.With(slog.String("component", "ffmpeg"))},
		TargetCodec:    targetCodec,
		Source:         s,
		Logger:         l,
		processing:     syncset.New[string](),
	}
}

func (p *Processor) Run(ctx context.Context) error {
	p.Logger.Info("started")
	defer p.Logger.Info("stopped")
	for {
		select {
		case <-ctx.Done():
			return nil
		case file, ok := <-p.Source:
			if !ok {
				return nil
			}
			if err := p.process(ctx, file); err != nil {
				p.Logger.Error("conversion failed", "err", err, "path", file.Path)
			}
		}
	}
}

func (p *Processor) process(ctx context.Context, file inspector.Video) error {
	l := p.Logger.With(slog.String("path", file.Path))
	l.Debug("file received")

	p.received.Add(1)

	outfile := file.NameWithCodec(p.TargetCodec)
	bitrate, reason, shouldConvert, err := p.shouldConvert(outfile, file)
	if err != nil {
		return fmt.Errorf("shouldConvert: %w", err)
	}
	if !shouldConvert {
		l.Warn("should not convert", slog.String("reason", reason))
		return nil
	}

	l.Info("converting file",
		slog.Any("in", file.Stats),
		slog.Group("out",
			slog.String("codec", p.TargetCodec),
			slog.Int("bitrate", int(float64(bitrate)/1024)),
		),
	)

	p.accepted.Add(1)
	p.processing.Add(file.Path)
	defer p.processing.Remove(file.Path)

	start := time.Now()

	if err := p.convert(ctx, l, file, outfile, bitrate); err != nil {
		return err
	}

	compression, err := calculateCompression(file.Path, outfile)
	if err != nil {
		l.Warn("failed to calculate compression ratio", "err", err)
	}

	if err = os.Remove(file.Path); err != nil {
		l.Warn("failed to remove source file", "err", err)
	}

	l.Info("converted file",
		slog.Any("compression", compression),
		slog.Duration("duration", time.Since(start)),
	)
	return nil
}

func (p *Processor) shouldConvert(outfile string, file inspector.Video) (int, string, bool, error) {
	if outfileInfo, err := os.Stat(outfile); err == nil && outfileInfo.ModTime().After(file.ModTime) {
		return 0, "file already converted", false, nil
	}

	bitrate, ok := getMinimumBitRate(file, p.TargetCodec)
	if !ok {
		return 0, "", false, errors.New("unable to determine minimum bitrate")
	}
	// if we convert video's at the minimum bitrate, we won't get any compression. 20% seems to get reasonable results.
	bitrate = int(float64(bitrate) * 1.2)

	if file.Stats.BitRate() < bitrate {
		return 0, fmt.Sprintf("input video's bitrate is too low (bitrate: %d, minimum: %d)", file.Stats.BitRate()/1024, bitrate/1024), false, nil
	}
	return bitrate, "", true, nil
}

func (p *Processor) convert(ctx context.Context, l *slog.Logger, file inspector.Video, outfile string, bitrate int) error {
	var lastRatio float64
	var called, done int
	err := p.VideoConvertor.ConvertWithProgress(ctx, file.Path, outfile, p.TargetCodec, bitrate, func(progress ffmpeg.Progress) {
		called++
		ratio := float64(progress.Converted) / float64(file.Stats.Duration())
		if ratio-lastRatio < loggingInterval {
			return
		}
		done++
		lastRatio = ratio
		remaining := file.Stats.Duration() - progress.Converted
		remaining = time.Duration(float64(remaining) / progress.Speed)
		l.Info("converting",
			slog.Int("progress(%)", int(100*ratio)),
			slog.Float64("speed", progress.Speed),
			slog.Duration("expected", remaining),
		)
	})

	if err != nil {
		if err2 := os.Remove(outfile); !errors.Is(err2, os.ErrNotExist) {
			l.Error("failed to remove partial output file", "err", err2)
		}
		return err
	}

	if err = os.Chown(outfile, 568, 568); err != nil {
		l.Warn("failed to set ownership", "err", err)
	}
	return nil
}
