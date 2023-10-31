package video

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParser(t *testing.T) {
	tt := []struct {
		name     string
		input    string
		wantOK   bool
		want     VideoInfo
		isVideo  bool
		asString string
	}{
		{
			name:     "valid episode",
			input:    "foo.S01E01.field.field.field.mp4",
			wantOK:   true,
			want:     VideoInfo{Name: "foo", Extension: "mp4", IsSeries: true, Episode: "S01E01"},
			isVideo:  true,
			asString: "foo.S01E01.mp4",
		},
		{
			name:     "valid multi-episode",
			input:    "foo.S01E01E02.field.field.field.mp4",
			wantOK:   true,
			want:     VideoInfo{Name: "foo", Extension: "mp4", IsSeries: true, Episode: "S01E01E02"},
			isVideo:  true,
			asString: "foo.S01E01E02.mp4",
		},
		{
			name:     "episode subtitles",
			input:    "foo.S01E01.field.field.field.srt",
			wantOK:   true,
			want:     VideoInfo{Name: "foo", Extension: "srt", IsSeries: true, Episode: "S01E01"},
			isVideo:  false,
			asString: "foo.S01E01.srt",
		},
		{
			name:     "movie (dots)",
			input:    "foo.bar.2021.field.field.mp4",
			wantOK:   true,
			want:     VideoInfo{Name: "foo.bar.2021", Extension: "mp4"},
			isVideo:  true,
			asString: "foo.bar.2021.mp4",
		},
		{
			name:     "movie (spaces)",
			input:    "foo bar (2021)-field-field.mp4",
			wantOK:   true,
			want:     VideoInfo{Name: "foo bar (2021)", Extension: "mp4"},
			isVideo:  true,
			asString: "foo bar (2021).mp4",
		},
		{
			name:     "movie (no brackets)",
			input:    "foo bar snafu 2021 1080p DTS AC3 x264-3Li.mkv",
			wantOK:   true,
			want:     VideoInfo{Name: "foo bar snafu 2021", Extension: "mkv"},
			isVideo:  true,
			asString: "foo bar snafu 2021.mkv",
		},
		{
			name:     "movie subtitles (spaces)",
			input:    "foo bar (2021)-field-field.srt",
			wantOK:   true,
			want:     VideoInfo{Name: "foo bar (2021)", Extension: "srt"},
			isVideo:  false,
			asString: "foo bar (2021).srt",
		},
		{
			name:     "movie without year",
			input:    "foo.bar.mp4",
			wantOK:   true,
			want:     VideoInfo{Name: "foo.bar", Extension: "mp4"},
			isVideo:  true,
			asString: "foo.bar.mp4",
		},
		{
			name:     "text file",
			input:    "foo.bar.txt",
			wantOK:   true,
			want:     VideoInfo{Name: "foo.bar", Extension: "txt"},
			isVideo:  false,
			asString: "foo.bar.txt",
		},
		{
			name:   "empty",
			input:  "",
			wantOK: false,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			info, ok := ParseVideoFilename(tc.input)
			assert.Equal(t, tc.wantOK, ok)
			if ok {
				assert.Equal(t, tc.want, info)
				assert.Equal(t, tc.isVideo, info.IsVideo())
				assert.Equal(t, tc.asString, info.String())
			}
		})
	}
}
