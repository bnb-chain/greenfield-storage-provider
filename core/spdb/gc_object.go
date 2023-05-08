package spdb

import "github.com/bnb-chain/greenfield-storage-provider/core/task"

type GCObjectInfoDB interface {
	SetGCObjectProcess(task string, deletingBlock uint64, deletingObject uint64) error
	DeleteGCObjectProcess(task string) error
	GetAllGCObjectTask(task string) []task.GCObjectTask
}
