package store

import types "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"

type MetaDB interface {
	SetIntegrityHash(primary types.StorageProviderInfo, secondary []types.StorageProviderInfo) error
}
