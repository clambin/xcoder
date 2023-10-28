package feeder

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"time"
)

type Video struct {
	Path     string
	ModTime  time.Time
	Info     VideoInfo
	Duration time.Duration
	Codec    string
}

func parseVideoFile(path string, d fs.DirEntry) (Video, error) {
	if d.IsDir() {
		return Video{}, nil
	}
	fileInfo, err := d.Info()
	if err != nil {
		return Video{}, fmt.Errorf("file info: %w", err)
	}
	videoInfo, ok := parseVideoInfo(d.Name())
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
	var filename string
	if s.Info.IsSeries {
		filename = fmt.Sprintf("%s.%s.%s.%s", s.Info.Name, s.Info.Episode, codec, s.Info.Extension)
	} else {
		filename = s.Info.Name + "." + codec + "." + s.Info.Extension
	}
	return filepath.Join(
		filepath.Dir(s.Path),
		filename,
	)
}
