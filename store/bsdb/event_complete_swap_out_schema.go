package bsdb

import (
	"github.com/forbole/juno/v4/common"
)

type EventCompleteSwapOut struct {
	ID                         uint64      `gorm:"column:id;primaryKey"`
	StorageProviderId          uint32      `gorm:"column:storage_provider_id;index:idx_sp_id"`
	GlobalVirtualGroupFamilyId uint32      `gorm:"column:global_virtual_group_family_id;index:idx_vgf_id"`
	GlobalVirtualGroupIds      Uint32Array `gorm:"column:global_virtual_group_ids;type:TEXT"`

	CreateAt     int64       `gorm:"column:create_at"`
	CreateTxHash common.Hash `gorm:"column:create_tx_hash;type:BINARY(32);not null"`
	CreateTime   int64       `gorm:"column:create_time"` // seconds
}

func (*EventCompleteSwapOut) TableName() string {
	return EventCompleteSwapOutTableName
}
