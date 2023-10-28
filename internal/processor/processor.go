package processor

import (
	"context"
	"fmt"
	"github.com/clambin/vidconv/internal/feeder"
	"github.com/clambin/vidconv/internal/ffmpeg"
	"github.com/clambin/vidconv/internal/health"
	"github.com/clambin/vidconv/pkg/synclist"
	"log/slog"
	"os"
	"sync/atomic"
	"time"
)

type Processor struct {
	VideoConvertor
	TargetCodec string
	Source      chan feeder.Video
	Logger      *slog.Logger
	processing  synclist.UniqueList[string]
	received    atomic.Int64
	accepted    atomic.Int64
}

var _ health.HealthChecker = &Processor{}

type VideoConvertor interface {
	Convert(ctx context.Context, input, output, targetCodec string) error
}

func New(s chan feeder.Video, l *slog.Logger, targetCodec string) *Processor {
	return &Processor{
		VideoConvertor: ffmpeg.Processor{Logger: l.With("component", "ffmpeg")},
		TargetCodec:    targetCodec,
		Source:         s,
		Logger:         l,
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
				p.Logger.Error("failed to convert file", "err", err, "path", file.Path)
			}
		}
	}
}

func (p *Processor) process(ctx context.Context, file feeder.Video) error {
	l := p.Logger.With("path", file.Path)
	l.Debug("file received")

	p.received.Add(1)

	outfile := file.NameWithCodec(p.TargetCodec)
	fileInfo, err := os.Stat(outfile)
	if err == nil && fileInfo.ModTime().After(file.ModTime) {
		l.Debug("file already converted")
		return nil
	}

	if file.Codec == p.TargetCodec {
		l.Debug("file already in target codec. skipping", "codec", file.Codec)
		return nil
	}

	l.Info("converting file", "inCodec", file.Codec, "outCodec", p.TargetCodec)

	p.processing.Add(file.Path)
	defer p.processing.Remove(file.Path)
	p.accepted.Add(1)

	start := time.Now()
	if err = p.VideoConvertor.Convert(ctx, file.Path, outfile, p.TargetCodec); err != nil {
		if err2 := os.Remove(outfile); err2 != nil {
			l.Error("failed to remove partial output file", "err", err2)
		}
		return fmt.Errorf("conversion failed: %w", err)
	}

	if err = os.Chown(outfile, 568, 568); err != nil {
		l.Warn("failed to set ownership", "err", err)
	}

	compression, err := calculateCompression(file.Path, outfile)
	if err != nil {
		l.Warn("failed to calculate compression ratio", "err", err)
	}

	if err = os.Remove(file.Path); err != nil {
		l.Warn("failed to remove source file", "err", err)
	}

	l.Info("converted file", "inCodec", file.Codec, "outCodec", p.TargetCodec, "compression", compression, "duration", time.Since(start))
	return nil
}

type compressionFactor float64

func (c compressionFactor) LogValue() slog.Value {
	return slog.StringValue(fmt.Sprintf("%.2f", float64(c)))
}

func calculateCompression(before, after string) (compressionFactor, error) {
	fileInfoBefore, err := os.Stat(before)
	if err != nil {
		return 0, fmt.Errorf("stat failed (%s): %w", before, err)
	}
	fileInfoAfter, err := os.Stat(after)
	if err != nil {
		return 0, fmt.Errorf("stat failed (%s): %w", after, err)
	}

	return compressionFactor(float64(fileInfoAfter.Size()) / float64(fileInfoBefore.Size())), nil
}

type Health struct {
	Received   int64
	Accepted   int64
	Processing []string
}

func (p *Processor) Health(_ context.Context) any {
	return Health{
		Received:   p.received.Load(),
		Accepted:   p.accepted.Load(),
		Processing: p.processing.ListOrdered()}
}
