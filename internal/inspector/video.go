package inspector

import (
	"fmt"
	"github.com/clambin/vidconv/pkg/ffmpeg"
	"io/fs"
	"path/filepath"
	"strings"
	"time"
)

type Video struct {
	Path    string
	ModTime time.Time
	Info    VideoInfo
	Stats   ffmpeg.Probe
}

func parseVideoFile(path string, d fs.DirEntry) (Video, error) {
	if d.IsDir() {
		return Video{}, nil
	}
	fileInfo, err := d.Info()
	if err != nil {
		return Video{}, fmt.Errorf("file info: %w", err)
	}
	videoInfo, ok := parseVideoFilename(d.Name())
	if !ok {
		return Video{}, nil
	}
	return Video{
		Path:    path,
		ModTime: fileInfo.ModTime(),
		Info:    videoInfo,
	}, nil
}

func (s Video) String() string {
	return filepath.Join(filepath.Dir(s.Path), s.Info.String())
}

func (s Video) NameWithCodec(codec string) string {
	var components []string
	if s.Info.IsSeries {
		components = []string{s.Info.Name, s.Info.Episode, codec, s.Info.Extension}
	} else {
		components = []string{s.Info.Name, codec, s.Info.Extension}
	}
	return filepath.Join(
		filepath.Dir(s.Path),
		strings.Join(components, "."),
	)
}
