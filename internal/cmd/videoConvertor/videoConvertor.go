package videoConvertor

import (
	"context"
	"github.com/clambin/videoConvertor/internal/configuration"
	"github.com/clambin/videoConvertor/internal/converter"
	"github.com/clambin/videoConvertor/internal/ffmpeg"
	"github.com/clambin/videoConvertor/internal/preprocessor"
	"github.com/clambin/videoConvertor/internal/scanner"
	"github.com/clambin/videoConvertor/internal/ui"
	"github.com/clambin/videoConvertor/internal/worklist"
	"github.com/rivo/tview"
	"golang.org/x/sync/errgroup"
	"io"
	"log/slog"
	"time"
)

func Run(ctx context.Context, _ io.Writer) error {
	cfg, err := configuration.GetConfiguration()
	if err != nil {
		slog.Error("invalid configuration", "err", err)

	}

	var list worklist.WorkList
	list.SetActive(cfg.Active)

	u := ui.New(&list, cfg)

	var opt slog.HandlerOptions
	if cfg.Debug {
		opt.Level = slog.LevelDebug
	}
	l := slog.New(slog.NewTextHandler(u.LogViewer, nil))

	ff := ffmpeg.Processor{Logger: l.With("component", "ffmpeg")}
	c := converter.New(&ff, &list, cfg, l.With("component", "converter"))

	a := tview.NewApplication().SetRoot(u.Root, true)

	subCtx, cancel := context.WithCancel(ctx)
	itemCh := make(chan *worklist.WorkItem)

	var g errgroup.Group
	g.Go(func() error { return scanner.Scan(subCtx, cfg.Input, &list, itemCh, l.With("component", "scanner")) })
	const inspectorCount = 4
	for range inspectorCount {
		g.Go(func() error {
			preprocessor.Run(subCtx, itemCh, &ff, cfg.Profile, l.With("component", "preprocessor"))
			return nil
		})
	}
	const converterCount = 2
	for range converterCount {
		g.Go(func() error { c.Run(subCtx); return nil })
	}

	g.Go(func() error { u.Run(subCtx, a, 250*time.Millisecond); return nil })
	_ = a.Run()
	cancel()
	return g.Wait()
}
