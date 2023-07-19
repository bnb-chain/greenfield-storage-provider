package bsdb

import (
	"github.com/forbole/juno/v4/common"
)

// GlobalVirtualGroupFamily defines a set of physical nodes, which only serve part of the buckets
type GlobalVirtualGroupFamily struct {
	ID                         uint64         `gorm:"column:id;primaryKey"`
	GlobalVirtualGroupFamilyId uint32         `gorm:"column:global_virtual_group_family_id;index:idx_vgf_id"`
	PrimarySpId                uint32         `gorm:"column:primary_sp_id;index:idx_primary_sp_id"`
	GlobalVirtualGroupIds      Uint32Array    `gorm:"column:global_virtual_group_ids;type:MEDIUMTEXT"`
	VirtualPaymentAddress      common.Address `gorm:"column:virtual_payment_address;type:BINARY(20)"`

	CreateAt     int64       `gorm:"column:create_at"`
	CreateTxHash common.Hash `gorm:"column:create_tx_hash;type:BINARY(32);not null"`
	CreateTime   int64       `gorm:"column:create_time"` // seconds
	UpdateAt     int64       `gorm:"column:update_at"`
	UpdateTxHash common.Hash `gorm:"column:update_tx_hash;type:BINARY(32);not null"`
	UpdateTime   int64       `gorm:"column:update_time"` // seconds
	Removed      bool        `gorm:"column:removed;default:false"`
}

// TableName is used to set VirtualGroupFamily table name in database
func (g *GlobalVirtualGroupFamily) TableName() string {
	return GlobalVirtualGroupFamilyTableName
}
