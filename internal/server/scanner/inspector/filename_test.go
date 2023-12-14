package inspector

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_makeTargetFilename(t *testing.T) {
	tt := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "valid episode",
			input: "/tmp/foo.S01E01.field.field.field.mp4",
			want:  "/directory/foo.s01e01.hevc.mkv",
		},
		{
			name:  "valid multi-episode",
			input: "/tmp/foo.S01E01E02.field.field.field.mp4",
			want:  "/directory/foo.s01e01e02.hevc.mkv",
		},
		{
			name:  "movie (dots)",
			input: "/tmp/foo.bar.2021.field.field.mp4",
			want:  "/directory/foo.bar.2021.hevc.mkv",
		},
		{
			name:  "movie (spaces)",
			input: "/tmp/foo bar (2021)-field-field.mp4",
			want:  "/directory/foo bar (2021).hevc.mkv",
		},
		{
			name:  "movie (no brackets)",
			input: "/tmp/foo bar snafu 2021 1080p DTS AC3 x264-3Li.mkv",
			want:  "/directory/foo bar snafu 2021.hevc.mkv",
		},
		{
			name:  "movie without year",
			input: "/tmp/foo.bar.mp4",
			want:  "/directory/foo.bar.hevc.mkv",
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			target := makeTargetFilename(tc.input, "/directory", "hevc", "mkv")
			assert.Equal(t, tc.want, target)
		})
	}
}
