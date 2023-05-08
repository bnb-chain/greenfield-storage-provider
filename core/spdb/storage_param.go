package spdb

import storagetypes "github.com/bnb-chain/greenfield/x/storage/types"

// StorageParamDB interface
type StorageParamDB interface {
	// GetStorageParams return storage params
	GetStorageParams() (*storagetypes.Params, error)
	// SetStorageParams set(maybe overwrite) storage params
	SetStorageParams(params *storagetypes.Params) error
}
