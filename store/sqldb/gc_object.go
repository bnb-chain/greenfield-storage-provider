package sqldb

import "github.com/bnb-chain/greenfield-storage-provider/core/task"

func (s *SpDBImpl) SetGCObjectProgress(task string, deletingBlock uint64, deletingObject uint64) error {
	return nil
}
func (s *SpDBImpl) DeleteGCObjectProgress(task string) error           { return nil }
func (s *SpDBImpl) GetAllGCObjectTask(task string) []task.GCObjectTask { return nil }
