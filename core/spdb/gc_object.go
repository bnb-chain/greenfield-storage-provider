package spdb

import "github.com/bnb-chain/greenfield-storage-provider/core/task"

type GCObjectInfoDB interface {
	SetGCObjectProgress(task string, deletingBlock uint64, deletingObject uint64) error
	DeleteGCObjectProgress(task string) error
	GetAllGCObjectTask(task string) []task.GCObjectTask
}
