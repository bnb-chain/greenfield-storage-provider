package types

import "errors"

var (
	// ErrObjectDangling defines the object pointer nil error
	ErrObjectDangling = errors.New("object pointer dangling")
)

// defines the task retry limits
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
	// DefaultUploadObjectRate defines default upload object speed(MB/s)
	DefaultUploadObjectRate = 50
)

// defines the task timeout
const (
	DefaultUploadMinTimeout     = 2
	DefaultUploadMaxTimeout     = 30
	DefaultSealObjectTimeout    = 20
	DefaultGCObjectTimeout      = 60
	DefaultGCZombiePieceTimeout = 60
	DefaultGCStoreTimeout       = 60
)

// ComputeTransferDataTime computes the upload object timeout.
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
