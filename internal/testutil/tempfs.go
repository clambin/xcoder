package testutil

import (
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func MakeTempFS(t *testing.T, videoFiles []string) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "videofiles"), 0755))
	for _, videoFile := range videoFiles {
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "videofiles", videoFile), nil, 0644))
	}
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "videofiles", "foo.2021.srt"), nil, 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "videofiles", "foo.2021.hid"), nil, 0000))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "hidden"), 0000))

	return tmpDir
}
