package spdb

import "github.com/bnb-chain/greenfield-storage-provider/core/task"

type GCObjectInfoDB interface {
	SetGCObjectProgress(taskKey string, deletingBlockID uint64, deletingObjectID uint64) error
	DeleteGCObjectProgress(taskKey string) error
	GetAllGCObjectTask(taskKey string) []task.GCObjectTask
}
