package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

var (
	verifyCmd = &cobra.Command{
		Use:   "verify",
		Short: "Verify media files",
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				verify(cmd.Context(), arg)
			}
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(verifyCmd)
}

func verify(ctx context.Context, path string) {
	panic("broken")
	// TODO: move this to ffmpeg
	/*
		l := slog.New(slog.NewTextHandler(os.Stderr, nil))
		t := transcoder.Transcoder{Logger: l.With("component", "transcoder")}
		stats, err := t.Scan(ctx, path)
		if err != nil {
			fmt.Printf("\r%s FAIL: %v\n", path, err)
			return
		}

		tmpDir, err := os.MkdirTemp(os.TempDir(), "xcoder")
		if err != nil {
			panic(err)
		}
		tempSocketPath := filepath.Join(tmpDir, "ffmpeg-verify.sock")

		f := ffmpeg.Decode(path, "-hwaccel", "videotoolbox").Muxer("null").LogLevel("error").NoStats().Progress(func(p ffmpeg.Progress) {
			progress := p.Converted.Seconds() / stats.Duration.Seconds()
			fmt.Printf("\r%s ... %.1f%%", path, 100*progress)
		}, tempSocketPath)

		err = f.Run(ctx)
		if err == nil {
			fmt.Println(" PASS")
		} else {
			fmt.Println(" FAIL: ", err.Error())
		}
	*/
}
