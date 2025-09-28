package pipeline

import (
	"cmp"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Configuration struct {
	Log
	Input       string
	ProfileName string
	Profile     Profile
	Active      bool
	Remove      bool
	Overwrite   bool
}

type Log struct {
	Level  string
	Format string
}

// Logger returns a new logger with the given options, writing to w.
// If w is nil, it defaults to os.Stderr.
func (l Log) Logger(w io.Writer, opts *slog.HandlerOptions) *slog.Logger {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	if strings.ToLower(l.Level) == "debug" {
		opts = &slog.HandlerOptions{Level: slog.LevelDebug}
	}
	var h slog.Handler
	switch strings.ToLower(l.Format) {
	case "json":
		h = slog.NewJSONHandler(cmp.Or[io.Writer](w, os.Stderr), opts)
	case "text":
		h = slog.NewTextHandler(cmp.Or[io.Writer](w, os.Stderr), opts)
	default:
		panic(fmt.Sprintf("invalid format: %s", l.Format))
	}
	return slog.New(h)
}

func GetConfigurationFromViper(v *viper.Viper, args []string) (cfg Configuration, err error) {
	if len(args) != 1 {
		return cfg, fmt.Errorf("invalid number of arguments: %d", len(args))
	}
	cfg.Input = args[0]
	cfg.Active = v.GetBool("active")
	cfg.Format = v.GetString("log.format")
	cfg.Level = v.GetString("log.level")
	cfg.Overwrite = v.GetBool("overwrite")
	cfg.Remove = v.GetBool("remove")
	cfg.ProfileName = v.GetString("profile")
	if cfg.Profile, err = GetProfile(cfg.ProfileName); err != nil {
		return Configuration{}, err
	}
	return cfg, nil
}
