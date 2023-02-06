package jobsql

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	ptypesv1pb "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/store/jobdb"
)

// need to install mysql
func TestJobMeta(t *testing.T) {
	var (
		txHash = []byte("testHash-" + string(time.Now().Format("2006-01-02 15:04:05")))
		jobID  uint64
	)
	// case1 CreateUploadPayloadJob
	{
		jmi, _ := NewJobMetaImpl(DefaultJobSqlDBConfig)
		_, err := jmi.CreateUploadPayloadJob(
			txHash,
			&ptypesv1pb.ObjectInfo{BucketName: "testBucket", ObjectName: "testObject"})
		assert.Equal(t, nil, err)
	}
	// case2 SetObjectCreateHeight/SetObjectCreateHeightAndObjectID
	{
		jmi, _ := NewJobMetaImpl(DefaultJobSqlDBConfig)
		err := jmi.SetObjectCreateHeight(txHash, 100)
		assert.Equal(t, nil, err)
		err = jmi.SetObjectCreateHeightAndObjectID(txHash, 100, 200)
		assert.Equal(t, nil, err)
	}
	// case3 GetObjectInfo
	{
		jmi, _ := NewJobMetaImpl(DefaultJobSqlDBConfig)
		info, err := jmi.GetObjectInfo(txHash)
		assert.Equal(t, nil, err)
		jobID = info.JobId
	}
	// case4 ScanObjectInfo
	{
		jmi, _ := NewJobMetaImpl(DefaultJobSqlDBConfig)
		objects, err := jmi.ScanObjectInfo(0, 10)
		assert.Equal(t, nil, err)
		assert.True(t, len(objects) >= 1)
	}
	// case4 GetJobContext
	{
		jmi, _ := NewJobMetaImpl(DefaultJobSqlDBConfig)
		_, err := jmi.GetJobContext(jobID)
		assert.Equal(t, nil, err)
	}
	// case5 SetUploadPayloadJobState
	{
		jmi, _ := NewJobMetaImpl(DefaultJobSqlDBConfig)
		err := jmi.SetUploadPayloadJobState(jobID, "JOB_STATE_DONE", time.Now().Unix())
		assert.Equal(t, nil, err)
	}
	// case6 SetUploadPayloadJobJobError
	{
		jmi, _ := NewJobMetaImpl(DefaultJobSqlDBConfig)
		err := jmi.SetUploadPayloadJobJobError(jobID, "JOB_STATE_ERROR", "job-err-msg", time.Now().Unix())
		assert.Equal(t, nil, err)
	}
	// clear piece_job table
	{
		jmi, _ := NewJobMetaImpl(DefaultJobSqlDBConfig)
		key1 := hex.EncodeToString([]byte("123"))
		key2 := hex.EncodeToString([]byte("456"))
		jmi.db.Exec("DELETE FROM piece_job where create_hash=?", key1)
		jmi.db.Exec("DELETE FROM piece_job where create_hash=?", key2)
	}

	// case7 SetPrimaryPieceJobDone/GetPrimaryJob
	{
		jmi, _ := NewJobMetaImpl(DefaultJobSqlDBConfig)
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
		jmi, _ := NewJobMetaImpl(DefaultJobSqlDBConfig)
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

func TestJobMetaV2(t *testing.T) {
	var (
		objectID = uint64(time.Now().Unix())
		jobID    uint64
	)
	// case1 CreateUploadPayloadJobV2
	{
		jmi, _ := NewJobMetaImpl(DefaultJobSqlDBConfig)
		_, err := jmi.CreateUploadPayloadJobV2(
			&ptypesv1pb.ObjectInfo{BucketName: "testBucket", ObjectName: "testObject", ObjectId: objectID})
		assert.Equal(t, nil, err)
	}
	// case2 GetObjectInfoV2
	{
		jmi, _ := NewJobMetaImpl(DefaultJobSqlDBConfig)
		info, err := jmi.GetObjectInfoV2(objectID)
		assert.Equal(t, nil, err)
		jobID = info.JobId
	}
	// case3 ScanObjectInfoV2
	{
		jmi, _ := NewJobMetaImpl(DefaultJobSqlDBConfig)
		objects, err := jmi.ScanObjectInfoV2(0, 10)
		assert.Equal(t, nil, err)
		assert.True(t, len(objects) >= 1)
	}
	// case4 GetJobContextV2
	{
		jmi, _ := NewJobMetaImpl(DefaultJobSqlDBConfig)
		_, err := jmi.GetJobContextV2(jobID)
		assert.Equal(t, nil, err)
	}
	// case5 SetUploadPayloadJobStateV2
	{
		jmi, _ := NewJobMetaImpl(DefaultJobSqlDBConfig)
		err := jmi.SetUploadPayloadJobState(jobID, "JOB_STATE_DONE", time.Now().Unix())
		assert.Equal(t, nil, err)
	}
	// case6 SetUploadPayloadJobJobErrorV2
	{
		jmi, _ := NewJobMetaImpl(DefaultJobSqlDBConfig)
		err := jmi.SetUploadPayloadJobJobError(jobID, "JOB_STATE_ERROR", "job-err-msg", time.Now().Unix())
		assert.Equal(t, nil, err)
	}
	// clear piece_job_v2 table
	{
		jmi, _ := NewJobMetaImpl(DefaultJobSqlDBConfig)
		jmi.db.Exec("DELETE FROM piece_job_v2 where object_id=?", 123)
		jmi.db.Exec("DELETE FROM piece_job_v2 where object_id=?", 456)
	}
	// case7 SetPrimaryPieceJobDoneV2/GetPrimaryJobV2
	{
		jmi, _ := NewJobMetaImpl(DefaultJobSqlDBConfig)
		err := jmi.SetPrimaryPieceJobDoneV2(123, &jobdb.PieceJob{
			PieceId:         0,
			Checksum:        [][]byte{[]byte("123-0-sum")},
			StorageProvider: "123-0-sp",
		})
		assert.Equal(t, nil, err)
		err = jmi.SetPrimaryPieceJobDoneV2(123, &jobdb.PieceJob{
			PieceId:         1,
			Checksum:        [][]byte{[]byte("123-1-sum")},
			StorageProvider: "123-1-sp",
		})
		assert.Equal(t, nil, err)
		pieces, err := jmi.GetPrimaryJobV2(123)
		assert.Equal(t, nil, err)
		assert.Equal(t, 2, len(pieces))
	}
	// case8 SetSecondaryPieceJobDoneV2/GetSecondaryJobV2
	{
		jmi, _ := NewJobMetaImpl(DefaultJobSqlDBConfig)
		err := jmi.SetSecondaryPieceJobDoneV2(456, &jobdb.PieceJob{
			PieceId:         0,
			Checksum:        [][]byte{[]byte("456-0-sum")},
			StorageProvider: "456-0-sp",
		})
		assert.Equal(t, nil, err)
		err = jmi.SetSecondaryPieceJobDoneV2(456, &jobdb.PieceJob{
			PieceId:         1,
			Checksum:        [][]byte{[]byte("456-1-sum")},
			StorageProvider: "456-1-sp",
		})
		assert.Equal(t, nil, err)
		pieces, err := jmi.GetSecondaryJobV2(456)
		assert.Equal(t, nil, err)
		assert.Equal(t, 2, len(pieces))
	}
}
