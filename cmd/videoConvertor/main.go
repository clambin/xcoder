package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "net/http/pprof"

	"github.com/clambin/videoConvertor/internal/configuration"
	"github.com/clambin/videoConvertor/internal/pipeline"
	"github.com/clambin/videoConvertor/internal/ui"
	"github.com/rivo/tview"
	"golang.org/x/sync/errgroup"
)

func main() {
	go func() {
		_ = http.ListenAndServe(":6060", nil)
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := Run(ctx, os.Stderr); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "failed to run. error:", err.Error())
		os.Exit(1)
	}
}

func Run(ctx context.Context, _ io.Writer) error {
	cfg, err := configuration.GetConfiguration()
	if err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	var queue pipeline.Queue
	queue.SetActive(cfg.Active)

	u := ui.New(&queue, cfg)
	l := cfg.Logger(u.LogViewer, nil)
	a := tview.NewApplication().SetRoot(u.Root, true)

	var g errgroup.Group
	subCtx, cancel := context.WithCancel(ctx)
	g.Go(func() error { return pipeline.Run(subCtx, cfg, &queue, l) })
	g.Go(func() error { u.Run(subCtx, a, 250*time.Millisecond); return nil })
	_ = a.Run()

	cancel()
	return g.Wait()
}
