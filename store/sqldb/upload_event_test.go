package sqldb

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	storage_types "github.com/bnb-chain/greenfield/x/storage/types"
)

func TestSpDBImpl_InsertPutEventSuccess1(t *testing.T) {
	t.Log("Success case description: upload object task")
	task := &gfsptask.GfSpUploadObjectTask{
		Task: &gfsptask.GfSpTask{
			Address:      "mockAddress",
			CreateTime:   1,
			UpdateTime:   2,
			Timeout:      0,
			TaskPriority: 1,
			Retry:        1,
			MaxRetry:     1,
			UserAddress:  "mockUserAddress",
			Logs:         "mockLogs",
			Err:          gfsperrors.MakeGfSpError(mockDBInternalError),
		},
		VirtualGroupFamilyId: 1,
		ObjectInfo: &storage_types.ObjectInfo{
			Id:         sdkmath.NewUint(10),
			BucketName: "mockBucketName",
			ObjectName: "mockObjectName",
		},
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `put_object_event_log` (`update_time`,`object_id`,`bucket`,`object`,`state`,`error`,`logs`) VALUES (?,?,?,?,?,?,?)").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.InsertPutEvent(task)
	assert.Nil(t, err)
}

func TestSpDBImpl_InsertPutEventSuccess2(t *testing.T) {
	t.Log("Success case description: replicate piece task")
	task := &gfsptask.GfSpReplicatePieceTask{
		Task: &gfsptask.GfSpTask{
			Address:      "mockAddress",
			CreateTime:   1,
			UpdateTime:   2,
			Timeout:      0,
			TaskPriority: 1,
			Retry:        1,
			MaxRetry:     1,
			UserAddress:  "mockUserAddress",
			Logs:         "mockLogs",
			Err:          gfsperrors.MakeGfSpError(mockDBInternalError),
		},
		ObjectInfo: &storage_types.ObjectInfo{
			Id:         sdkmath.NewUint(10),
			BucketName: "mockBucketName",
			ObjectName: "mockObjectName",
		},
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `put_object_event_log` (`update_time`,`object_id`,`bucket`,`object`,`state`,`error`,`logs`) VALUES (?,?,?,?,?,?,?)").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.InsertPutEvent(task)
	assert.Nil(t, err)
}

func TestSpDBImpl_InsertPutEventSuccess3(t *testing.T) {
	t.Log("Success case description: seal object task")
	task := &gfsptask.GfSpSealObjectTask{
		Task: &gfsptask.GfSpTask{
			Address:      "mockAddress",
			CreateTime:   1,
			UpdateTime:   2,
			Timeout:      0,
			TaskPriority: 1,
			Retry:        1,
			MaxRetry:     1,
			UserAddress:  "mockUserAddress",
			Logs:         "mockLogs",
			Err:          gfsperrors.MakeGfSpError(mockDBInternalError),
		},
		ObjectInfo: &storage_types.ObjectInfo{
			Id:         sdkmath.NewUint(10),
			BucketName: "mockBucketName",
			ObjectName: "mockObjectName",
		},
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `put_object_event_log` (`update_time`,`object_id`,`bucket`,`object`,`state`,`error`,`logs`) VALUES (?,?,?,?,?,?,?)").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.InsertPutEvent(task)
	assert.Nil(t, err)
}

func TestSpDBImpl_InsertUploadEventSuccess1(t *testing.T) {
	t.Log("Success case description: task.Error is not nil")
	task := &gfsptask.GfSpUploadObjectTask{
		Task: &gfsptask.GfSpTask{
			Address:      "mockAddress",
			CreateTime:   1,
			UpdateTime:   2,
			Timeout:      0,
			TaskPriority: 1,
			Retry:        1,
			MaxRetry:     1,
			UserAddress:  "mockUserAddress",
			Logs:         "mockLogs",
			Err:          gfsperrors.MakeGfSpError(mockDBInternalError),
		},
		VirtualGroupFamilyId: 1,
		ObjectInfo: &storage_types.ObjectInfo{
			Id:         sdkmath.NewUint(10),
			BucketName: "mockBucketName",
			ObjectName: "mockObjectName",
		},
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `put_object_event_log` (`update_time`,`object_id`,`bucket`,`object`,`state`,`error`,`logs`) VALUES (?,?,?,?,?,?,?)").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.InsertUploadEvent(task)
	assert.Nil(t, err)
}

func TestSpDBImpl_InsertUploadEventSuccess2(t *testing.T) {
	t.Log("Success case description: task.Error is nil")
	task := &gfsptask.GfSpUploadObjectTask{
		Task: &gfsptask.GfSpTask{
			Address:      "mockAddress",
			CreateTime:   time.Now().Unix() - 3,
			UpdateTime:   2,
			Timeout:      0,
			TaskPriority: 1,
			Retry:        1,
			MaxRetry:     1,
			UserAddress:  "mockUserAddress",
			Logs:         "mockLogs",
		},
		VirtualGroupFamilyId: 1,
		ObjectInfo: &storage_types.ObjectInfo{
			Id:         sdkmath.NewUint(10),
			BucketName: "mockBucketName",
			ObjectName: "mockObjectName",
		},
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `put_object_event_log` (`update_time`,`object_id`,`bucket`,`object`,`state`,`error`,`logs`) VALUES (?,?,?,?,?,?,?)").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.InsertUploadEvent(task)
	assert.Nil(t, err)
}

func TestSpDBImpl_InsertReplicateEventSuccess1(t *testing.T) {
	t.Log("Success case description: task.Error is not nil")
	task := &gfsptask.GfSpReplicatePieceTask{
		Task: &gfsptask.GfSpTask{
			Address:      "mockAddress",
			CreateTime:   time.Now().Unix() - 3,
			UpdateTime:   2,
			Timeout:      0,
			TaskPriority: 1,
			Retry:        1,
			MaxRetry:     1,
			UserAddress:  "mockUserAddress",
			Logs:         "mockLogs",
			Err:          gfsperrors.MakeGfSpError(mockDBInternalError),
		},
		ObjectInfo: &storage_types.ObjectInfo{
			Id:         sdkmath.NewUint(10),
			BucketName: "mockBucketName",
			ObjectName: "mockObjectName",
		},
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `put_object_event_log` (`update_time`,`object_id`,`bucket`,`object`,`state`,`error`,`logs`) VALUES (?,?,?,?,?,?,?)").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.InsertReplicateEvent(task)
	assert.Nil(t, err)
}

func TestSpDBImpl_InsertReplicateEventSuccess2(t *testing.T) {
	t.Log("Success case description: object created time is greater than 10s")
	task := &gfsptask.GfSpReplicatePieceTask{
		Task: &gfsptask.GfSpTask{
			Address:      "mockAddress",
			CreateTime:   time.Now().Unix() - 11,
			UpdateTime:   2,
			Timeout:      0,
			TaskPriority: 1,
			Retry:        1,
			MaxRetry:     1,
			UserAddress:  "mockUserAddress",
			Logs:         "mockLogs",
		},
		ObjectInfo: &storage_types.ObjectInfo{
			Id:         sdkmath.NewUint(10),
			BucketName: "mockBucketName",
			ObjectName: "mockObjectName",
		},
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `put_object_event_log` (`update_time`,`object_id`,`bucket`,`object`,`state`,`error`,`logs`) VALUES (?,?,?,?,?,?,?)").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.InsertReplicateEvent(task)
	assert.Nil(t, err)
}

func TestSpDBImpl_InsertReplicateEventSuccess3(t *testing.T) {
	t.Log("Success case description: object is sealed")
	task := &gfsptask.GfSpReplicatePieceTask{
		Task: &gfsptask.GfSpTask{
			Address:      "mockAddress",
			CreateTime:   time.Now().Unix(),
			UpdateTime:   2,
			Timeout:      0,
			TaskPriority: 1,
			Retry:        1,
			MaxRetry:     1,
			UserAddress:  "mockUserAddress",
			Logs:         "mockLogs",
		},
		ObjectInfo: &storage_types.ObjectInfo{
			Id:         sdkmath.NewUint(10),
			BucketName: "mockBucketName",
			ObjectName: "mockObjectName",
		},
		Sealed: true,
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `put_object_event_log` (`update_time`,`object_id`,`bucket`,`object`,`state`,`error`,`logs`) VALUES (?,?,?,?,?,?,?)").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.InsertReplicateEvent(task)
	assert.Nil(t, err)
}

func TestSpDBImpl_InsertSealEventSuccess1(t *testing.T) {
	t.Log("Success case description: task.Error is not nil")
	task := &gfsptask.GfSpSealObjectTask{
		Task: &gfsptask.GfSpTask{
			Address:      "mockAddress",
			CreateTime:   time.Now().Unix() - 3,
			UpdateTime:   2,
			Timeout:      0,
			TaskPriority: 1,
			Retry:        1,
			MaxRetry:     1,
			UserAddress:  "mockUserAddress",
			Logs:         "mockLogs",
			Err:          gfsperrors.MakeGfSpError(mockDBInternalError),
		},
		ObjectInfo: &storage_types.ObjectInfo{
			Id:         sdkmath.NewUint(10),
			BucketName: "mockBucketName",
			ObjectName: "mockObjectName",
		},
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `put_object_event_log` (`update_time`,`object_id`,`bucket`,`object`,`state`,`error`,`logs`) VALUES (?,?,?,?,?,?,?)").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.InsertSealEvent(task)
	assert.Nil(t, err)
}

func TestSpDBImpl_InsertSealEventSuccess2(t *testing.T) {
	t.Log("Success case description: object created time is greater than 10s")
	task := &gfsptask.GfSpSealObjectTask{
		Task: &gfsptask.GfSpTask{
			Address:      "mockAddress",
			CreateTime:   time.Now().Unix() - 11,
			UpdateTime:   2,
			Timeout:      0,
			TaskPriority: 1,
			Retry:        1,
			MaxRetry:     1,
			UserAddress:  "mockUserAddress",
			Logs:         "mockLogs",
		},
		ObjectInfo: &storage_types.ObjectInfo{
			Id:         sdkmath.NewUint(10),
			BucketName: "mockBucketName",
			ObjectName: "mockObjectName",
		},
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `put_object_event_log` (`update_time`,`object_id`,`bucket`,`object`,`state`,`error`,`logs`) VALUES (?,?,?,?,?,?,?)").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.InsertSealEvent(task)
	assert.Nil(t, err)
}

func TestSpDBImpl_InsertSealEventSuccess3(t *testing.T) {
	t.Log("Success case description: the last case")
	task := &gfsptask.GfSpSealObjectTask{
		Task: &gfsptask.GfSpTask{
			Address:      "mockAddress",
			CreateTime:   time.Now().Unix(),
			UpdateTime:   2,
			Timeout:      0,
			TaskPriority: 1,
			Retry:        1,
			MaxRetry:     1,
			UserAddress:  "mockUserAddress",
			Logs:         "mockLogs",
		},
		ObjectInfo: &storage_types.ObjectInfo{
			Id:         sdkmath.NewUint(10),
			BucketName: "mockBucketName",
			ObjectName: "mockObjectName",
		},
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `put_object_event_log` (`update_time`,`object_id`,`bucket`,`object`,`state`,`error`,`logs`) VALUES (?,?,?,?,?,?,?)").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.InsertSealEvent(task)
	assert.Nil(t, err)
}
