package mediafiles

import (
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
)

var validExtensions = []string{".mp4", ".mkv", ".avi", ".mov"}

// FindMediaFiles iterates through all media files below baseDir and calls f(path)
func FindMediaFiles(baseDir string, f func(string)) error {
	return filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if slices.Contains(validExtensions, strings.ToLower(filepath.Ext(path))) {
			f(path)
		}
		return nil
	})
}
