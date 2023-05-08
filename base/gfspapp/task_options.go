package gfspapp

import coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"

const (
	NotUseTimeout     int64 = 0
	MinUploadTime     int64 = 2
	MaxUploadTime     int64 = 30
	MinReplicateTime  int64 = 2
	MaxReplicateTime  int64 = 60
	MinReceiveTime    int64 = 2
	MaxReceiveTime    int64 = 5
	MinSealObjectTime int64 = 2
	MaxSealObjectTime int64 = 5
	MinDownloadTime   int64 = 2
	MaxDownloadTime   int64 = 60
	MinGcObjectTime   int64 = 300
	MaxGcObjectTime   int64 = 600
	MinGcZombieTime   int64 = 300
	MaxGcZombieTime   int64 = 600
	MinGCMetaTime     int64 = 300
	MaxGCMetaTime     int64 = 600
)

func (g *GfSpBaseApp) TaskTimeout(task coretask.Task) int64 {
	switch task.Type() {
	case coretask.TypeTaskCreateBucketApproval:
		return NotUseTimeout
	case coretask.TypeTaskCreateObjectApproval:
		return NotUseTimeout
	case coretask.TypeTaskReplicatePieceApproval:
		return NotUseTimeout
	case coretask.TypeTaskUpload:
		uploadTask := task.(coretask.UploadObjectTask)
		timeout := int64(uploadTask.GetObjectInfo().GetPayloadSize()) / g.uploadSpeed
		if timeout < MinUploadTime {
			return MinUploadTime
		}
		if timeout > MaxUploadTime {
			return MaxUploadTime
		}
		return timeout
	case coretask.TypeTaskReplicatePiece:
		replicateTask := task.(coretask.ReplicatePieceTask)
		timeout := int64(replicateTask.GetObjectInfo().GetPayloadSize()) / g.replicateSpeed
		if timeout < MinReplicateTime {
			return MinReplicateTime
		}
		if timeout > MaxReplicateTime {
			return MaxReplicateTime
		}
		return timeout
	case coretask.TypeTaskReceivePiece:
		receiveTask := task.(coretask.ReceivePieceTask)
		timeout := receiveTask.GetPieceSize() / g.replicateSpeed
		if timeout < MinReceiveTime {
			return MinReceiveTime
		}
		if timeout > MaxReceiveTime {
			return MaxReceiveTime
		}
		return timeout
	case coretask.TypeTaskSealObject:
		if g.sealObjectTimeout < MinSealObjectTime {
			return MinSealObjectTime
		}
		if g.sealObjectTimeout > MaxSealObjectTime {
			return MaxSealObjectTime
		}
		return g.sealObjectTimeout
	case coretask.TypeTaskDownloadObject:
		downloadTask := task.(coretask.DownloadObjectTask)
		timeout := downloadTask.GetSize() / g.downloadSpeed
		if timeout < MinDownloadTime {
			return MinDownloadTime
		}
		if timeout > MaxDownloadTime {
			return MinDownloadTime
		}
		return timeout
	case coretask.TypeTaskChallengePiece:
		challengTask := task.(coretask.ChallengePieceTask)
		timeout := challengTask.GetPieceDataSize() / g.downloadSpeed
		if timeout < MinDownloadTime {
			return MinDownloadTime
		}
		if timeout > MaxDownloadTime {
			return MinDownloadTime
		}
		return timeout
	case coretask.TypeTaskGCObject:
		if g.gcObjectTimeout < MinGcObjectTime {
			return MinGcObjectTime
		}
		if g.gcObjectTimeout > MaxGcObjectTime {
			return MaxGcObjectTime
		}
		return g.gcObjectTimeout
	case coretask.TypeTaskGCZombiePiece:
		if g.gcZombieTimeout < MinGcZombieTime {
			return MinGcZombieTime
		}
		if g.gcZombieTimeout > MaxGcZombieTime {
			return MaxGcZombieTime
		}
		return g.gcZombieTimeout
	case coretask.TypeTaskGCMeta:
		if g.gcMetaTimeout < MinGCMetaTime {
			return MinGCMetaTime
		}
		if g.gcMetaTimeout > MaxGCMetaTime {
			return MaxGCMetaTime
		}
		return g.gcMetaTimeout
	}
	return NotUseTimeout
}

