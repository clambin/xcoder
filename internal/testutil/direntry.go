package testutil

import (
	"io/fs"
	"time"
)

type FakeDirEntry struct {
	FModTime time.Time
	FName    string
	FIsDir   bool
}

func (f FakeDirEntry) Name() string {
	return f.FName
}

func (f FakeDirEntry) IsDir() bool {
	return f.FIsDir
}

func (f FakeDirEntry) Type() fs.FileMode {
	panic("implement me")
}

func (f FakeDirEntry) Info() (fs.FileInfo, error) {
	return fakeFileInfo{f}, nil
}

var _ fs.DirEntry = FakeDirEntry{}

type fakeFileInfo struct {
	FakeDirEntry
}

func (f fakeFileInfo) Size() int64 {
	panic("implement me")
}

func (f fakeFileInfo) Mode() fs.FileMode {
	panic("implement me")
}

func (f fakeFileInfo) ModTime() time.Time {
	return f.FakeDirEntry.FModTime
}

func (f fakeFileInfo) Sys() any {
	panic("implement me")
}

var _ fs.FileInfo = &fakeFileInfo{}
