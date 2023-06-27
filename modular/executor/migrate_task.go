package executor

import storagetypes "github.com/bnb-chain/greenfield/x/storage/types"

// HandleMigratePieceTask currently get and handle data one by one; in the future, use concurrency
func (e *ExecuteModular) HandleMigratePieceTask(srcSPEndpoint string, objectInfo *storagetypes.ObjectInfo) error {
	// send requests to srcSPEndpoint to get data
	// migrate object to own Uploader or receiver
	return nil
}
