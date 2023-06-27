package sqldb

import (
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
)

func (s *SpDBImpl) InsertPutEvent(task coretask.Task) error {
	go func() {
		switch t := task.(type) {
		case *gfsptask.GfSpUploadObjectTask:
			_ = s.InsertUploadEvent(t)
		case *gfsptask.GfSpReplicatePieceTask:
			_ = s.InsertReplicateEvent(t)
		case *gfsptask.GfSpSealObjectTask:
			_ = s.InsertSealEvent(t)
		}
	}()
	return nil
}

func (s *SpDBImpl) InsertUploadEvent(task coretask.UploadObjectTask) error {
	updateTime := time.Now().String()
	s.db.Create(&PutObjectEventTable{
		UpdateTime: updateTime,
		ObjectID:   task.GetObjectInfo().Id.Uint64(),
		Bucket:     task.GetObjectInfo().GetBucketName(),
		Object:     task.GetObjectInfo().GetObjectName(),
		State:      "Upload",
		Error:      task.Error().Error(),
		Logs:       task.GetLogs(),
	})

	if task.Error() != nil {
		s.db.Create(&UploadFailedTable{
			UpdateTime: updateTime,
			ObjectID:   task.GetObjectInfo().Id.Uint64(),
			Bucket:     task.GetObjectInfo().GetBucketName(),
			Object:     task.GetObjectInfo().GetObjectName(),
			Error:      task.Error().Error(),
			Logs:       task.GetLogs(),
		})
	} else if task.GetCreateTime()-time.Now().Unix() > 2 {
		s.db.Create(&UploadTimeoutTable{
			UpdateTime: updateTime,
			ObjectID:   task.GetObjectInfo().Id.Uint64(),
			Bucket:     task.GetObjectInfo().GetBucketName(),
			Object:     task.GetObjectInfo().GetObjectName(),
			Error:      task.Error().Error(),
			Logs:       task.GetLogs(),
		})
	}
	return nil
}

func (s *SpDBImpl) InsertReplicateEvent(task coretask.ReplicatePieceTask) error {
	updateTime := time.Now().String()
	state := "replicate"
	if task.GetSealed() {
		state = "Replicate + Seal"
	}
	s.db.Create(&PutObjectEventTable{
		UpdateTime: updateTime,
		ObjectID:   task.GetObjectInfo().Id.Uint64(),
		Bucket:     task.GetObjectInfo().GetBucketName(),
		Object:     task.GetObjectInfo().GetObjectName(),
		State:      state,
		Error:      task.Error().Error(),
		Logs:       task.GetLogs(),
	})

	if task.Error() != nil {
		s.db.Create(&ReplicateFailedTable{
			UpdateTime: updateTime,
			ObjectID:   task.GetObjectInfo().Id.Uint64(),
			Bucket:     task.GetObjectInfo().GetBucketName(),
			Object:     task.GetObjectInfo().GetObjectName(),
			Error:      task.Error().Error(),
			Logs:       task.GetLogs(),
		})
	} else if task.GetCreateTime()-time.Now().Unix() > 10 {
		s.db.Create(&ReplicateTimeoutTable{
			UpdateTime: updateTime,
			ObjectID:   task.GetObjectInfo().Id.Uint64(),
			Bucket:     task.GetObjectInfo().GetBucketName(),
			Object:     task.GetObjectInfo().GetObjectName(),
			Error:      task.Error().Error(),
			Logs:       task.GetLogs(),
		})
	} else if task.GetSealed() {
		s.db.Create(&PutObjectSuccessTable{
			UpdateTime: updateTime,
			ObjectID:   task.GetObjectInfo().Id.Uint64(),
			Bucket:     task.GetObjectInfo().GetBucketName(),
			Object:     task.GetObjectInfo().GetObjectName(),
			State:      "replicate+seal",
			Error:      task.Error().Error(),
			Logs:       task.GetLogs(),
		})
	}
	return nil
}

func (s *SpDBImpl) InsertSealEvent(task coretask.SealObjectTask) error {
	updateTime := time.Now().String()
	s.db.Create(&PutObjectEventTable{
		UpdateTime: updateTime,
		ObjectID:   task.GetObjectInfo().Id.Uint64(),
		Bucket:     task.GetObjectInfo().GetBucketName(),
		Object:     task.GetObjectInfo().GetObjectName(),
		State:      "Seal",
		Error:      task.Error().Error(),
		Logs:       task.GetLogs(),
	})

	if task.Error() != nil {
		s.db.Create(&SealFailedTable{
			UpdateTime: updateTime,
			ObjectID:   task.GetObjectInfo().Id.Uint64(),
			Bucket:     task.GetObjectInfo().GetBucketName(),
			Object:     task.GetObjectInfo().GetObjectName(),
			Error:      task.Error().Error(),
			Logs:       task.GetLogs(),
		})
	} else if task.GetCreateTime()-time.Now().Unix() > 10 {
		s.db.Create(&SealTimeoutTable{
			UpdateTime: updateTime,
			ObjectID:   task.GetObjectInfo().Id.Uint64(),
			Bucket:     task.GetObjectInfo().GetBucketName(),
			Object:     task.GetObjectInfo().GetObjectName(),
			Error:      task.Error().Error(),
			Logs:       task.GetLogs(),
		})
	} else {
		s.db.Create(&PutObjectSuccessTable{
			UpdateTime: updateTime,
			ObjectID:   task.GetObjectInfo().Id.Uint64(),
			Bucket:     task.GetObjectInfo().GetBucketName(),
			Object:     task.GetObjectInfo().GetObjectName(),
			State:      "seal",
			Error:      task.Error().Error(),
			Logs:       task.GetLogs(),
		})
	}
	return nil
}
