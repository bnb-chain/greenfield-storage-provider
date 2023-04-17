package types

import "errors"

var (
	// ErrObjectDangling defines the object pointer nil error
	ErrObjectDangling = errors.New("object pointer dangling")
)

const (
	DefaultReplicatePieceTaskRetryLimit  int64 = 3
	DefaultReplicatePieceTaskMemoryLimit int64 = 40 * 1024 * 1024
	DefaultSealObjectTaskRetryLimit      int64 = 10
	DefaultDownloadObjectTaskRetryLimit  int64 = 3
	DefaultGCObjectTaskRetryLimit        int64 = 3
	DefaultGCZombiePieceTaskRetryLimit   int64 = 3
	DefaultGCStoreTaskRetryLimit         int64 = 3
)

const (
	DefaultUploadObjectRate = 200
)

const (
	DefaultUploadMinTimeout     = 2
	DefaultUploadMaxTimeout     = 10
	DefaultSealObjectTimeout    = 20
	DefaultGCObjectTimeout      = 60
	DefaultGCZombiePieceTimeout = 60
	DefaultGCStoreTimeout       = 60
)

func ComputeTransferDataTime(size uint64) int64 {
	mSize := size / 1024
	timeout := int64(mSize / uint64(DefaultUploadObjectRate))
	if timeout < DefaultUploadMinTimeout {
		return DefaultUploadMinTimeout
	}
	if timeout > DefaultUploadMaxTimeout {
		return DefaultUploadMaxTimeout
	}
	return timeout
}
