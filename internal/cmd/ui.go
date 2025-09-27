package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"codeberg.org/clambin/go-common/charmer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/clambin/xcoder/internal/pipeline"
	"github.com/clambin/xcoder/internal/tui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
)

var (
	uiCmd = &cobra.Command{
		Use:   "ui",
		Short: "Interactively transcode video files",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUI(cmd.Context(), viper.GetViper())
		},
	}

	uiArgs = charmer.Arguments{
		"active":     {Default: false, Help: "start processor in active mode"},
		"input":      {Default: ".", Help: "input directory"},
		"log.format": {Default: "json", Help: "log format"},
		"log.level":  {Default: "info", Help: "log level"},
		"overwrite":  {Default: false, Help: "overwrite existing files"},
		"remove":     {Default: false, Help: "remove source files after successful transcoding"},
		"profile":    {Default: "hevc-high", Help: "transcoding profile"},
	}
)

func init() {
	rootCmd.AddCommand(uiCmd)
	if err := charmer.SetPersistentFlags(uiCmd, viper.GetViper(), uiArgs); err != nil {
		panic(err)
	}
}

func runUI(ctx context.Context, v *viper.Viper) error {
	cfg, err := pipeline.GetConfigurationFromViper(v)
	if err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	var queue pipeline.Queue
	queue.SetActive(cfg.Active)

	u := tui.New(&queue, cfg)
	a := tea.NewProgram(u, tea.WithAltScreen(), tea.WithoutCatchPanics())

	// TODO: respect cfg.Log
	l := slog.New(slog.NewTextHandler(u.LogWriter(), &slog.HandlerOptions{Level: slog.LevelInfo}))
	l.Info("starting program")

	var g errgroup.Group
	subCtx, cancel := context.WithCancel(ctx)

	g.Go(func() error { return pipeline.Run(subCtx, cfg, &queue, l) })

	_, err = a.Run()
	cancel()

	return errors.Join(err, g.Wait())
}
