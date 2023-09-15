package gfspapp

import (
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
)

const (
	// MiB defines the MB size
	MiB = 1024 * 1024
	// MinSpeed defines the min speed for data transfer
	MinSpeed = 10 * MiB
	// NotUseTimeout defines the default task timeout.
	NotUseTimeout int64 = 0
	// MinUploadTime defines the min timeout to upload object.
	MinUploadTime int64 = 2
	// MaxUploadTime defines the max timeout to upload object.
	MaxUploadTime int64 = 300
	// MinReplicateTime defines the min timeout to replicate object.
	MinReplicateTime int64 = 90
	// MaxReplicateTime defines the max timeout to replicate object.
	MaxReplicateTime int64 = 500
	// MinReceiveTime defines the min timeout to confirm the received piece whether is sealed on greenfield.
	MinReceiveTime int64 = 90
	// MaxReceiveTime defines the max timeout to confirm the received piece whether is sealed on greenfield.
	MaxReceiveTime int64 = 300
	// MinSealObjectTime defines the min timeout to seal object to greenfield.
	MinSealObjectTime int64 = 90
	// MaxSealObjectTime defines the max timeout to seal object to greenfield.
	MaxSealObjectTime int64 = 300
	// MinDownloadTime defines the min timeout to download object.
	MinDownloadTime int64 = 2
	// MaxDownloadTime defines the max timeout to download object.
	MaxDownloadTime int64 = 300
	// MinGCObjectTime defines the min timeout to gc object.
	MinGCObjectTime int64 = 300
	// MaxGCObjectTime defines the max timeout to gc object.
	MaxGCObjectTime int64 = 600
	// MinGCZombieTime defines the min timeout to gc zombie piece.
	MinGCZombieTime int64 = 300
	// MaxGCZombieTime defines the max timeout to gc zombie piece.
	MaxGCZombieTime int64 = 600
	// MinGCMetaTime defines the min timeout to gc meta.
	MinGCMetaTime int64 = 300
	// MaxGCMetaTime defines the max timeout to gc meta.
	MaxGCMetaTime int64 = 600
	// MinRecoveryTime defines the min timeout to recovery object.
	MinRecoveryTime int64 = 10
	// MaxRecoveryTime defines the max timeout to replicate object.
	MaxRecoveryTime int64 = 50
	// MinMigrateGVGTime defines the min timeout to migrate gvg.
	MinMigrateGVGTime int64 = 1800 // 0.5 hour
	// MaxMigrateGVGTime defines the max timeout to migrate gvg.
	MaxMigrateGVGTime int64 = 3600 // 1 hour
	// MinGCBucketMigrationTime defines the min timeout to gc bucket migration.
	MinGCBucketMigrationTime int64 = 300 // 0.5 hour
	// MaxGCBucketMigrationTime defines the max timeout to gc bucket migration.
	MaxGCBucketMigrationTime int64 = 600 // 1 hour

	// NotUseRetry defines the default task max retry.
	NotUseRetry int64 = 0
	// MinReplicateRetry defines the min retry number to replicate object.
	MinReplicateRetry = 3
	// MaxReplicateRetry defines the max retry number to replicate object.
	MaxReplicateRetry = 6
	// MinReceiveConfirmRetry defines the min retry number to confirm received piece is sealed on greenfield.
	MinReceiveConfirmRetry = 1
	// MaxReceiveConfirmRetry defines the max retry number to confirm received piece is sealed on greenfield.
	MaxReceiveConfirmRetry = 3
	// MinSealObjectRetry defines the min retry number to seal object.
	MinSealObjectRetry = 3
	// MaxSealObjectRetry defines the max retry number to seal object.
	MaxSealObjectRetry = 10
	// MinGCObjectRetry defines the min retry number to gc object.
	MinGCObjectRetry = 3
	// MaxGCObjectRetry defines the min retry number to gc object.
	MaxGCObjectRetry = 5
	// MinRecoveryRetry defines the min retry number to recovery piece.
	MinRecoveryRetry = 2
	// MaxRecoveryRetry  defines the max retry number to recovery piece.
	MaxRecoveryRetry = 3
	// MinMigrateGVGRetry defines the min retry number to migrate gvg.
	MinMigrateGVGRetry = 2
	// MaxMigrateGVGRetry  defines the max retry number to migrate gvg.
	MaxMigrateGVGRetry = 3
	// MinGCBucketMigrationRetry defines the min retry number to gc bucket migration.
	MinGCBucketMigrationRetry = 3
	// MaxBucketMigrationRetry  defines the max retry number to gc bucket migration.
	MaxBucketMigrationRetry = 5
)