const (
	NotUseRetry            int64 = 0
	MinReplicateRetry            = 2
	MaxReplicateRetry            = 6
	MinReceiveConfirmRetry       = 2
	MaxReceiveConfirmRetry       = 6
	MinSealObjectRetry           = 3
	MaxSealObjectRetry           = 10
	MinGCObjectRetry             = 2
	MaxGCObjectRetry             = 5
)

func (g *GfSpBaseApp) TaskMaxRetry(task coretask.Task) int64 {
	switch task.Type() {
	case coretask.TypeTaskCreateBucketApproval:
		return NotUseRetry
	case coretask.TypeTaskCreateObjectApproval:
		return NotUseRetry
	case coretask.TypeTaskReplicatePieceApproval:
		return NotUseRetry
	case coretask.TypeTaskUpload:
		return NotUseRetry
	case coretask.TypeTaskReplicatePiece:
		if g.replicateRetry < MinReplicateRetry {
			return MinReplicateRetry
		}
		if g.replicateRetry > MaxReplicateRetry {
			return MaxReplicateRetry
		}
		return g.replicateRetry
	case coretask.TypeTaskReceivePiece:
		if g.receiveConfirmRetry < MinReceiveConfirmRetry {
			return MinReceiveConfirmRetry
		}
		if g.receiveConfirmRetry > MaxReceiveConfirmRetry {
			return MaxReceiveConfirmRetry
		}
		return g.receiveConfirmRetry
	case coretask.TypeTaskSealObject:
		if g.sealObjectRetry < MinSealObjectRetry {
			return MinSealObjectRetry
		}
		if g.sealObjectRetry > MaxSealObjectRetry {
			return MaxSealObjectRetry
		}
		return g.sealObjectRetry
	case coretask.TypeTaskDownloadObject:
		return NotUseRetry
	case coretask.TypeTaskChallengePiece:
		return NotUseRetry
	case coretask.TypeTaskGCObject:
		if g.gcObjectRetry < MinGCObjectRetry {
			return MinGCObjectRetry
		}
		if g.gcObjectRetry > MaxGCObjectRetry {
			return MaxGCObjectRetry
		}
		return g.gcObjectRetry
	case coretask.TypeTaskGCZombiePiece:
		if g.gcZombieRetry < MinGCObjectRetry {
			return MinGCObjectRetry
		}
		if g.gcZombieRetry > MaxGCObjectRetry {
			return MaxGCObjectRetry
		}
		return g.gcZombieRetry
	case coretask.TypeTaskGCMeta:
		if g.gcMetaRetry < MinGCObjectRetry {
			return MinGCObjectRetry
		}
		if g.gcMetaRetry > MaxGCObjectRetry {
			return MaxGCObjectRetry
		}
		return g.gcMetaRetry
	}
	return 0
}

func (g *GfSpBaseApp) TaskPriority(task coretask.Task) coretask.TPriority {
	switch task.Type() {
	case coretask.TypeTaskCreateBucketApproval:
		return coretask.UnSchedulingPriority
	case coretask.TypeTaskCreateObjectApproval:
		return coretask.UnSchedulingPriority
	case coretask.TypeTaskReplicatePieceApproval:
		return coretask.UnSchedulingPriority
	case coretask.TypeTaskUpload:
		return coretask.UnSchedulingPriority
	case coretask.TypeTaskReplicatePiece:
		return coretask.DefaultLargerTaskPriority
	case coretask.TypeTaskReceivePiece:
		return coretask.MaxTaskPriority
	case coretask.TypeTaskSealObject:
		return coretask.MaxTaskPriority
	case coretask.TypeTaskDownloadObject:
		return coretask.UnSchedulingPriority
	case coretask.TypeTaskChallengePiece:
		return coretask.UnSchedulingPriority
	case coretask.TypeTaskGCObject:
		return coretask.UnSchedulingPriority
	case coretask.TypeTaskGCZombiePiece:
		return coretask.UnSchedulingPriority
	case coretask.TypeTaskGCMeta:
		return coretask.UnSchedulingPriority
	}
	return coretask.UnKnownTaskPriority
}

func (g *GfSpBaseApp) TaskPriorityLevel(task coretask.Task) coretask.TPriorityLevel {
	if task.GetPriority() <= coretask.DefaultSmallerPriority {
		return coretask.TLowPriorityLevel
	}
	if task.GetPriority() > coretask.DefaultLargerTaskPriority {
		return coretask.THighPriorityLevel
	}
	return coretask.TMediumPriorityLevel
}
