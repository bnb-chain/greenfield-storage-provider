package storage

import (
	"fmt"
	"io/fs"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/pkg/sftp"
	"github.com/stretchr/testify/assert"
)

type fakeFileInfo struct {
	dir      bool
	basename string
	modTime  time.Time
	contents string
}

func (f *fakeFileInfo) Name() string       { return f.basename }
func (f *fakeFileInfo) Sys() any           { return fakeFileInfoSys() }
func (f *fakeFileInfo) ModTime() time.Time { return f.modTime }
func (f *fakeFileInfo) IsDir() bool        { return f.dir }
func (f *fakeFileInfo) Size() int64        { return int64(len(f.contents)) }
func (f *fakeFileInfo) Mode() fs.FileMode {
	if f.dir {
		return 0755 | fs.ModeDir
	}
	return 0644
}
func fakeFileInfoSys() interface{} {
	return &syscall.Stat_t{Uid: 65501, Gid: 65501}
}

type fakeSFTPInfo struct {
	dir      bool
	basename string
	modTime  time.Time
	contents string
}

func (f *fakeSFTPInfo) Name() string       { return f.basename }
func (f *fakeSFTPInfo) Sys() any           { return fakeSFTPFileInfoSys() }
func (f *fakeSFTPInfo) ModTime() time.Time { return f.modTime }
func (f *fakeSFTPInfo) IsDir() bool        { return f.dir }
func (f *fakeSFTPInfo) Size() int64        { return int64(len(f.contents)) }
func (f *fakeSFTPInfo) Mode() fs.FileMode {
	if f.dir {
		return 0755 | fs.ModeDir
	}
	return 0644
}
func fakeSFTPFileInfoSys() interface{} {
	return &sftp.FileStat{UID: 65500, GID: 65500}
}

func Test_getOwnerGroup(t *testing.T) {
	cases := []struct {
		name          string
		info          os.FileInfo
		wantedResult1 string
		wantedResult2 string
	}{
		{
			name: "syscall Stat_t",
			info: &fakeFileInfo{
				basename: "mock",
				modTime:  time.Date(2023, 8, 1, 10, 30, 0, 0, time.UTC),
				contents: "mock contents",
			},
			wantedResult1: "65501",
			wantedResult2: "65501",
		},
		{
			name: "sftp FileStat",
			info: &fakeSFTPInfo{
				basename: "mock sftp",
				modTime:  time.Date(2023, 7, 1, 11, 30, 0, 0, time.UTC),
				contents: "mock sftp contents",
			},
			wantedResult1: "65500",
			wantedResult2: "65500",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result1, result2 := getOwnerGroup(tt.info)
			assert.Equal(t, tt.wantedResult1, result1)
			assert.Equal(t, tt.wantedResult2, result2)
		})
	}
}

func Test_userName(t *testing.T) {
	uidMap[1] = "test1"
	cases := []struct {
		name         string
		uid          int
		wantedResult string
	}{
		{
			name:         "1",
			uid:          0,
			wantedResult: "root",
		},
		{
			name:         "2",
			uid:          1,
			wantedResult: "test1",
		},
		{
			name:         "3",
			uid:          100000,
			wantedResult: "100000",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := userName(tt.uid)
			assert.Equal(t, tt.wantedResult, result)
			fmt.Println(result)
		})
	}
}

func Test_groupName(t *testing.T) {
	gidMap[9999] = "test1"
	cases := []struct {
		name         string
		gid          int
		wantedResult string
	}{
		{
			name:         "1",
			gid:          9999,
			wantedResult: "test1",
		},
		{
			name:         "2",
			gid:          1,
			wantedResult: "daemon",
		},
		{
			name:         "3",
			gid:          100000,
			wantedResult: "100000",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := groupName(tt.gid)
			assert.Equal(t, tt.wantedResult, result)
			fmt.Println(result)
		})
	}
}
