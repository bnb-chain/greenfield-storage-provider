package bsdb

import (
	"github.com/forbole/juno/v4/common"
)

type EventSwapOut struct {
	ID                         uint64             `gorm:"column:id;primaryKey"`
	StorageProviderId          uint32             `gorm:"column:storage_provider_id;index:idx_sp_id"`
	GlobalVirtualGroupFamilyId uint32             `gorm:"column:global_virtual_group_family_id;index:idx_vgf_id"`
	GlobalVirtualGroupIds      common.Uint32Array `gorm:"column:global_virtual_group_ids;type:TEXT"`
	SuccessorSpId              uint32             `gorm:"column:successor_sp_id"`

	CreateAt     int64       `gorm:"column:create_at"`
	CreateTxHash common.Hash `gorm:"column:create_tx_hash;type:BINARY(32);not null"`
	CreateTime   int64       `gorm:"column:create_time"` // seconds
}

func (*EventSwapOut) TableName() string {
	return EventSwapOutTableName
}
