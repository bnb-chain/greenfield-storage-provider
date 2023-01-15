package jobsql

import (
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	types "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/store/jobdb"
)

// need to install mysql
func TestJobMeta(t *testing.T) {
	var (
		txHash = []byte("testHash-" + string(time.Now().Format("2006-01-02 15:04:05")))
		jobId  uint64
	)
	// case1 CreateUploadPayloadJob
	{
		jmi, _ := NewJobMetaImpl(DefaultDBOption)
		_, err := jmi.CreateUploadPayloadJob(
			txHash,
			&types.ObjectInfo{BucketName: "testBucket", ObjectName: "testObject"})
		assert.Equal(t, nil, err)
	}
	fmt.Println(string(txHash))
	// case2 SetObjectCreateHeight/SetObjectCreateHeightAndObjectID
	{
		jmi, _ := NewJobMetaImpl(DefaultDBOption)
		err := jmi.SetObjectCreateHeight(txHash, 100)
		assert.Equal(t, nil, err)
		err = jmi.SetObjectCreateHeightAndObjectID(txHash, 100, 200)
		assert.Equal(t, nil, err)
	}
	// case3 GetObjectInfo
	{
		jmi, _ := NewJobMetaImpl(DefaultDBOption)
		info, err := jmi.GetObjectInfo(txHash)
		assert.Equal(t, nil, err)
		jobId = info.JobId
	}
	// case4 ScanObjectInfo
	{
		jmi, _ := NewJobMetaImpl(DefaultDBOption)
		objects, err := jmi.ScanObjectInfo(0, 10)
		assert.Equal(t, nil, err)
		assert.True(t, len(objects) >= 1)
	}
	// case4 GetJobContext
	{
		jmi, _ := NewJobMetaImpl(DefaultDBOption)
		_, err := jmi.GetJobContext(jobId)
		assert.Equal(t, nil, err)
	}
	// case5 SetUploadPayloadJobState
	{
		jmi, _ := NewJobMetaImpl(DefaultDBOption)
		err := jmi.SetUploadPayloadJobState(jobId, "JOB_STATE_DONE", time.Now().Unix())
		assert.Equal(t, nil, err)
	}
	// case6 SetUploadPayloadJobJobError
	{
		jmi, _ := NewJobMetaImpl(DefaultDBOption)
		err := jmi.SetUploadPayloadJobJobError(jobId, "JOB_STATE_ERROR", "job-err-msg", time.Now().Unix())
		assert.Equal(t, nil, err)
	}
	// clear piecejob table
	{
		jmi, _ := NewJobMetaImpl(DefaultDBOption)
		key1 := hex.EncodeToString([]byte("123"))
		key2 := hex.EncodeToString([]byte("456"))
		jmi.db.Exec("DELETE FROM piece_job where create_hash=?", key1)
		jmi.db.Exec("DELETE FROM piece_job where create_hash=?", key2)
	}

	// case7 SetPrimaryPieceJobDone/GetPrimaryJob
	{
		jmi, _ := NewJobMetaImpl(DefaultDBOption)
		err := jmi.SetPrimaryPieceJobDone([]byte("123"), &jobdb.PieceJob{
			PieceId:         0,
			Checksum:        [][]byte{[]byte("123-0-sum")},
			StorageProvider: "123-0-sp",
		})
		assert.Equal(t, nil, err)
		err = jmi.SetPrimaryPieceJobDone([]byte("123"), &jobdb.PieceJob{
			PieceId:         1,
			Checksum:        [][]byte{[]byte("123-1-sum")},
			StorageProvider: "123-1-sp",
		})
		assert.Equal(t, nil, err)
		pieces, err := jmi.GetPrimaryJob([]byte("123"))
		assert.Equal(t, nil, err)
		assert.Equal(t, 2, len(pieces))
	}
	// case8 SetSecondaryPieceJobDone/GetSecondaryJob
	{
		jmi, _ := NewJobMetaImpl(DefaultDBOption)
		err := jmi.SetSecondaryPieceJobDone([]byte("456"), &jobdb.PieceJob{
			PieceId:         0,
			Checksum:        [][]byte{[]byte("456-0-sum")},
			StorageProvider: "456-0-sp",
		})
		assert.Equal(t, nil, err)
		err = jmi.SetSecondaryPieceJobDone([]byte("456"), &jobdb.PieceJob{
			PieceId:         1,
			Checksum:        [][]byte{[]byte("456-1-sum")},
			StorageProvider: "456-1-sp",
		})
		assert.Equal(t, nil, err)
		pieces, err := jmi.GetSecondaryJob([]byte("456"))
		assert.Equal(t, nil, err)
		assert.Equal(t, 2, len(pieces))
	}
}
