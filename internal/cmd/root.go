package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"

	"codeberg.org/clambin/go-common/charmer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/clambin/xcoder/internal/pipeline"
	"github.com/clambin/xcoder/internal/tui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
)

var (
	rootCmd = cobra.Command{
		Use:   "xcoder [flags] [directory]",
		Short: "Transcode media files",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUI(cmd.Context(), viper.GetViper(), args)
		},
	}

	configFilename string

	uiArgs = charmer.Arguments{
		"active":     {Default: false, Help: "start processor in active mode"},
		"log.format": {Default: "text", Help: "log format"},
		"log.level":  {Default: "info", Help: "log level"},
		"overwrite":  {Default: false, Help: "overwrite existing files"},
		"remove":     {Default: false, Help: "remove source files after successful transcoding"},
		"profile":    {Default: "hevc-high", Help: "transcoding profile"},
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		panic("could not read build info")
	}
	rootCmd.Version = buildInfo.Main.Version

	cobra.OnInitialize(initConfig)
	rootCmd.Flags().StringVar(&configFilename, "config", "", "Configuration file")
	if err := charmer.SetFlags(&rootCmd, viper.GetViper(), uiArgs); err != nil {
		panic(err)
	}
}

func initConfig() {
	if configFilename != "" {
		viper.SetConfigFile(configFilename)
	} else {
		viper.AddConfigPath(mustConfigDir())
		viper.SetConfigName("config")
	}
	if err := viper.ReadInConfig(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "failed to read config file: "+err.Error())
	}
}

func mustConfigDir() string {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		panic("failed to get user config dir: " + err.Error())
	}
	cfgDir = filepath.Join(cfgDir, "com.github.clambin.xcoder")
	if err = os.MkdirAll(cfgDir, os.ModePerm); err != nil {
		panic("failed to create config dir: " + err.Error())
	}
	return cfgDir
}

func runUI(ctx context.Context, v *viper.Viper, args []string) error {
	if len(args) == 0 {
		args = []string{"."}
	}
	cfg, err := pipeline.GetConfigurationFromViper(v, args)
	if err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	var queue pipeline.Queue
	queue.SetActive(cfg.Active)

	u := tui.New(&queue, cfg)
	a := tea.NewProgram(u, tea.WithAltScreen(), tea.WithoutCatchPanics())

	l := cfg.Logger(u.LogWriter(), nil)
	l.Info("starting program")

	var g errgroup.Group
	subCtx, cancel := context.WithCancel(ctx)

	g.Go(func() error { return pipeline.Run(subCtx, cfg, &queue, l) })

	_, err = a.Run()
	cancel()

	return errors.Join(err, g.Wait())
}
