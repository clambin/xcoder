package feeder

import (
	"github.com/stretchr/testify/assert"
	"io/fs"
	"testing"
	"time"
)

func Test_parseVideoFile(t *testing.T) {
	timestamp := time.Date(2023, time.October, 30, 0, 0, 0, 0, time.UTC)
	type args struct {
		path string
		d    fs.DirEntry
	}
	tests := []struct {
		name     string
		args     args
		want     Video
		wantErr  assert.ErrorAssertionFunc
		isVideo  bool
		asString string
	}{
		{
			name: "valid episode",
			args: args{
				path: "foo/foo.bar.S01E01.field.field.field.mp4",
				d: &fakeDirEntry{
					name:    "foo.bar.S01E01.field.field.field.mp4",
					modTime: timestamp,
				},
			},
			want: Video{
				Path:    "foo/foo.bar.S01E01.field.field.field.mp4",
				ModTime: timestamp,
				Info: VideoInfo{
					Name:      "foo.bar",
					Extension: "mp4",
					IsSeries:  true,
					Episode:   "S01E01",
				},
			},
			wantErr: assert.NoError,
			isVideo: true,
		},
		{
			name: "valid move",
			args: args{
				path: "foo/foo.bar.2021.field.field.field.mp4",
				d: &fakeDirEntry{
					name:    "foo.bar.2021.field.field.field.mp4",
					modTime: timestamp,
				},
			},
			want: Video{
				Path:    "foo/foo.bar.2021.field.field.field.mp4",
				ModTime: timestamp,
				Info: VideoInfo{
					Name:      "foo.bar.2021",
					Extension: "mp4",
				},
			},
			wantErr: assert.NoError,
			isVideo: true,
		},
		{
			name: "directory",
			args: args{
				path: "foo/bar",
				d: &fakeDirEntry{
					name:    "bar",
					isDir:   true,
					modTime: timestamp,
				},
			},
			wantErr: assert.NoError,
			isVideo: false,
		},
		{
			name: "text file",
			args: args{
				path: "foo/release.txt",
				d: &fakeDirEntry{
					name:    "release.txt",
					modTime: timestamp,
				},
			},
			wantErr: assert.NoError,
			isVideo: false,
		},
		{
			name: "empty",
			args: args{
				path: "",
				d: &fakeDirEntry{
					modTime: timestamp,
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseVideoFile(tt.args.path, tt.args.d)
			tt.wantErr(t, err)
			if err == nil {
				assert.Equal(t, tt.isVideo, got.Info.IsVideo())
				if tt.isVideo {
					assert.Equal(t, tt.want, got)
				}
			}
		})
	}
}

func TestVideo_String(t *testing.T) {
	v := Video{
		Path: "/tmp/foo.bar.2021.x264.mkv",
		Info: VideoInfo{
			Name:      "foo.bar.2021",
			Extension: "mkv",
		},
	}

	assert.Equal(t, "/tmp/foo.bar.2021.mkv", v.String())
	assert.Equal(t, "/tmp/foo.bar.2021.hevc.mkv", v.NameWithCodec("hevc"))

	v = Video{
		Path: "/tmp/foo.bar.S02E21.x264.mkv",
		Info: VideoInfo{
			Name:      "foo.bar",
			Extension: "mkv",
			IsSeries:  true,
			Episode:   "S02E21",
		},
	}

	assert.Equal(t, "/tmp/foo.bar.S02E21.mkv", v.String())
	assert.Equal(t, "/tmp/foo.bar.S02E21.hevc.mkv", v.NameWithCodec("hevc"))
}

type fakeDirEntry struct {
	name    string
	isDir   bool
	modTime time.Time
}

func (f fakeDirEntry) Name() string {
	return f.name
}

func (f fakeDirEntry) IsDir() bool {
	return f.isDir
}

func (f fakeDirEntry) Type() fs.FileMode {
	panic("implement me")
}

func (f fakeDirEntry) Info() (fs.FileInfo, error) {
	return fakeFileInfo{f}, nil
}

var _ fs.DirEntry = fakeDirEntry{}

type fakeFileInfo struct {
	fakeDirEntry
}

func (f fakeFileInfo) Size() int64 {
	panic("implement me")
}

func (f fakeFileInfo) Mode() fs.FileMode {
	panic("implement me")
}

func (f fakeFileInfo) ModTime() time.Time {
	return f.fakeDirEntry.modTime
}

func (f fakeFileInfo) Sys() any {
	panic("implement me")
}

var _ fs.FileInfo = &fakeFileInfo{}
