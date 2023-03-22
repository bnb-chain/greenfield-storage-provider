package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

const (
	dirSuffix = "/"
	windowsOS = "windows"
)

type diskFileStore struct {
	root string
	DefaultObjectStorage
}

func newDiskFileStore(cfg ObjectStorageConfig) (ObjectStorage, error) {
	// For Windows, the path looks like /C:/a/b/c/
	endPoint := cfg.BucketURL
	if runtime.GOOS == windowsOS && strings.HasPrefix(endPoint, "/") {
		endPoint = endPoint[1:]
	}
	return &diskFileStore{root: endPoint}, nil
}

func (d *diskFileStore) String() string {
	if runtime.GOOS == windowsOS {
		return "file:///" + d.root
	}
	return "file://" + d.root
}

func (d *diskFileStore) CreateBucket(ctx context.Context) error {
	rootPath := d.root
	log.Debugf("directory path: %s", rootPath)
	if err := os.MkdirAll(rootPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s : %q", rootPath, err)
	}
	return nil
}

func (d *diskFileStore) GetObject(ctx context.Context, key string, offset, limit int64) (io.ReadCloser, error) {
	p := d.path(key)

	f, err := os.Open(p)
	if err != nil {
		log.Errorw("failed to get object due to open file", "error", err)
		return nil, err
	}

	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		log.Errorw("failed to get object due to stat file", "error", err)
		return nil, err
	}
	if info.IsDir() {
		_ = f.Close()
		return io.NopCloser(bytes.NewBuffer([]byte{})), nil
	}

	if offset > 0 {
		if _, err = f.Seek(offset, 0); err != nil {
			_ = f.Close()
			return nil, err
		}
	}
	if limit > 0 {
		defer f.Close()
		buf := make([]byte, limit)
		n, err := f.Read(buf)
		if err != nil {
			log.Errorw("failed to get object due to read file", "error", err)
			return nil, err
		}
		return io.NopCloser(bytes.NewBuffer(buf[:n])), nil
	}
	return f, nil
}

func (d *diskFileStore) PutObject(ctx context.Context, key string, reader io.Reader) error {
	p := d.path(key)

	if strings.HasSuffix(key, dirSuffix) || key == "" && strings.HasSuffix(d.root, dirSuffix) {
		return os.MkdirAll(p, os.FileMode(0755))
	}

	tmp := filepath.Join(filepath.Dir(p), "."+filepath.Base(p)+".tmp"+strconv.Itoa(rand.Int()))
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(filepath.Dir(p), os.FileMode(0755)); err != nil {
			log.Errorw("failed to put object due to mkdir", "error", err)
			return err
		}
		f, err = os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	}
	if err != nil {
		log.Errorw("failed to put object due to open file", "error", err)
		return err
	}
	defer func() {
		if err != nil {
			_ = os.Remove(tmp)
		}
	}()

	buf := bufPool.Get().(*[]byte)
	defer bufPool.Put(buf)
	_, err = io.CopyBuffer(f, reader, *buf)
	if err != nil {
		log.Errorw("failed to put object due to copy buffer", "error", err)
		_ = f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		log.Errorw("failed to put object due to close file", "error", err)
		return err
	}

	return os.Rename(tmp, p)
}

func (d *diskFileStore) DeleteObject(ctx context.Context, key string) error {
	err := os.Remove(d.path(key))
	if err != nil && os.IsNotExist(err) {
		log.Errorw("failed to delete object due to remove file", "error", err)
		err = nil
	}
	return err
}

func (d *diskFileStore) HeadBucket(ctx context.Context) error {
	if _, err := os.Stat(d.root); err != nil {
		if os.IsNotExist(err) {
			return errors.ErrNoSuchBucket
		}
		return err
	}
	return nil
}

func (d *diskFileStore) HeadObject(ctx context.Context, key string) (Object, error) {
	p := d.path(key)
	fileInfo, err := os.Stat(p)
	if err != nil {
		log.Errorw("failed to Head object due to Stat file", "error", err)
		return nil, err
	}
	size := fileInfo.Size()
	if fileInfo.IsDir() {
		size = 0
	}
	owner, group := getOwnerGroup(fileInfo)

	var isSymlink bool
	return &file{
		object{
			key,
			size,
			fileInfo.ModTime(),
			fileInfo.IsDir(),
		},
		owner,
		group,
		fileInfo.Mode(),
		isSymlink,
	}, nil
}

func (d *diskFileStore) path(key string) string {
	return filepath.Join(d.root, key)
}
