package cmd

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"strings"

	tea "charm.land/bubbletea/v2"
	"codeberg.org/clambin/go-common/charmer"
	"github.com/clambin/xcoder/internal/mediafiles"
	"github.com/clambin/xcoder/internal/transcoder"
	"github.com/clambin/xcoder/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

	var q transcoder.WorkItems

	profileName := v.GetString("profile")
	profile, err := transcoder.GetProfile(profileName)
	if err != nil {
		return fmt.Errorf("invalid profile name %q: %w", profileName, err)
	}

	r, logger, err := getLogger(v)
	if err != nil {
		return fmt.Errorf("invalid logger parameters: %w", err)
	}

	cfg := transcoder.Configuration{
		BaseDir:         args[0],
		Profile:         profile,
		OverwriteTarget: v.GetBool("overwrite"),
		RemoveSource:    v.GetBool("remove"),
	}

	tr := transcoder.New(&q, cfg, logger)
	tr.SetActive(v.GetBool("active"))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	go func() { _ = tr.Run(ctx) }()

	go func() {
		err := mediafiles.FindMediaFiles(cfg.BaseDir, func(path string) {
			tr.AddMediaFile(path)
		})
		if err != nil {
			logger.Error("failed to scan media files", "error", err)
		}
	}()

	u := ui.New(&q, tr, r, ui.DefaultKeyMap(), ui.DefaultStyles())
	a := tea.NewProgram(u, tea.WithoutCatchPanics())
	_, err = a.Run()
	return err
}

func getLogger(v *viper.Viper) (io.Reader, *slog.Logger, error) {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(v.GetString("log.level"))); err != nil {
		return nil, nil, fmt.Errorf("invalid log level %q: %w", v.GetString("log.level"), err)
	}
	opts := slog.HandlerOptions{Level: lvl}

	r, w := io.Pipe()
	switch strings.ToLower(v.GetString("log.format")) {
	case "text":
		return r, slog.New(slog.NewTextHandler(w, &opts)), nil
	case "json":
		return r, slog.New(slog.NewJSONHandler(w, &opts)), nil
	default:
		return nil, nil, fmt.Errorf("invalid log format %q", v.GetString("log.format"))
	}
}
