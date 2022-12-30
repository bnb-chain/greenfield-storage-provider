package database

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	types "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
)

// need to install mysql
func TestJobMeta(t *testing.T) {
	var (
		txHash = []byte("testHash-" + string(time.Now().Format("2006-01-02 15:04:05")))
		jobId  uint64
	)
	// case1 CreateUploadPayloadJob
	{
		jmi, _ := NewJobMetaImpl()
		err := jmi.CreateUploadPayloadJob(
			txHash,
			&types.ObjectInfo{BucketName: "testBucket", ObjectName: "testObject"})
		assert.Equal(t, nil, err)
	}
	fmt.Println(string(txHash))
	// case2 SetObjectCreateHeight
	{
		jmi, _ := NewJobMetaImpl()
		err := jmi.SetObjectCreateHeight(txHash, 100)
		assert.Equal(t, nil, err)
	}
	// case3 GetObjectInfo
	{
		jmi, _ := NewJobMetaImpl()
		info, err := jmi.GetObjectInfo(txHash)
		assert.Equal(t, nil, err)
		jobId = info.JobId
	}
	// case4 ScanObjectInfo
	{
		jmi, _ := NewJobMetaImpl()
		objects, err := jmi.ScanObjectInfo(0, 10)
		assert.Equal(t, nil, err)
		assert.True(t, len(objects) > 1)
	}
	// case4 GetJobContext
	{
		jmi, _ := NewJobMetaImpl()
		_, err := jmi.GetJobContext(jobId)
		assert.Equal(t, nil, err)
	}
	// case5 SetUploadPayloadJobState
	{
		jmi, _ := NewJobMetaImpl()
		err := jmi.SetUploadPayloadJobState(jobId, "JOB_STATE_DONE", time.Now().Unix())
		assert.Equal(t, nil, err)
	}
	// case6 SetUploadPayloadJobJobError
	{
		jmi, _ := NewJobMetaImpl()
		err := jmi.SetUploadPayloadJobJobError(jobId, "JOB_STATE_ERROR", "job-err-msg", time.Now().Unix())
		assert.Equal(t, nil, err)
	}
	// clear piecejob table
	{
		jmi, _ := NewJobMetaImpl()
		jmi.db.Exec("DELETE FROM piece_job where create_hash='123'")
		jmi.db.Exec("DELETE FROM piece_job where create_hash='456'")
	}

	// case7 SetPrimaryPieceJobDone/GetPrimaryJob
	{
		jmi, _ := NewJobMetaImpl()
		err := jmi.SetPrimaryPieceJobDone([]byte("123"), &PieceJob{
			PieceId:         0,
			CheckSum:        []byte("123-0-sum"),
			StorageProvider: "123-0-sp",
		})
		assert.Equal(t, nil, err)
		err = jmi.SetPrimaryPieceJobDone([]byte("123"), &PieceJob{
			PieceId:         1,
			CheckSum:        []byte("123-1-sum"),
			StorageProvider: "123-1-sp",
		})
		assert.Equal(t, nil, err)
		pieces, err := jmi.GetPrimaryJob([]byte("123"))
		assert.Equal(t, nil, err)
		assert.Equal(t, 2, len(pieces))
	}
	// case8 SetSecondaryPieceJobDone/GetSecondaryJob
	{
		jmi, _ := NewJobMetaImpl()
		err := jmi.SetSecondaryPieceJobDone([]byte("456"), &PieceJob{
			PieceId:         0,
			CheckSum:        []byte("456-0-sum"),
			StorageProvider: "456-0-sp",
		})
		assert.Equal(t, nil, err)
		err = jmi.SetSecondaryPieceJobDone([]byte("456"), &PieceJob{
			PieceId:         1,
			CheckSum:        []byte("456-1-sum"),
			StorageProvider: "456-1-sp",
		})
		assert.Equal(t, nil, err)
		pieces, err := jmi.GetSecondaryJob([]byte("456"))
		assert.Equal(t, nil, err)
		assert.Equal(t, 2, len(pieces))
	}
}