// TaskTimeout returns the task timeout by task type and some task need payload size
// to compute, example: upload, download, etc.
func (g *GfSpBaseApp) TaskTimeout(task coretask.Task, size uint64) int64 {
	switch task.Type() {
	case coretask.TypeTaskCreateBucketApproval:
		return NotUseTimeout
	case coretask.TypeTaskCreateObjectApproval:
		return NotUseTimeout
	case coretask.TypeTaskReplicatePieceApproval:
		return NotUseTimeout
	case coretask.TypeTaskUpload:
		timeout := int64(size) / (g.uploadSpeed + 1) / (MinSpeed)
		if timeout < MinUploadTime {
			return MinUploadTime
		}
		if timeout > MaxUploadTime {
			return MaxUploadTime
		}
		return timeout
	case coretask.TypeTaskReplicatePiece:
		timeout := int64(size) / (g.replicateSpeed + 1) / (MinSpeed)
		if timeout < MinReplicateTime {
			return MinReplicateTime
		}
		if timeout > MaxReplicateTime {
			return MaxReplicateTime
		}
		return timeout
	case coretask.TypeTaskReceivePiece:
		timeout := int64(size) / (g.replicateSpeed + 1) / (MinSpeed)
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
		timeout := int64(size) / (g.downloadSpeed + 1) / (MinSpeed)
		if timeout < MinDownloadTime {
			return MinDownloadTime
		}
		if timeout > MaxDownloadTime {
			return MaxDownloadTime
		}
		return timeout
	case coretask.TypeTaskChallengePiece:
		timeout := int64(size) / (g.downloadSpeed + 1) / (MinSpeed)
		if timeout < MinDownloadTime {
			return MinDownloadTime
		}
		if timeout > MaxDownloadTime {
			return MaxDownloadTime
		}
		return timeout
	case coretask.TypeTaskGCObject:
		if g.gcObjectTimeout < MinGCObjectTime {
			return MinGCObjectTime
		}
		if g.gcObjectTimeout > MaxGCObjectTime {
			return MaxGCObjectTime
		}
		return g.gcObjectTimeout
	case coretask.TypeTaskGCZombiePiece:
		if g.gcZombieTimeout < MinGCZombieTime {
			return MinGCZombieTime
		}
		if g.gcZombieTimeout > MaxGCZombieTime {
			return MaxGCZombieTime
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
	case coretask.TypeTaskRecoverPiece:
		timeout := int64(size)/(g.replicateSpeed+1)/(MinSpeed) + 1
		if timeout < MinRecoveryTime {
			return MinRecoveryTime
		}
		if timeout > MaxRecoveryTime {
			return MaxRecoveryTime
		}
		return timeout
	case coretask.TypeTaskMigrateGVG:
		if g.migrateGVGTimeout < MinMigrateGVGTime {
			return MinMigrateGVGTime
		}
		if g.migrateGVGTimeout > MaxMigrateGVGTime {
			return MaxMigrateGVGTime
		}
		return g.migrateGVGTimeout
	case coretask.TypeTaskGCBucketMigration:
		if g.gcBucketMigrationTimeout < MinGCBucketMigrationTime {
			return MinGCBucketMigrationTime
		}
		if g.gcBucketMigrationTimeout > MaxGCBucketMigrationTime {
			return MaxGCBucketMigrationTime
		}
		return g.gcBucketMigrationTimeout
	}
	return NotUseTimeout
}

// TaskMaxRetry returns the task max retry by task type.
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
	case coretask.TypeTaskRecoverPiece:
		if g.recoveryRetry < MinRecoveryRetry {
			return MinRecoveryRetry
		}
		if g.recoveryRetry > MaxRecoveryRetry {
			return MaxRecoveryRetry
		}
		return g.recoveryRetry
	case coretask.TypeTaskMigrateGVG:
		if g.migrateGVGRetry < MinMigrateGVGRetry {
			return MinMigrateGVGRetry
		}
		if g.migrateGVGRetry > MaxMigrateGVGRetry {
			return MaxMigrateGVGRetry
		}
		return g.migrateGVGRetry
	case coretask.TypeTaskGCBucketMigration:
		if g.gcBucketMigrationRetry < MinGCBucketMigrationRetry {
			return MinMigrateGVGRetry
		}
		if g.gcBucketMigrationRetry > MaxBucketMigrationRetry {
			return MaxMigrateGVGRetry
		}
		return g.gcBucketMigrationRetry
	default:
		return NotUseRetry
	}
}

// TaskPriority returns the task priority by task type, it is the default options.
// the task priority support self define and dynamic settings.
func (g *GfSpBaseApp) TaskPriority(task coretask.Task) coretask.TPriority {
	switch task.Type() {
	case coretask.TypeTaskCreateBucketApproval:
		return coretask.UnSchedulingPriority
	case coretask.TypeTaskMigrateBucketApproval:
		return coretask.UnSchedulingPriority
	case coretask.TypeTaskCreateObjectApproval:
		return coretask.UnSchedulingPriority
	case coretask.TypeTaskReplicatePieceApproval:
		return coretask.UnSchedulingPriority
	case coretask.TypeTaskUpload:
		return coretask.UnSchedulingPriority
	case coretask.TypeTaskReplicatePiece:
		return coretask.MaxTaskPriority
	case coretask.TypeTaskReceivePiece:
		return coretask.DefaultSmallerPriority / 4
	case coretask.TypeTaskSealObject:
		return coretask.DefaultSmallerPriority
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
	case coretask.TypeTaskRecoverPiece:
		return coretask.DefaultSmallerPriority / 4
	case coretask.TypeTaskMigrateGVG:
		return coretask.DefaultSmallerPriority
	case coretask.TypeTaskGCBucketMigration:
		return coretask.DefaultSmallerPriority
	default:
		return coretask.UnKnownTaskPriority
	}
}

// TaskPriorityLevel returns the task priority level, it is computed by task priority.
func (g *GfSpBaseApp) TaskPriorityLevel(task coretask.Task) coretask.TPriorityLevel {
	if task.GetPriority() <= coretask.DefaultSmallerPriority {
		return coretask.TLowPriorityLevel
	}
	if task.GetPriority() > coretask.DefaultLargerTaskPriority {
		return coretask.THighPriorityLevel
	}
	return coretask.TMediumPriorityLevel
}
