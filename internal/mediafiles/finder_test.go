package mediafiles

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindMediaFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp(t.TempDir(), "")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "foo.mkv"), []byte{}, 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "foo.txt"), []byte{}, 0644))

	var foundFiles []string
	require.NoError(t, FindMediaFiles(tmpDir, func(path string) {
		foundFiles = append(foundFiles, path)
	}))

	assert.Equal(t, []string{filepath.Join(tmpDir, "foo.mkv")}, foundFiles)
}
